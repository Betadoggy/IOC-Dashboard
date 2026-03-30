# 인천공항 안전 관리 모니터링 체계 대시보드


## Getting Started
Open http://localhost:8080 with your browser to see the result.

### (참고용) 다운로드 링크

[VS Code](https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user)

[Go Lang](https://go.dev/dl/go1.26.1.windows-amd64.msi)

[Python 3.12.10](https://www.python.org/ftp/python/3.12.10/python-3.12.10-amd64.exe)

[Git 2.53.0.2](https://github.com/git-for-windows/git/releases/download/v2.53.0.windows.2/Git-2.53.0.2-64-bit.exe)

### 명령어

Golang PATH 지정
```
$env:Path += ";C:\Users\90915\go\bin"
```

templ설치
```
go install github.com/a-h/templ/cmd/templ@latest
```

air 설치
```
go install github.com/air-verse/air@latest
```

Golang 빌드(현재 폴더 이름으로 실행 파일 생성)
```
go build .
```
