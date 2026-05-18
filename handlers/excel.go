package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// LoadCategoryMap: 카테고리 매핑 엑셀 정보를 로드합니다.
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

	sheetName := f.GetSheetList()[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	for i, row := range rows {
		if i == 0 {
			continue // 헤더 스킵
		}

		assign := func(m map[string]string, kIdx, vIdx int) {
			if len(row) > vIdx {
				k, v := strings.TrimSpace(row[kIdx]), strings.TrimSpace(row[vIdx])
				if k != "" {
					m[k] = v
				}
			}
		}

		assign(cm.Main, 0, 1)   // 대분류 A, B열
		assign(cm.Medium, 2, 3) // 중분류 C, D열
		assign(cm.Small, 4, 5)  // 소분류 E, F열
	}

	return cm, nil
}

// LoadExcel: 메인 상황/이벤트 데이터를 로드하고 정제합니다.
func LoadExcel(path string) ([]CrisisData, error) {
	cm, err := LoadCategoryMap("assets/category.xlsx")
	if err != nil {
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

			cL := strings.TrimSpace(row[8])
			cM := fmt.Sprintf("%s-%s", cL, strings.TrimSpace(row[9]))
			cS := fmt.Sprintf("%s-%s", cM, strings.TrimSpace(row[10]))

			getMap := func(m map[string]string, key, raw string) string {
				if val, ok := m[key]; ok && val != "" {
					return val
				}
				return raw
			}

			tL, tM, tS := getMap(cm.Main, cL, cL), getMap(cm.Medium, cM, strings.TrimSpace(row[9])), getMap(cm.Small, cS, strings.TrimSpace(row[10]))

			parts := []string{tL}
			if tM != "" {
				parts = append(parts, tM)
			}
			if tS != "" {
				parts = append(parts, tS)
			}

			fullType := strings.Join(parts, ">")
			typeMain := tL
			rawType := strings.TrimSpace(row[7])

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
				RawType:    rawType,
				Location:   row[13],
				Category:   category,
			})
		}
	}
	return allData, nil
}

// parseFlexTime: 문자열 형식의 시간을 time.Time 객체로 변환합니다.
func parseFlexTime(raw string) (time.Time, error) {
	layout := "1/2/06 15:04"
	t, err := time.Parse(layout, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %w", err)
	}
	return t, nil
}
