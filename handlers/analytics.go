package handlers

import (
	"sort"
	"strconv"
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

// KPI 계산 (구조체 및 함수)
type DashboardKPIs struct {
	TotalCount int
	PeakHour   int
	TopType    string
}

func GetDashboardKPIs(filtered []CrisisData, allData []CrisisData) DashboardKPIs {
	kpi := DashboardKPIs{TotalCount: len(filtered), PeakHour: -1, TopType: "N/A"}
	if len(filtered) == 0 {
		return kpi
	}

	hCounts := make(map[int]int)
	maxH, maxHC := 0, 0
	tCounts := make(map[string]int)
	maxT, maxTC := "", 0

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
	}
	kpi.PeakHour = maxH
	kpi.TopType = maxT
	return kpi
}
