package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/xuri/excelize/v2"
)

// CategoryMapper는 대/중/소분류 매핑 데이터를 담는 구조체입니다.
type CategoryMapper struct {
	Large  map[string]string // Key: "1"
	Medium map[string]string // Key: "1-8"
	Small  map[string]string // Key: "1-8-3"
}

// loadCategoryMapping은 category.xlsx 파일을 읽어 매핑 구조체를 생성합니다.
func loadCategoryMapping(filePath string) (*CategoryMapper, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("카테고리 파일 열기 실패: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("카테고리 시트 읽기 실패: %w", err)
	}

	mapper := &CategoryMapper{
		Large:  make(map[string]string),
		Medium: make(map[string]string),
		Small:  make(map[string]string),
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) > 1 && strings.TrimSpace(row[0]) != "" {
			mapper.Large[strings.TrimSpace(row[0])] = strings.TrimSpace(row[1])
		}
		if len(row) > 3 && strings.TrimSpace(row[2]) != "" {
			mapper.Medium[strings.TrimSpace(row[2])] = strings.TrimSpace(row[3])
		}
		if len(row) > 5 && strings.TrimSpace(row[4]) != "" {
			mapper.Small[strings.TrimSpace(row[4])] = strings.TrimSpace(row[5])
		}
	}

	return mapper, nil
}

// processDataAndMap은 data.xlsx의 I, J, K열을 읽어 매핑 결과를 출력합니다.
func processDataAndMap(dataPath string, mapper *CategoryMapper) error {
	f, err := excelize.OpenFile(dataPath)
	if err != nil {
		return fmt.Errorf("데이터 파일 열기 실패: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("데이터 시트 읽기 실패: %w", err)
	}

	fmt.Println("\n========================= I, J, K열 실제 데이터 매핑 결과 =========================")

	// 엑셀 열 인덱스 정의 (A=0, B=1, ..., I=8, J=9, K=10)
	const (
		idxLarge  = 8  // I열
		idxMedium = 9  // J열
		idxSmall  = 10 // K열
	)

	for i, row := range rows {
		if i == 0 {
			// 헤더 행에서 I, J, K열의 제목이 무엇인지 확인용 출력
			if len(row) > idxSmall {
				fmt.Printf("[Row %d] 타겟 헤더 정보 -> I열(%s), J열(%s), K열(%s)\n",
					i+1, row[idxLarge], row[idxMedium], row[idxSmall])
			} else {
				fmt.Printf("[Row %d] 원본 헤더 전체: %v (K열까지 데이터가 존재하지 않습니다)\n", i+1, row)
			}
			fmt.Println(strings.Repeat("-", 90))
			continue
		}

		// 최소한 K열(인덱스 10)까지는 데이터 배열이 채워져 있어야 파싱 가능
		if len(row) <= idxSmall {
			continue
		}

		// I, J, K열 데이터 추출 및 공백 제거
		rawLarge := strings.TrimSpace(row[idxLarge])
		rawMedium := strings.TrimSpace(row[idxMedium])
		rawSmall := strings.TrimSpace(row[idxSmall])

		// 대/중/소분류 중 하나라도 비어있으면 매핑에서 제외
		if rawLarge == "" || rawMedium == "" || rawSmall == "" {
			continue
		}

		// 복합 키 생성 규칙 적용
		mediumKey := fmt.Sprintf("%s-%s", rawLarge, rawMedium)
		smallKey := fmt.Sprintf("%s-%s-%s", rawLarge, rawMedium, rawSmall)

		// 카테고리 텍스트 매핑 (없으면 에러 추적을 위해 코드 그대로 노출하거나 알 수 없음 처리)
		largeName := mapper.Large[rawLarge]
		if largeName == "" {
			largeName = "알 수 없음(" + rawLarge + ")"
		}

		mediumName := mapper.Medium[mediumKey]
		if mediumName == "" {
			mediumName = "알 수 없음(" + mediumKey + ")"
		}

		smallName := mapper.Small[smallKey]
		if smallName == "" {
			smallName = "알 수 없음(" + smallKey + ")"
		}

		// 상위 15개 행만 샘플 출력
		if i <= 15 {
			fmt.Printf("[Row %02d] 원본 코드(I,J,K): (%s, %s, %s) -> 변환 명칭: [%s] > [%s] > [%s]\n",
				i+1, rawLarge, rawMedium, rawSmall, largeName, mediumName, smallName)
		} else if i == 16 {
			fmt.Println("... 이하 샘플 생략 ...")
		}
	}

	return nil
}

func main() {
	// 1. 카테고리 사전 데이터 로드
	mapper, err := loadCategoryMapping("data/category.xlsx")
	if err != nil {
		log.Fatalf("카테고리 로드 실패: %v", err)
	}
	fmt.Println("✓ 카테고리 매핑 테이블 사전 로드 완료")

	// 2. 실제 메인 이벤트 데이터 매핑 테스트 실행
	err = processDataAndMap("data/data.xlsx", mapper)
	if err != nil {
		log.Fatalf("데이터 매핑 실패: %v", err)
	}
}
