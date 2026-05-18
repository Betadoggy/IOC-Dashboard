package handlers

import (
	"fmt"
	"sort"
	"strings"
)

// 구조체 정의
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

type LocationAnalysis struct {
	TopLocations []TypeCount
	TrendData    []TrendPoint
}

// getTopAndTrend: 빈도가 높은 상위 N개 아이템 및 추이를 축출하는 공통 헬퍼
func getTopAndTrend(typeMap map[string]int, trendMap map[string]map[string]int, topN int) ([]TypeCount, []TrendPoint) {
	var topTypes []TypeCount
	for name, count := range typeMap {
		topTypes = append(topTypes, TypeCount{Name: name, Count: count})
	}
	sort.Slice(topTypes, func(i, j int) bool {
		return topTypes[i].Count > topTypes[j].Count
	})
	if len(topTypes) > topN {
		topTypes = topTypes[:topN]
	}

	var trendData []TrendPoint
	topTypeNames := make(map[string]bool, len(topTypes))
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

	return topTypes, trendData
}

// GetTypeAnalysis: 유형별 추이 및 점유율 계산
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

	topTypes, trendData := getTopAndTrend(typeMap, trendMap, 5)

	return TypeAnalysis{TopTypes: topTypes, TrendData: trendData}
}

// GetLocationAnalysis: 위치별 추이 및 점유율 계산
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

	topLocations, trendData := getTopAndTrend(locMap, trendMap, 5)

	return LocationAnalysis{TopLocations: topLocations, TrendData: trendData}
}
