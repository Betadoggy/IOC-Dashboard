package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// 1. 구조체 정의
type CrisisData struct {
	Timestamp  string
	ResolvedAt string
	Severity   string
	Year       int
	Month      int
	Day        int
	Hour       int
	Type       string
	TypeMain   string
	Location   string
	Category   string
}

// CategoryMap은 코드 번호를 텍스트로 변환하기 위한 맵입니다.
type CategoryMap struct {
	Main   map[string]string // A열(코드) -> B열(명칭)
	Medium map[string]string // C열(코드) -> D열(명칭)
	Small  map[string]string // E열(코드) -> F열(명칭)
}

func LoadCategoryMap(path string) (*CategoryMap, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cm := &CategoryMap{
		Main:   make(map[string]string),
		Medium: make(map[string]string),
		Small:  make(map[string]string),
	}

	// 첫 번째 시트 이름 가져오기
	sheetName := f.GetSheetList()[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	for i, row := range rows {
		if i == 0 { continue } // 헤더 스킵

		// 헬퍼 함수: 범위 체크 후 맵에 데이터 삽입
		assign := func(m map[string]string, kIdx, vIdx int) {
			if len(row) > vIdx {
				k, v := strings.TrimSpace(row[kIdx]), strings.TrimSpace(row[vIdx])
				if k != "" { m[k] = v }
			}
		}

		assign(cm.Main, 0, 1)   // 대분류 A, B열
		assign(cm.Medium, 2, 3) // 중분류 C, D열
		assign(cm.Small, 4, 5)  // 소분류 E, F열
	}

	return cm, nil
}

func LoadExcel(path string) ([]CrisisData, error) {
	// 1. 카테고리 매핑 정보 먼저 로드
	cm, err := LoadCategoryMap("assets/category.xlsx")
	if err != nil {
		// 카테고리 로드 실패 시 에러를 반환하거나 빈 맵으로 진행 (여기서는 에러 반환)
		return nil, fmt.Errorf("category load fail: %v", err)
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var allData []CrisisData
	sheets := f.GetSheetList()

	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			fmt.Printf("[%s] 시트 읽기 실패: %v\n", sheetName, err)
			continue
		}

		if len(rows) <= 1 {
			continue
		}

		for i, row := range rows {
			if i == 0 || len(row) < 16 {
				continue
			}

			rawTime := strings.TrimSpace(row[1])
			t, err := parseFlexTime(rawTime)
			if err != nil {
				continue
			}

			// 1. 값 추출 및 코드 조합 (한 줄씩 정리)
			cL := strings.TrimSpace(row[8])
			cM := fmt.Sprintf("%s-%s", cL, strings.TrimSpace(row[9])) // 1-1 형태
			cS := fmt.Sprintf("%s-%s", cM, strings.TrimSpace(row[10])) // 1-1-1 형태

			// 2. 매핑 함수 (값이 없으면 원본 코드 반환)
			getMap := func(m map[string]string, key, raw string) string {
				if val, ok := m[key]; ok && val != "" { return val }
				return raw
			}

			tL, tM, tS := getMap(cm.Main, cL, cL), getMap(cm.Medium, cM, strings.TrimSpace(row[9])), getMap(cm.Small, cS, strings.TrimSpace(row[10]))

			// 3. fullType 조립 (Slice와 Join 활용)
			parts := []string{tL}
			if tM != "" { parts = append(parts, tM) }
			if tS != "" { parts = append(parts, tS) }

			fullType := strings.Join(parts, ">")
			typeMain := tL

			rawResolved := strings.TrimSpace(row[2])
			severity := strings.TrimSpace(row[5])
			category := strings.TrimSpace(row[15])

			allData = append(allData, CrisisData{
				Timestamp:  rawTime,
				ResolvedAt: rawResolved,
				Severity:   severity,
				Year:       t.Year(),
				Month:      int(t.Month()),
				Day:        t.Day(),
				Hour:       t.Hour(),
				Type:       fullType,
				TypeMain:   typeMain,
				Location:   row[13],
				Category:   category,
			})
		}
	}
	return allData, nil
}

func parseFlexTime(raw string) (time.Time, error) {
    // 특정 레이아웃 파싱, 엑셀에서는 2006-01-02 15:04:05로 보이더라도 실제 처리는 1/2/06 15:04
    layout := "1/2/06 15:04"
    
    t, err := time.Parse(layout, raw)
    if err != nil {
        return time.Time{}, fmt.Errorf("invalid time format (expected 1/2/06 15:04): %w", err)
    }
    
    return t, nil
}