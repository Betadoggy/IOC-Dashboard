package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// 1. 구조체 정의 (다른 파일에서도 쓰이도록 대문자로 시작)
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
	Category   string // 추가: 엑셀 P열의 '이벤트' 또는 '상황' 구분값 저장
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
			// P열(index 15)까지 안전하게 읽기 위해 길이를 16 이상으로 체크
			if i == 0 || len(row) < 16 {
				continue
			}

			rawTime := strings.TrimSpace(row[1])
			t, err := parseFlexTime(rawTime)
			if err != nil {
				continue
			}

			tLarge := ""
			if len(row) > 8 {
				tLarge = strings.TrimSpace(row[8])
			}

			tMedium := ""
			if len(row) > 9 {
				tMedium = strings.TrimSpace(row[9])
			}

			tSmall := ""
			if len(row) > 10 {
				tSmall = strings.TrimSpace(row[10])
			}

			// I, J, K를 다시 ">"로 합쳐서 기존 row[7]처럼 만들기
			fullType := tLarge
			if tMedium != "" {
				fullType += ">" + tMedium
			}
			if tSmall != "" {
				fullType += ">" + tSmall
			}

			typeMain := tLarge // 기존 로직의 typeParts[0]과 동일한 역할

			rawResolved := ""
			if len(row) > 2 {
				rawResolved = strings.TrimSpace(row[2])
			}

			severity := ""
			if len(row) > 5 {
				severity = strings.TrimSpace(row[5])
			}

			// P열(구분) 데이터 추출 로직 추가
			category := "상황"   // 기본값
			if len(row) > 15 { // P열이 존재할 경우
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
				Category:   category, // 매핑
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
