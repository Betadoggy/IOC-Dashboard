package handlers

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 연도 목록 추출
func GetUniqueYears(data []CrisisData) []string {
	yearMap := make(map[int]bool)
	for _, d := range data {
		if d.Year > 0 {
			yearMap[d.Year] = true
		}
	}
	var years []int
	for y := range yearMap {
		years = append(years, y)
	}
	sort.Ints(years)
	var result []string
	for _, y := range years {
		result = append(result, strconv.Itoa(y))
	}
	return result
}

// 통합 통계: 월별, 시간별, 히트맵, 등급 분포를 한 번에 수집합니다.
func GetAggregateStats(data []CrisisData) ([12]int, [24]int, [12][24]int, [7][24]int, map[string]int) {
	var monthly [12]int
	var hourly [24]int
	var heatmap [12][24]int
	var weekdayHeatmap [7][24]int
	severityCounts := make(map[string]int)
	for _, item := range data {
		if item.Month >= 1 && item.Month <= 12 {
			monthly[item.Month-1]++
		}
		if item.Hour >= 0 && item.Hour <= 23 {
			hourly[item.Hour]++
		}
		if item.Month >= 1 && item.Month <= 12 && item.Hour >= 0 && item.Hour <= 23 {
			heatmap[item.Month-1][item.Hour]++
		}
		if item.Year > 0 && item.Month >= 1 && item.Month <= 12 && item.Day >= 1 && item.Hour >= 0 && item.Hour <= 23 {
			wd := int(time.Date(item.Year, time.Month(item.Month), item.Day, 0, 0, 0, 0, time.UTC).Weekday())
			weekdayHeatmap[wd][item.Hour]++
		}
		if item.Severity != "" {
			severityCounts[item.Severity]++
		}
	}
	return monthly, hourly, heatmap, weekdayHeatmap, severityCounts
}

// 시계열 연도별 집계를 정렬된 라벨/값 형태로 반환합니다.
func GetYearlySeries(data []CrisisData) ([]string, []int) {
	yearMap := make(map[int]int)
	for _, item := range data {
		if item.Year > 0 {
			yearMap[item.Year]++
		}
	}

	if len(yearMap) == 0 {
		return []string{}, []int{}
	}

	years := make([]int, 0, len(yearMap))
	for year := range yearMap {
		years = append(years, year)
	}
	sort.Ints(years)

	labels := make([]string, 0, len(years))
	counts := make([]int, 0, len(years))
	for _, year := range years {
		labels = append(labels, strconv.Itoa(year))
		counts = append(counts, yearMap[year])
	}

	return labels, counts
}

// KPI 계산 (구조체 및 함수)
type DashboardKPIs struct {
	TotalCount   int
	DailyAverage float64
	PeakHour     int
	TopType      string
	MTTRByType   map[string]float64
	LTIDays      float64
}

func GetDashboardKPIs(filtered []CrisisData, allData []CrisisData) DashboardKPIs {
	kpi := DashboardKPIs{
		TotalCount:   len(filtered),
		DailyAverage: 0.0,
		PeakHour:     -1,
		TopType:      "N/A",
		MTTRByType:   map[string]float64{},
		LTIDays:      -1,
	}
	if len(filtered) == 0 {
		return kpi
	}

	hCounts := make(map[int]int)
	maxH, maxHC := 0, 0
	tCounts := make(map[string]int)
	maxT, maxTC := "", 0
	mttrSumByType := make(map[string]float64)
	mttrCountByType := make(map[string]int)

	for _, d := range filtered {
		hCounts[d.Hour]++
		if hCounts[d.Hour] > maxHC {
			maxHC = hCounts[d.Hour]
			maxH = d.Hour
		}
		if d.TypeMain != "" {
			tCounts[d.TypeMain]++
			if tCounts[d.TypeMain] > maxTC {
				maxTC = tCounts[d.TypeMain]
				maxT = d.TypeMain
			}
		}

		occurredAt, errOccurred := parseFlexTime(d.Timestamp)
		resolvedRaw := d.ResolvedAt
		if errOccurred == nil && resolvedRaw != "" {
			resolvedAt, errResolved := parseFlexTime(resolvedRaw)
			if errResolved == nil && !resolvedAt.Before(occurredAt) {
				hours := resolvedAt.Sub(occurredAt).Hours()
				typeKey := d.TypeMain
				if typeKey == "" {
					typeKey = "기타"
				}
				mttrSumByType[typeKey] += hours
				mttrCountByType[typeKey]++
			}
		}
	}
	kpi.PeakHour = maxH
	kpi.TopType = maxT

	totalDays := calcTotalDays(filtered)
	if totalDays > 0 {
		kpi.DailyAverage = float64(len(filtered)) / float64(totalDays)
	}

	for typeKey, sum := range mttrSumByType {
		count := mttrCountByType[typeKey]
		if count > 0 {
			kpi.MTTRByType[typeKey] = sum / float64(count)
		}
	}

	now := time.Now()
	var lastSevere *time.Time
	for _, d := range allData {
		if d.Severity != "심각" {
			continue
		}
		occurredAt, err := parseFlexTime(d.Timestamp)
		if err != nil {
			continue
		}
		if lastSevere == nil || occurredAt.After(*lastSevere) {
			t := occurredAt
			lastSevere = &t
		}
	}
	if lastSevere != nil && now.After(*lastSevere) {
		kpi.LTIDays = now.Sub(*lastSevere).Hours() / 24.0
	}

	return kpi
}

