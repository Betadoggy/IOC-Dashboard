package handlers

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// 월별 집계
func GetMonthlyCounts(data []CrisisData) [12]int {
	var counts [12]int
	for _, item := range data {
		if item.Month >= 1 && item.Month <= 12 {
			counts[item.Month-1]++
		}
	}
	return counts
}

// 시간대별 집계
func GetHourlyCounts(data []CrisisData) [24]int {
	var counts [24]int
	for _, item := range data {
		if item.Hour >= 0 && item.Hour <= 23 {
			counts[item.Hour]++
		}
	}
	return counts
}

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

// 히트맵 데이터 추출
func GetHeatmapData(data []CrisisData) [12][24]int {
	var heatmap [12][24]int
	for _, item := range data {
		if item.Month >= 1 && item.Month <= 12 && item.Hour >= 0 && item.Hour <= 23 {
			heatmap[item.Month-1][item.Hour]++
		}
	}
	return heatmap
}

// 일별 통합 통계: 월별, 시간별, 히트맵을 한 번에 수집해 루프를 줄입니다.
func GetAggregateStats(data []CrisisData) ([12]int, [24]int, [12][24]int) {
	var monthly [12]int
	var hourly [24]int
	var heatmap [12][24]int
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
	}
	return monthly, hourly, heatmap
}

// KPI 계산 (구조체 및 함수)
type DashboardKPIs struct {
	TotalCount   int
	DailyAverage float64
	PeakHour     int
	TopType      string
}

func GetDashboardKPIs(filtered []CrisisData, allData []CrisisData) DashboardKPIs {
	kpi := DashboardKPIs{TotalCount: len(filtered), DailyAverage: 0.0, PeakHour: -1, TopType: "N/A"}
	if len(filtered) == 0 {
		return kpi
	}

	hCounts := make(map[int]int)
	maxH, maxHC := 0, 0
	tCounts := make(map[string]int)
	maxT, maxTC := "", 0
	uniqueDays := make(map[string]bool)

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
		dateKey := fmt.Sprintf("%d-%02d-%02d", d.Year, d.Month, d.Day)
		uniqueDays[dateKey] = true
	}
	kpi.PeakHour = maxH
	kpi.TopType = maxT
	if len(uniqueDays) > 0 {
		kpi.DailyAverage = float64(len(filtered)) / float64(len(uniqueDays))
	}
	return kpi
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
		if groupCol == "일" {
			period = fmt.Sprintf("%d-%02d-%02d", d.Year, d.Month, d.Day)
		} else {
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
		if groupCol == "일" {
			period = fmt.Sprintf("%d-%02d-%02d", d.Year, d.Month, d.Day)
		} else {
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
