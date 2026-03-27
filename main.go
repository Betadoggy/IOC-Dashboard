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

func applyFilters(data []handlers.CrisisData, year, month, day, hour int) []handlers.CrisisData {
	if year == -1 && month == -1 && day == -1 && hour == -1 {
		return data
	}
	filtered := make([]handlers.CrisisData, 0, len(data))
	for _, d := range data {
		if year != -1 && d.Year != year {
			continue
		}
		if month != -1 && d.Month != month {
			continue
		}
		if day != -1 && d.Day != day {
			continue
		}
		if hour != -1 && d.Hour != hour {
			continue
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

		fYear := c.QueryParam("year")
		fMonth := c.QueryParam("month")
		fDay := c.QueryParam("day")
		fHour := c.QueryParam("hour")
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
		year := parseQueryInt(fYear)
		month := parseQueryInt(fMonth)
		day := parseQueryInt(fDay)
		hour := parseQueryInt(fHour)
		filtered := applyFilters(allData, year, month, day, hour)

		// 2. 통계 계산 (기존 차트용)
		monthly, hourly, heatmap, weekdayHeatmap, severityCounts := handlers.GetAggregateStats(filtered)

		// 3. 신규 데이터 추출 (연도 목록, KPI)
		availableYears := handlers.GetUniqueYears(allData)
		kpis := handlers.GetDashboardKPIs(filtered, allData)

		// 4. 유형 및 위치 분석
		groupCol := "월"
		if fMonth != "" {
			groupCol = "일"
		}
		typeAnalysis := handlers.GetTypeAnalysis(filtered, typeLevel, groupCol)
		locAnalysis := handlers.GetLocationAnalysis(filtered, locLevel, groupCol)

		// 5. 렌더링
		return c.Render(http.StatusOK, "", views.Dashboard(
			len(filtered),
			filtered,
			monthly,
			hourly,
			mode,
			fYear,
			fMonth,
			fDay,
			fHour,
			availableYears,
			kpis,
			heatmap,
			weekdayHeatmap,
			severityCounts,
			typeLevel,
			locLevel,
			typeAnalysis,
			locAnalysis,
		))
	})

	e.GET("/api/analysis", func(c echo.Context) error {
		mode := c.QueryParam("mode")
		if mode == "" {
			mode = "situation"
		}

		fYear := c.QueryParam("year")
		fMonth := c.QueryParam("month")
		fDay := c.QueryParam("day")
		fHour := c.QueryParam("hour")
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
		year := parseQueryInt(fYear)
		month := parseQueryInt(fMonth)
		day := parseQueryInt(fDay)
		hour := parseQueryInt(fHour)
		filtered := applyFilters(allData, year, month, day, hour)

		groupCol := "월"
		if month != 0 {
			groupCol = "일"
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
