package main

import (
	"diplom/pkg/auth"
	"diplom/pkg/handlers"
	middleware2 "diplom/pkg/middleware"
	"diplom/pkg/render"
	"diplom/pkg/reports"
	"flag"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var root = flag.String("root", "/daria", "root directory of a project")

func main() {
	flag.Parse()

	e := echo.New()
	logger := e.Logger.(*log.Logger)

	db, err := gorm.Open(sqlite.Open(*root+"/db/gorm.db"), &gorm.Config{})
	must(logger, err)

	e.Use(middleware.Logger())
	e.Renderer = render.New(*root + "/frontend/templates")

	authRepo := auth.New(db, logger)
	reportRepo := reports.New(*root + "/reports")
	handler := handlers.New(authRepo, logger, reportRepo)

	e.GET("/", handler.Index)
	e.GET("/signup", handler.SignUpPage)
	e.POST("/signup", handler.SignUp)
	e.GET("/login", handler.LogInPage)
	e.POST("/login", handler.LogIn)
	e.GET("/signout", handler.SignOut)
	e.GET("/report", handler.Report)
	e.POST("/report", handler.AddReport)
	e.GET("/delete_report", handler.DeleteReport)

	e.Static("/assets", *root+"/frontend/assets")
	e.Static("/source_reports", *root+"/reports")
	e.Use(middleware2.Auth(authRepo, logger))
	e.Logger.Fatal(e.Start(":1323"))
}

func must(logger echo.Logger, err error) {
	if err != nil {
		logger.Fatal(err)
	}
}
