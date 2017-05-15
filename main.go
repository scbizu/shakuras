package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"kara"
	"model"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/av/pktque"
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

//FrameDropper defines the drop frame
type FrameDropper struct {
	Interval     int
	n            int
	skipping     bool
	DelaySkip    time.Duration
	lasttime     time.Time
	lastpkttime  time.Duration
	delay        time.Duration
	SkipInterval int
}

//User user
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

//ModifyPacket implement the  ModifyPacket interface
func (fp *FrameDropper) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if fp.DelaySkip != 0 && pkt.Idx == int8(videoidx) {
		now := time.Now()
		if !fp.lasttime.IsZero() {
			realdiff := now.Sub(fp.lasttime)
			pktdiff := pkt.Time - fp.lastpkttime
			fp.delay += realdiff - pktdiff
		}
		fp.lasttime = time.Now()
		fp.lastpkttime = pkt.Time

		if !fp.skipping {
			if fp.delay > fp.DelaySkip {
				fp.skipping = true
				fp.delay = 0
			}
		} else {
			if pkt.IsKeyFrame {
				fp.skipping = false
			}
		}
		if fp.skipping {
			drop = true
		}

		if fp.SkipInterval != 0 && pkt.IsKeyFrame {
			if fp.n == fp.SkipInterval {
				fp.n = 0
				fp.skipping = true
			}
			fp.n++
		}
	}

	if fp.Interval != 0 {
		if fp.n >= fp.Interval && pkt.Idx == int8(videoidx) && !pkt.IsKeyFrame {
			drop = true
			fp.n = 0
		}
		fp.n++
	}

	return
}

//forwardviaFFmpeg will forward the rtmp address to another address
//need ffmpeg plugin
func forwardviaFFmpeg(src string, dst string) {
	exec.Command("ffmpeg", "-re", "-i", src, "-acodec", "libfaac", "-ab", "128k", "-vcodec", "libx264", "-s", "640x360", "-b:v", "500k", "-preset", "medium", "-vprofile", "baseline", "-r", "25 ", "-f", "flv", dst)
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
		//fork
		forwardviaFFmpeg("rtmp://localhost/test", "rtmp://localhost/md")
		//fork
		forwardviaFFmpeg("rtmp://localhost/test", "rtmp://localhost/md2")
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
			streams, err := cursor.Streams()
			if err != nil {
				panic(err)
			}

			query := r.URL.Query()
			if q := query.Get("delaygop"); q != "" {
				n := 0
				fmt.Sscanf(q, "%d", &n)
				cursor = ch.que.DelayedGopCount(n)
			} else if q := query.Get("delaytime"); q != "" {
				dur, _ := time.ParseDuration(q)
				cursor = ch.que.DelayedTime(dur)
			}

			filters := pktque.Filters{}

			if q := query.Get("waitkey"); q != "" {
				filters = append(filters, &pktque.WaitKeyFrame{})
			}

			filters = append(filters, &pktque.FixTime{StartFromZero: true, MakeIncrement: true})

			if q := query.Get("framedrop"); q != "" {
				n := 0
				fmt.Sscanf(q, "%d", &n)
				filters = append(filters, &FrameDropper{Interval: n})
			}

			if q := query.Get("delayskip"); q != "" {
				dur, _ := time.ParseDuration(q)
				skipper := &FrameDropper{DelaySkip: dur}
				if q := query.Get("skipinterval"); q != "" {
					n := 0
					fmt.Sscanf(q, "%d", &n)
					skipper.SkipInterval = n
				}
				filters = append(filters, skipper)
			}

			demuxer := &pktque.FilterDemuxer{
				Filter:  filters,
				Demuxer: cursor,
			}

			muxer.WriteHeader(streams)
			avutil.CopyPackets(muxer, demuxer)
			muxer.WriteTrailer()
			// avutil.CopyFile(muxer, cursor)

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

	e.PUT("/regist", func(c echo.Context) error {

		user := new(User)
		if err := c.Bind(user); err != nil {
			return c.JSON(http.StatusBadRequest, "bad request")
		}
		username := user.Username
		pwd := user.Password
		if username == "" || pwd == "" {
			return c.JSON(http.StatusBadRequest, "bad request")
		}
		uhash := md5hash(username)
		userInfo := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			username,
			pwd,
		}
		userInfoJSON, err := json.Marshal(userInfo)
		if err != nil {
			c.JSON(http.StatusBadGateway, err.Error())
		}
		//check username if not exist
		model.DefaultCachePath = ".usercache"
		err = model.CreateBucket("userinfo")
		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}
		uinfo, db, err := model.SelectDB("userinfo", uhash)
		if err != nil {
			return c.JSON(http.StatusBadGateway, err.Error())
		}
		db.Close()
		if uinfo == nil {
			err = model.InsertDB("userinfo", uhash, string(userInfoJSON))
			if err != nil {
				return c.JSON(http.StatusBadGateway, err.Error())
			}
			return c.JSON(http.StatusOK, "success")
		}
		return c.JSON(http.StatusUnprocessableEntity, "duplicate username")
	})

	e.POST("/login", func(c echo.Context) error {
		user := new(User)
		if err := c.Bind(user); err != nil {
			return c.JSON(http.StatusBadRequest, "Bad Request")
		}
		username := user.Username
		password := user.Password
		if username == "" || password == "" {
			return c.JSON(http.StatusBadRequest, "Bad Request")
		}
		uhash := md5hash(username)
		model.DefaultCachePath = ".usercache"
		uinfo, _, err := model.SelectDB("userinfo", uhash)
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, err.Error())
		}
		// db.Close()
		realInfo := make(map[string]string)
		err = json.Unmarshal(uinfo, &realInfo)

		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, err.Error())
		}
		realPassword := realInfo["password"]
		if password == realPassword {
			cookie := new(http.Cookie)
			cookie.Name = "username"
			cookie.Value = username
			cookie.Expires = time.Now().Add(time.Hour * 12)
			c.SetCookie(cookie)
			return c.JSON(http.StatusOK, "OK")
		}
		return c.JSON(http.StatusBadRequest, "password not correct")
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

func md5hash(rawstr string) string {
	hasher := md5.New()
	hasher.Write([]byte(rawstr))
	return hex.EncodeToString(hasher.Sum(nil))
}
