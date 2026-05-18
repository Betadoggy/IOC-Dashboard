package main

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// TemplRenderer: templ 컴포넌트를 echo에서 렌더링하기 위한 구조체
type TemplRenderer struct{}

func (t *TemplRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if component, ok := data.(templ.Component); ok {
		return component.Render(context.Background(), w)
	}
	return fmt.Errorf("not a templ component")
}
