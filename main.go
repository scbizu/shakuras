package main

import (
	"encoding/json"
	"kara"
	"model"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

type avinfo struct {
	AID   string   `json:"aid"`
	Tags  []string `json:"tags"`
	Title string   `json:"title"`
}

var (
	cachepath = "./zealot/.avcache"
	videopath = "./static/video/"
)

func main() {
	h := kara.NewHub()
	go h.Run()
	e := echo.New()

	e.Static("/", "static")
	e.Static("/watch/js", "static/js")
	e.Static("/watch/css", "static/css")
	e.Static("/watch/video", "static/video")

	e.File("/", "static/chat.html")

	e.File("/watch/*", "static/chat.html")

	e.GET("/video", func(c echo.Context) error {
		vid := c.QueryParam("vid")
		if vid == "" {
			return c.JSON(404, "no such source.")
		}
		video, err := os.Open(videopath + vid + ".flv")
		if err != nil {
			return c.JSON(404, "no such source.")
		}
		return c.Stream(200, "video/mp4", video)
	})

	e.GET("/ws", func(c echo.Context) error {
		kara.ServeWs(h, c.Response().Writer(), c.Request())
		return nil
	})

	e.GET("/videotags", func(c echo.Context) error {
		data, err := model.ChangeType(cachepath, model.Bucketname)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		tags := []string{}
		info := make(map[string][]string)
		err = json.Unmarshal([]byte(data), &info)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		for k := range info {
			tags = append(tags, k)
		}
		return c.JSON(http.StatusOK, tags)
	})

	e.GET("/getSeries", func(c echo.Context) error {
		tagName := c.QueryParam("tagname")

		data, err := model.ChangeType(cachepath, model.Bucketname)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		info := make(map[string][]string)
		err = json.Unmarshal([]byte(data), &info)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		if info[tagName] != nil {
			res := []map[string]interface{}{}
			for _, v := range info[tagName] {
				resinfo := make(map[string]interface{})
				json.Unmarshal([]byte(v), &resinfo)
				res = append(res, resinfo)
			}
			return c.JSON(http.StatusOK, res)
		}
		return c.JSON(http.StatusOK, "no this tag")
	})

	e.GET("/firstvideo", func(c echo.Context) error {
		data, err := model.GetFirstVID(cachepath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		var firstvid string
		for k := range data {
			firstvid = k
		}
		video, err := os.Open(videopath + firstvid + ".flv")
		if err != nil {
			return c.JSON(http.StatusNotFound, "no such source.")
		}
		return c.Stream(http.StatusOK, "video/mp4", video)
	})

	// e.GET("/watch", func(c echo.Context)error{
	// 	tagname:=c.Param("v")
	//
	// })

	e.Logger.Fatal(e.Start(":8090"))
}
