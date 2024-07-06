package main

import (
	"go-db/config"
	_handler "go-db/go-db/delivery/http"
	_usecase "go-db/go-db/usecase"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	e *echo.Echo
)

func init() {
	e = echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	config.InitializeConfig()
}

func main() {

	godbUsecase := _usecase.NewGoDBUsecase()
	_handler.NewGoDBHandler(e, godbUsecase)

	e.Logger.Fatal(e.Start(":" + config.PORT))
}
