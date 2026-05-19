package main

import (
	"IOC-Dashboard/handlers"
	"IOC-Dashboard/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

// registerRoutes: 라우팅 설정
func registerRoutes(e *echo.Echo) {
	e.GET("/", handleDashboard)
	e.GET("/api/analysis", handleAnalysisAPI)
}

// handleDashboard: 메인 대시보드 렌더링 핸들러
func handleDashboard(c echo.Context) error {
	mode := c.QueryParam("mode")
	if mode == "" {
		mode = "situation"
	}

	// 쿼리 파라미터 수집 (기본값 설정 포함)
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

	// 해당 모드의 원본 데이터 가져오기
	cacheMutex.RLock()
	allData := cache[mode]
	cacheMutex.RUnlock()

	// 필터 적용
	startYear := parseQueryInt(fStartYear)
	startMonth := parseQueryInt(fStartMonth)
	endYear := parseQueryInt(fEndYear)
	endMonth := parseQueryInt(fEndMonth)
	filtered := applyFilters(allData, startYear, startMonth, endYear, endMonth)

	// 통계 데이터 계산
	monthly, hourly, heatmap, weekdayHeatmap, severityCounts := handlers.GetAggregateStats(filtered)
	yearlyLabels, yearlyCounts := handlers.GetYearlySeries(filtered)
	availableYears := handlers.GetUniqueYears(allData)
	kpis := handlers.GetDashboardKPIs(filtered, allData, typeLevel)
	weekdayTypeHeatmap := handlers.GetWeekdayTypeHeatmap(filtered, 7)
	weekdayLocationHeatmap := handlers.GetWeekdayLocationHeatmap(filtered, 7)

	groupCol := "월"
	if groupBy == "year" {
		groupCol = "연도"
	}

	typeAnalysis := handlers.GetTypeAnalysis(filtered, typeLevel, groupCol)
	locAnalysis := handlers.GetLocationAnalysis(filtered, locLevel, groupCol)

	// 뷰 렌더링
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
		weekdayTypeHeatmap,
		weekdayLocationHeatmap,
		handlers.GetSankeyData(filtered),
	))
}

// handleAnalysisAPI: 분석용 JSON API 핸들러
func handleAnalysisAPI(c echo.Context) error {
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

	filtered := applyFilters(allData, parseQueryInt(fStartYear), parseQueryInt(fStartMonth), parseQueryInt(fEndYear), parseQueryInt(fEndMonth))

	groupCol := "월"
	if groupBy == "year" {
		groupCol = "연도"
	}

	typeAnalysis := handlers.GetTypeAnalysis(filtered, typeLevel, groupCol)
	locAnalysis := handlers.GetLocationAnalysis(filtered, locLevel, groupCol)
	kpis := handlers.GetDashboardKPIs(filtered, allData, typeLevel)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"typeAnalysis": typeAnalysis,
		"locAnalysis":  locAnalysis,
		"mttrByType":   kpis.MTTRByType,
	})
}
