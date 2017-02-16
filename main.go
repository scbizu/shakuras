package main

import (
	"github.com/labstack/echo"
	"github.com/scbizu/letschat/chat"
)

func main() {
	h := chat.NewHub()
	go h.Run()
	e := echo.New()

	e.Static("/", "static")

	e.File("/", "static/chat.html")

	e.GET("/ws", func(c echo.Context) error {
		chat.ServeWs(h, c.Response().Writer(), c.Request())
		return nil
	})

	e.Logger.Fatal(e.Start(":8090"))
}
