package handlers

// CrisisData: 상황/이벤트 등의 사고 원본 구조체
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

// CategoryMap: 코드 번호를 텍스트 명칭으로 변환하기 위한 맵
type CategoryMap struct {
	Main   map[string]string // A열(코드) -> B열(명칭)
	Medium map[string]string // C열(코드) -> D열(명칭)
	Small  map[string]string // E열(코드) -> F열(명칭)
}
