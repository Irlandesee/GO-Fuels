package main

import (
	"net/http"
	_ "net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/labstack/echo/v4"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/gorm"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":9999"))
}
