package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/scbizu/letschat/chat"
)

func main() {
	h := chat.NewHub()
	go h.Run()
	e := echo.New()

	e.Static("/", "static")

	e.File("/", "static/chat.html")

	e.GET("/video", func(c echo.Context) error {
		vid := c.QueryParam("vid")
		if vid == "" {
			return c.JSON(404, "no such source.")
		}
		video, err := os.Open("./static/video/" + vid + ".flv")
		if err != nil {
			return c.JSON(404, "no such source.")
		}
		return c.Stream(200, "video/mp4", video)
	})

	e.GET("/ws", func(c echo.Context) error {
		chat.ServeWs(h, c.Response().Writer(), c.Request())
		return nil
	})

	e.Logger.Fatal(e.Start(":8090"))
}
