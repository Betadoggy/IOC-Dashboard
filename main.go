package main

import (
	"IOC-Dashboard/handlers"
	"fmt"
	"sync"

	"github.com/labstack/echo/v4"
)

// 캐시 및 동기화를 위한 글로벌 변수
var (
	cache      = make(map[string][]handlers.CrisisData)
	cacheMutex sync.RWMutex
)

func main() {
	fmt.Println("데이터 로딩 및 통합 분류 시작...")

	// 1. 단일 통합 파일 로드
	allRawData, err := handlers.LoadExcel("./data/data.xlsx")
	if err != nil {
		fmt.Printf("엑셀 로드 중 오류 발생: %v\n", err)
		return
	}

	// 2. 카테고리별 분리 로직
	var sitData []handlers.CrisisData
	var evtData []handlers.CrisisData

	for _, d := range allRawData {
		if d.Category == "이벤트" {
			evtData = append(evtData, d)
		} else {
			sitData = append(sitData, d)
		}
	}

	// 3. 분류된 데이터를 캐시에 할당
	cacheMutex.Lock()
	cache["situation"] = sitData
	cache["event"] = evtData
	cacheMutex.Unlock()

	fmt.Printf("분류 완료: 상황(%d건), 이벤트(%d건)\n", len(sitData), len(evtData))

	// 4. Echo 서버 설정
	e := echo.New()
	e.Renderer = &TemplRenderer{}

	// 5. 라우터 등록 (routes.go 에 분리된 핸들러 호출)
	registerRoutes(e)

	fmt.Println("서버 실행 중: http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
