package main

import (
	"IOC-Dashboard/handlers"
	"IOC-Dashboard/views"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

var (
	cache      = make(map[string][]handlers.CrisisData)
	cacheMutex sync.RWMutex
)

type TemplRenderer struct{}

func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if component, ok := data.(templ.Component); ok {
		return component.Render(context.Background(), w)
	}
	return fmt.Errorf("not a templ component")
}

func parseQueryInt(q string) int {
	if q == "" {
		return -1
	}
	v, _ := strconv.Atoi(q)
	return v
}

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

func main() {
	fmt.Println("데이터 로딩 및 캐싱 시작...")
	sitData, _ := handlers.LoadExcel("./data/상황.xlsx")
	evtData, _ := handlers.LoadExcel("./data/이벤트.xlsx")

	cacheMutex.Lock()
	cache["situation"] = sitData
	cache["event"] = evtData
	cacheMutex.Unlock()

	e := echo.New()
	e.Renderer = &TemplRenderer{}

	e.GET("/", func(c echo.Context) error {
		mode := c.QueryParam("mode")
		if mode == "" {
			mode = "situation"
		}

		fStartYear := c.QueryParam("start_year")
		if fStartYear == "" {
			fStartYear = "2023"
		}
		fStartMonth := c.QueryParam("start_month")
		if fStartMonth == "" {
			fStartMonth = "1"
		}
		fEndYear := c.QueryParam("end_year")
		if fEndYear == "" {
			fEndYear = "2025"
		}
		fEndMonth := c.QueryParam("end_month")
		if fEndMonth == "" {
			fEndMonth = "12"
		}
		groupBy := c.QueryParam("group_by")
		if groupBy == "" || groupBy == "all" {
			groupBy = "month"
		}
		typeLevel := c.QueryParam("type_level")
		if typeLevel == "" {
			typeLevel = "대분류"
		}
		locLevel := c.QueryParam("loc_level")
		if locLevel == "" {
			locLevel = "대분류"
		}
		activeTab := c.QueryParam("tab")
		switch activeTab {
		case "timeseries", "type", "location", "weekday", "hourly":
		default:
			activeTab = "timeseries"
		}

		cacheMutex.RLock()
		allData := cache[mode]
		cacheMutex.RUnlock()

		// 쿼리값 파싱 + 필터링
		startYear := parseQueryInt(fStartYear)
		startMonth := parseQueryInt(fStartMonth)
		endYear := parseQueryInt(fEndYear)
		endMonth := parseQueryInt(fEndMonth)
		filtered := applyFilters(allData, startYear, startMonth, endYear, endMonth)

		// 2. 통계 계산 (기존 차트용)
		monthly, hourly, heatmap, weekdayHeatmap, severityCounts := handlers.GetAggregateStats(filtered)
		yearlyLabels, yearlyCounts := handlers.GetYearlySeries(filtered)

		// 3. 신규 데이터 추출 (연도 목록, KPI)
		availableYears := handlers.GetUniqueYears(allData)
		kpis := handlers.GetDashboardKPIs(filtered, allData)

		// 4. 유형 및 위치 분석
		groupCol := "월"
		if groupBy == "year" {
			groupCol = "연도"
		}
		typeAnalysis := handlers.GetTypeAnalysis(filtered, typeLevel, groupCol)
		locAnalysis := handlers.GetLocationAnalysis(filtered, locLevel, groupCol)

		// 5. 렌더링
		return c.Render(http.StatusOK, "", views.Dashboard(
			len(filtered),
			monthly,
			yearlyLabels,
			yearlyCounts,
			hourly,
			mode,
			fStartYear,
			fStartMonth,
			fEndYear,
			fEndMonth,
			groupBy,
			availableYears,
			kpis,
			heatmap,
			weekdayHeatmap,
			severityCounts,
			typeLevel,
			locLevel,
			activeTab,
			typeAnalysis,
			locAnalysis,
		))
	})

	e.GET("/api/analysis", func(c echo.Context) error {
		mode := c.QueryParam("mode")
		if mode == "" {
			mode = "situation"
		}

		fStartYear := c.QueryParam("start_year")
		if fStartYear == "" {
			fStartYear = "2023"
		}
		fStartMonth := c.QueryParam("start_month")
		if fStartMonth == "" {
			fStartMonth = "1"
		}
		fEndYear := c.QueryParam("end_year")
		if fEndYear == "" {
			fEndYear = "2025"
		}
		fEndMonth := c.QueryParam("end_month")
		if fEndMonth == "" {
			fEndMonth = "12"
		}
		groupBy := c.QueryParam("group_by")
		if groupBy == "" || groupBy == "all" {
			groupBy = "month"
		}
		typeLevel := c.QueryParam("type_level")
		if typeLevel == "" {
			typeLevel = "대분류"
		}
		locLevel := c.QueryParam("loc_level")
		if locLevel == "" {
			locLevel = "대분류"
		}

		cacheMutex.RLock()
		allData := cache[mode]
		cacheMutex.RUnlock()

		// 쿼리값 파싱 + 필터링
		startYear := parseQueryInt(fStartYear)
		startMonth := parseQueryInt(fStartMonth)
		endYear := parseQueryInt(fEndYear)
		endMonth := parseQueryInt(fEndMonth)
		filtered := applyFilters(allData, startYear, startMonth, endYear, endMonth)

		groupCol := "월"
		if groupBy == "year" {
			groupCol = "연도"
		}
		typeAnalysis := handlers.GetTypeAnalysis(filtered, typeLevel, groupCol)
		locAnalysis := handlers.GetLocationAnalysis(filtered, locLevel, groupCol)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"typeAnalysis": typeAnalysis,
			"locAnalysis":  locAnalysis,
		})
	})

	fmt.Println("서버 실행 중: http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
