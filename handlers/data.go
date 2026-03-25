package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// 1. 구조체 정의 (다른 파일에서도 쓰이도록 대문자로 시작)
type CrisisData struct {
	Timestamp string
	Year      int
	Month     int
	Day       int
	Hour      int
	Type      string
	TypeMain  string
	Location  string
}

func LoadExcel(path string) ([]CrisisData, error) {
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
			if i == 0 || len(row) < 11 {
				continue
			}

			rawTime := row[2]
			t, err := parseFlexTime(rawTime)
			if err != nil {
				continue
			}

			// 계층 분리
			typeParts := strings.Split(row[8], ">")
			// locParts := strings.Split(row[10], ">") // 에러 원인: 선언 후 미사용 -> 삭제하거나 아래처럼 활용

			typeMain := ""
			if len(typeParts) > 0 {
				typeMain = strings.TrimSpace(typeParts[0])
			}

			allData = append(allData, CrisisData{
				Timestamp: rawTime,
				Year:      t.Year(),
				Month:     int(t.Month()),
				Day:       t.Day(),
				Hour:      t.Hour(),
				Type:      row[8],
				TypeMain:  typeMain,
				Location:  row[10], // row[10] 자체를 사용
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
