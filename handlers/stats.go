package handlers

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// GetUniqueYears: 데이터에서 유니크한 연도 목록 추출
func GetUniqueYears(data []CrisisData) []string {
	yearMap := make(map[int]bool)
	for _, d := range data {
		if d.Year > 0 {
			yearMap[d.Year] = true
		}
	}
	years := make([]int, 0, len(yearMap))
	for y := range yearMap {
		years = append(years, y)
	}
	sort.Ints(years)
	result := make([]string, len(years))
	for i, y := range years {
		result[i] = strconv.Itoa(y)
	}
	return result
}

// GetAggregateStats: 통합 통계 (월별, 시간별, 월-시간 히트맵, 요일-시간 히트맵, 등급 분포) 수집
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

// GetYearlySeries: 시계열 연도별 집계를 정렬된 라벨/값 형태로 반환
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

	labels := make([]string, len(years))
	counts := make([]int, len(years))
	for i, year := range years {
		labels[i] = strconv.Itoa(year)
		counts[i] = yearMap[year]
	}

	return labels, counts
}

type SankeyData struct {
	Labels  []string
	Sources []int
	Targets []int
	Values  []int
}

func GetSankeyData(data []CrisisData) SankeyData {
	countMap := make(map[string]map[string]int)
	labelIndex := make(map[string]int)
	labels := make([]string, 0)

	addLabel := func(label string) int {
		if idx, ok := labelIndex[label]; ok {
			return idx
		}
		idx := len(labels)
		labels = append(labels, label)
		labelIndex[label] = idx
		return idx
	}

	for _, item := range data {
		source := strings.TrimSpace(strings.Split(item.RawType, ">")[0])
		target := strings.TrimSpace(item.TypeMain)
		if source == "" || target == "" {
			continue
		}
		if countMap[source] == nil {
			countMap[source] = make(map[string]int)
		}
		countMap[source][target]++
	}

	sources := make([]int, 0)
	targets := make([]int, 0)
	values := make([]int, 0)

	for source, targetMap := range countMap {
		srcIdx := addLabel(source)
		for target, value := range targetMap {
			tgtIdx := addLabel(target)
			sources = append(sources, srcIdx)
			targets = append(targets, tgtIdx)
			values = append(values, value)
		}
	}

	return SankeyData{Labels: labels, Sources: sources, Targets: targets, Values: values}
}
