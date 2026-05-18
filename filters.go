package main

import (
	"IOC-Dashboard/handlers"
	"strconv"
)

// parseQueryInt: 쿼리 스트링의 숫자를 파싱 (기본값 -1)
func parseQueryInt(q string) int {
	if q == "" {
		return -1
	}
	v, _ := strconv.Atoi(q)
	return v
}

// applyFilters: 연도/월 필터링 로직
func applyFilters(data []handlers.CrisisData, startYear, startMonth, endYear, endMonth int) []handlers.CrisisData {
	if startYear == -1 && startMonth == -1 && endYear == -1 && endMonth == -1 {
		return data
	}
	filtered := make([]handlers.CrisisData, 0, len(data))
	for _, d := range data {
		ym := d.Year*100 + d.Month
		if startYear != -1 {
			startM := 1
			if startMonth != -1 {
				startM = startMonth
			}
			if ym < startYear*100+startM {
				continue
			}
		}
		if endYear != -1 {
			endM := 12
			if endMonth != -1 {
				endM = endMonth
			}
			if ym > endYear*100+endM {
				continue
			}
		}
		filtered = append(filtered, d)
	}
	return filtered
}
