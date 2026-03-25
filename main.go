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

		cacheMutex.RLock()
		allData := cache[mode]
		cacheMutex.RUnlock()

		// 1. 필터링 로직
		var filtered []handlers.CrisisData
		for _, d := range allData {
			if fYear != "" && strconv.Itoa(d.Year) != fYear {
				continue
			}
			if fMonth != "" && strconv.Itoa(d.Month) != fMonth {
				continue
			}
			if fDay != "" && strconv.Itoa(d.Day) != fDay {
				continue
			}
			if fHour != "" && strconv.Itoa(d.Hour) != fHour {
				continue
			}
			filtered = append(filtered, d)
		}

		// 2. 통계 계산 (기존 차트용)
		monthly := handlers.GetMonthlyCounts(filtered)
		hourly := handlers.GetHourlyCounts(filtered)

		// 3. 신규 데이터 추출 (연도 목록, 히트맵, KPI)
		availableYears := handlers.GetUniqueYears(allData)
		heatmap := handlers.GetHeatmapData(filtered)
		kpis := handlers.GetDashboardKPIs(filtered, allData)

		// 4. 렌더링
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
			heatmap,
			kpis,
		))
	})

	fmt.Println("서버 실행 중: http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
