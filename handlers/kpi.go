package handlers

import (
	"time"
)

// DashboardKPIs: 대시보드 지표 구조체
type DashboardKPIs struct {
	TotalCount   int
	DailyAverage float64
	PeakHour     int
	TopType      string
	MTTRByType   map[string]float64
	LTIDays      float64
}

// GetDashboardKPIs: 필터링 및 전체 데이터 분석을 기반으로 주요 KPI를 구합니다.
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

	hCounts := make(map[int]int, 24)
	maxH, maxHC := 0, 0
	tCounts := make(map[string]int)
	maxT, maxTC := "", 0
	mttrSumByType := make(map[string]float64)
	mttrCountByType := make(map[string]int)

	for _, d := range filtered {
		h := d.Hour
		hCounts[h]++
		if hCounts[h] > maxHC {
			maxHC = hCounts[h]
			maxH = h
		}
		if d.TypeMain != "" {
			t := d.TypeMain
			tCounts[t]++
			if tCounts[t] > maxTC {
				maxTC = tCounts[t]
				maxT = t
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
		if d.Severity == "심각" {
			occurredAt, err := parseFlexTime(d.Timestamp)
			if err == nil {
				if lastSevere == nil || occurredAt.After(*lastSevere) {
					t := occurredAt
					lastSevere = &t
				}
			}
		}
	}
	if lastSevere != nil && now.After(*lastSevere) {
		kpi.LTIDays = now.Sub(*lastSevere).Hours() / 24.0
	}

	return kpi
}

// calcTotalDays: 데이터의 최소~최대 날짜 범위를 일수로 반환
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