// 데이터의 최소~최대 날짜 범위를 일수로 반환
func calcTotalDays(filtered []CrisisData) int {
	if len(filtered) == 0 {
		return 0
	}
	var minDate, maxDate time.Time
	first := true
	for _, d := range filtered {
		if d.Year <= 0 || d.Month < 1 || d.Day < 1 {
			continue
		}
		dt := time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
		if first {
			minDate = dt
			maxDate = dt
			first = false
		} else {
			if dt.Before(minDate) {
				minDate = dt
			}
			if dt.After(maxDate) {
				maxDate = dt
			}
		}
	}
	if first {
		return 0
	}
	return int(maxDate.Sub(minDate).Hours()/24) + 1
}

// 유형 분석 구조체
type TypeCount struct {
	Name  string
	Count int
}

type TrendPoint struct {
	Period string
	Name   string
	Count  int
}

type TypeAnalysis struct {
	TopTypes  []TypeCount
	TrendData []TrendPoint
}

// 유형별 분석 함수
func GetTypeAnalysis(filtered []CrisisData, typeLevel string, groupCol string) TypeAnalysis {
	var typeCol string
	switch typeLevel {
	case "대분류":
		typeCol = "main"
	case "중분류":
		typeCol = "mid"
	case "소분류":
		typeCol = "sub"
	default:
		typeCol = "main"
	}

	typeMap := make(map[string]int)
	trendMap := make(map[string]map[string]int)

	for _, d := range filtered {
		var t string
		parts := strings.Split(d.Type, ">")
		switch typeCol {
		case "main":
			if len(parts) > 0 {
				t = strings.TrimSpace(parts[0])
			}
		case "mid":
			if len(parts) > 1 {
				t = strings.TrimSpace(parts[1])
			}
		case "sub":
			if len(parts) > 2 {
				t = strings.TrimSpace(parts[2])
			}
		}
		if t == "" {
			continue
		}
		typeMap[t]++

		var period string
		switch groupCol {
		case "연도":
			period = fmt.Sprintf("%d", d.Year)
		default:
			period = fmt.Sprintf("%d-%02d", d.Year, d.Month)
		}
		if trendMap[period] == nil {
			trendMap[period] = make(map[string]int)
		}
		trendMap[period][t]++
	}

	// Top 5 types
	var topTypes []TypeCount
	for name, count := range typeMap {
		topTypes = append(topTypes, TypeCount{Name: name, Count: count})
	}
	sort.Slice(topTypes, func(i, j int) bool {
		return topTypes[i].Count > topTypes[j].Count
	})
	if len(topTypes) > 5 {
		topTypes = topTypes[:5]
	}

	// Trend data for top types
	var trendData []TrendPoint
	topTypeNames := make(map[string]bool)
	for _, tt := range topTypes {
		topTypeNames[tt.Name] = true
	}
	for period, types := range trendMap {
		for name, count := range types {
			if topTypeNames[name] {
				trendData = append(trendData, TrendPoint{Period: period, Name: name, Count: count})
			}
		}
	}

	return TypeAnalysis{TopTypes: topTypes, TrendData: trendData}
}

// 위치 분석 구조체 (유형과 동일)
type LocationAnalysis struct {
	TopLocations []TypeCount
	TrendData    []TrendPoint
}

// 위치별 분석 함수
func GetLocationAnalysis(filtered []CrisisData, locLevel string, groupCol string) LocationAnalysis {
	var locCol string
	switch locLevel {
	case "대분류":
		locCol = "main"
	case "중분류":
		locCol = "mid"
	default:
		locCol = "main"
	}

	locMap := make(map[string]int)
	trendMap := make(map[string]map[string]int)

	for _, d := range filtered {
		var l string
		parts := strings.Split(d.Location, ">")
		switch locCol {
		case "main":
			if len(parts) > 0 {
				l = strings.TrimSpace(parts[0])
			}
		case "mid":
			if len(parts) > 1 {
				l = strings.TrimSpace(parts[1])
			}
		}
		if l == "" {
			continue
		}
		locMap[l]++

		var period string
		switch groupCol {
		case "연도":
			period = fmt.Sprintf("%d", d.Year)
		default:
			period = fmt.Sprintf("%d-%02d", d.Year, d.Month)
		}
		if trendMap[period] == nil {
			trendMap[period] = make(map[string]int)
		}
		trendMap[period][l]++
	}

	// Top 5 locations
	var topLocations []TypeCount
	for name, count := range locMap {
		topLocations = append(topLocations, TypeCount{Name: name, Count: count})
	}
	sort.Slice(topLocations, func(i, j int) bool {
		return topLocations[i].Count > topLocations[j].Count
	})
	if len(topLocations) > 5 {
		topLocations = topLocations[:5]
	}

	// Trend data for top locations
	var trendData []TrendPoint
	topLocNames := make(map[string]bool)
	for _, tl := range topLocations {
		topLocNames[tl.Name] = true
	}
	for period, locs := range trendMap {
		for name, count := range locs {
			if topLocNames[name] {
				trendData = append(trendData, TrendPoint{Period: period, Name: name, Count: count})
			}
		}
	}

	return LocationAnalysis{TopLocations: topLocations, TrendData: trendData}
}
