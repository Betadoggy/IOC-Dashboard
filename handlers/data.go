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

// CategoryMap은 숫자 코드를 텍스트로 변환하기 위한 맵입니다.
type CategoryMap struct {
	Main   map[string]string // A열(코드) -> B열(명칭)
	Medium map[string]string // C열(코드) -> D열(명칭)
	Small  map[string]string // E열(코드) -> F열(명칭)
}

// LoadCategoryMap은 assets/category.xlsx 파일을 읽어 매핑 정보를 반환합니다.
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

	rows, err := f.GetRows(f.GetSheetList()[0]) // 첫 번째 시트 사용
	if err != nil {
		return nil, err
	}

	for i, row := range rows {
		if i == 0 || len(row) < 2 { // 헤더 제외 및 최소 열 체크
			continue
		}
		// 대분류 (A, B)
		if len(row) > 1 {
			cm.Main[strings.TrimSpace(row[0])] = strings.TrimSpace(row[1])
		}
		// 중분류 (C, D)
		if len(row) > 3 {
			cm.Medium[strings.TrimSpace(row[2])] = strings.TrimSpace(row[3])
		}
		// 소분류 (E, F)
		if len(row) > 5 {
			cm.Small[strings.TrimSpace(row[4])] = strings.TrimSpace(row[5])
		}
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

			// 코드값 추출
			codeLarge := strings.TrimSpace(row[8])
			codeMedium := ""
			if len(row) > 9 {
				codeMedium = strings.TrimSpace(row[9])
			}
			codeSmall := ""
			if len(row) > 10 {
				codeSmall = strings.TrimSpace(row[10])
			}

			// 텍스트 매핑 (맵에 없으면 코드 그대로 노출)
			tLarge := cm.Main[codeLarge]
			if tLarge == "" {
				tLarge = codeLarge
			}

			tMedium := cm.Medium[codeMedium]
			if tMedium == "" {
				tMedium = codeMedium
			}

			tSmall := cm.Small[codeSmall]
			if tSmall == "" {
				tSmall = codeSmall
			}

			// fullType 조립
			fullType := tLarge
			if tMedium != "" {
				fullType += ">" + tMedium
			}
			if tSmall != "" {
				fullType += ">" + tSmall
			}

			typeMain := tLarge

			rawResolved := ""
			if len(row) > 2 {
				rawResolved = strings.TrimSpace(row[2])
			}

			severity := ""
			if len(row) > 5 {
				severity = strings.TrimSpace(row[5])
			}

			category := "상황"
			if len(row) > 15 {
				category = strings.TrimSpace(row[15])
			}

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
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"01-02-06 15:04:05",
		"1/2/06 15:04",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unknown time format")
}
