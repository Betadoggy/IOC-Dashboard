package handlers

import (
	"sort"
	"strings"
	"time"
)

type WeekdayTypeHeatmap struct {
	TypeNames []string
	Data      [][]int // [weekday][type]
}

type WeekdayLocationHeatmap struct {
	LocationNames []string
	Data          [][]int // [weekday][location]
}

// GetWeekdayTypeHeatmap: 요일별 - 유형별 분포 히트맵 계산
func GetWeekdayTypeHeatmap(filtered []CrisisData, topN int) WeekdayTypeHeatmap {
	typeMap := make(map[string]int)
	for _, d := range filtered {
		var t string
		parts := strings.Split(d.Type, ">")
		if len(parts) > 0 {
			t = strings.TrimSpace(parts[0])
		}
		if t == "" {
			continue
		}
		typeMap[t]++
	}

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

	typeIndex := make(map[string]int)
	typeNames := make([]string, len(topTypes))
	for i, tt := range topTypes {
		typeIndex[tt.Name] = i
		typeNames[i] = tt.Name
	}

	heatmapData := make([][]int, 7)
	for i := range heatmapData {
		heatmapData[i] = make([]int, len(topTypes))
	}

	for _, d := range filtered {
		if d.Year <= 0 || d.Month < 1 || d.Month > 12 || d.Day < 1 || d.Day > 31 {
			continue
		}

		wd := int(time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC).Weekday())

		var t string
		parts := strings.Split(d.Type, ">")
		if len(parts) > 0 {
			t = strings.TrimSpace(parts[0])
		}
		if t == "" {
			continue
		}

		if idx, ok := typeIndex[t]; ok {
			heatmapData[wd][idx]++
		}
	}

	return WeekdayTypeHeatmap{
		TypeNames: typeNames,
		Data:      heatmapData,
	}
}

// GetWeekdayLocationHeatmap: 요일별 - 위치별 분포 히트맵 계산
func GetWeekdayLocationHeatmap(filtered []CrisisData, topN int) WeekdayLocationHeatmap {
	locationMap := make(map[string]int)
	for _, d := range filtered {
		var l string
		parts := strings.Split(d.Location, ">")
		if len(parts) > 0 {
			l = strings.TrimSpace(parts[0])
		}
		if l == "" {
			continue
		}
		locationMap[l]++
	}

	var topLocations []TypeCount
	for name, count := range locationMap {
		topLocations = append(topLocations, TypeCount{Name: name, Count: count})
	}
	sort.Slice(topLocations, func(i, j int) bool {
		return topLocations[i].Count > topLocations[j].Count
	})
	if len(topLocations) > topN {
		topLocations = topLocations[:topN]
	}

	locationIndex := make(map[string]int)
	locationNames := make([]string, len(topLocations))
	for i, tl := range topLocations {
		locationIndex[tl.Name] = i
		locationNames[i] = tl.Name
	}

	heatmapData := make([][]int, 7)
	for i := range heatmapData {
		heatmapData[i] = make([]int, len(topLocations))
	}

	for _, d := range filtered {
		if d.Year <= 0 || d.Month < 1 || d.Month > 12 || d.Day < 1 || d.Day > 31 {
			continue
		}

		wd := int(time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC).Weekday())

		var l string
		parts := strings.Split(d.Location, ">")
		if len(parts) > 0 {
			l = strings.TrimSpace(parts[0])
		}
		if l == "" {
			continue
		}

		if idx, ok := locationIndex[l]; ok {
			heatmapData[wd][idx]++
		}
	}

	return WeekdayLocationHeatmap{
		LocationNames: locationNames,
		Data:          heatmapData,
	}
}
