package main

import (
	"encoding/json"
	"io"
	"kara"
	"model"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/av/pubsub"
	"github.com/nareix/joy4/format"
	"github.com/nareix/joy4/format/flv"
	"github.com/nareix/joy4/format/rtmp"
)

type avinfo struct {
	AID   string   `json:"aid"`
	Tags  []string `json:"tags"`
	Title string   `json:"title"`
}

//Channel is struct of the subpub queue
type Channel struct {
	que *pubsub.Queue
}

type writeFlusher struct {
	httpflusher http.Flusher
	io.Writer
}

var (
	cachepath = "./zealot/.avcache"
	videopath = "./static/video/"
)

func init() {
	format.RegisterAll()
}

func (w writeFlusher) Flush() error {
	w.httpflusher.Flush()
	return nil
}

func main() {
	//WS server
	h := kara.NewHub()
	go h.Run()

	//RTMP server
	s := new(rtmp.Server)
	mutex := new(sync.RWMutex)
	channels := make(map[string]*Channel)
	//hanleplay
	s.HandlePlay = func(conn *rtmp.Conn) {
		mutex.RLock()
		ch := channels[conn.URL.Path]
		mutex.RUnlock()

		if ch != nil {
			cursor := ch.que.Latest()
			avutil.CopyFile(conn, cursor)
		}
	}
	//HandlePublish
	s.HandlePublish = func(conn *rtmp.Conn) {
		streams, _ := conn.Streams()
		mutex.Lock()
		ch := channels[conn.URL.Path]
		if ch != nil {
			ch = nil
		} else {
			ch = new(Channel)
			ch.que = pubsub.NewQueue()
			ch.que.WriteHeader(streams)
			channels[conn.URL.Path] = ch
		}
		mutex.Unlock()
		if ch == nil {
			return
		}
		defer ch.que.Close()
		avutil.CopyPackets(ch.que, conn)
		mutex.Lock()
		delete(channels, conn.URL.Path)
		mutex.Unlock()
		return
	}

	go s.ListenAndServe()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mutex.RLock()
		ch := channels[r.URL.Path]
		mutex.RUnlock()

		if ch != nil {
			w.Header().Set("Content-Type", "video/x-flv")
			w.Header().Set("Transfer-Encoding", "chunked")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(200)
			flusher := w.(http.Flusher)
			flusher.Flush()

			muxer := flv.NewMuxerWriteFlusher(writeFlusher{httpflusher: flusher, Writer: w})
			cursor := ch.que.Latest()

			avutil.CopyFile(muxer, cursor)
		} else {
			http.NotFound(w, r)
		}
	})

	go http.ListenAndServe(":8091", nil)

	//Web Server
	e := echo.New()

	e.Static("/", "static")
	e.Static("/watch/js", "static/js")
	e.Static("/watch/css", "static/css")
	e.Static("/watch/video", "static/video")

	e.File("/", "static/chat.html")

	e.File("/live", "static/live.html")
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

	// e.GET("/v", func(c echo.Context) error {
	// 	mutex.RLock()
	// 	ch := channels[c.Request().URL.Path]
	// 	mutex.RUnlock()
	// 	if ch != nil {
	// 		c.Response().Header().Set("Content-Type", "video/x-flv")
	// 		c.Response().Header().Set("Transfer-Encoding", "chunked")
	// 		c.Response().WriteHeader(http.StatusOK)
	// 		flusher := c.Response().Writer().(http.Flusher)
	// 		c.Response().Flush()
	// 		muxer := flv.NewMuxerWriteFlusher(writeFlusher{httpflusher: flusher, Writer: c.Response().Writer()})
	// 		cursor := ch.que.Latest()
	//
	// 		avutil.CopyFile(muxer, cursor)
	// 		return c.JSON(http.StatusOK, nil)
	// 	}
	// 	return c.JSON(http.StatusNotFound, nil)
	//
	// })

	e.Logger.Fatal(e.Start(":8090"))
}
