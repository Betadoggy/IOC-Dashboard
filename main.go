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

// 캐시 및 동기화를 위한 변수
var (
	cache      = make(map[string][]handlers.CrisisData)
	cacheMutex sync.RWMutex
)

// TemplRenderer: templ 컴포넌트를 echo에서 렌더링하기 위한 구조체
type TemplRenderer struct{}

func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if component, ok := data.(templ.Component); ok {
		return component.Render(context.Background(), w)
	}
	return fmt.Errorf("not a templ component")
}

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

func main() {
	fmt.Println("데이터 로딩 및 통합 분류 시작...")

	// 1. 단일 통합 파일 로드 (상황/이벤트가 한 파일에 있음)
	// handlers.LoadExcel 내부에서 P열을 읽어 Category 필드에 넣어준다고 가정합니다.
	allRawData, err := handlers.LoadExcel("./data/data.xlsx")
	if err != nil {
		fmt.Printf("엑셀 로드 중 오류 발생: %v\n", err)
		return
	}

	// 2. 카테고리별 분리 로직
	var sitData []handlers.CrisisData
	var evtData []handlers.CrisisData

	for _, d := range allRawData {
		// P열 데이터가 담긴 Category 필드로 구분
		if d.Category == "이벤트" {
			evtData = append(evtData, d)
		} else {
			sitData = append(sitData, d)
		}
	}

	// 3. 분류된 데이터를 캐시에 할당 (기존 키값 유지)
	cacheMutex.Lock()
	cache["situation"] = sitData
	cache["event"] = evtData
	cacheMutex.Unlock()

	fmt.Printf("분류 완료: 상황(%d건), 이벤트(%d건)\n", len(sitData), len(evtData))

	// 4. Echo 서버 설정
	e := echo.New()
	e.Renderer = &TemplRenderer{}

	// 메인 대시보드 핸들러
	e.GET("/", func(c echo.Context) error {
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
		switch activeTab {
		case "timeseries", "type", "location", "weekday", "hourly":
		default:
			activeTab = "timeseries"
		}

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
		kpis := handlers.GetDashboardKPIs(filtered, allData)

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
		))
	})

	// 분석용 API 핸들러
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

		filtered := applyFilters(allData, parseQueryInt(fStartYear), parseQueryInt(fStartMonth), parseQueryInt(fEndYear), parseQueryInt(fEndMonth))

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
