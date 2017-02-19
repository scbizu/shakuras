package chat

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

//Client is the user client .
type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	sendchannel chan []byte
}

const (
	//ReadMaxLimit defines the max size of read webscoket
	ReadMaxLimit = 512
	//ReadDeadline defines  the fresh msg(from clients).
	ReadDeadline = 60 * time.Second
	//WritePeriod defines  the fresh sending msg(from hub)
	WritePeriod = ReadDeadline * 9 / 10
	//WriterWait defines the writer rest time .
	WriterWait = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//ReadMsg will read msg from frontend websocket connection.
func (cli *Client) ReadMsg() error {
	defer func() {
		//ban quited user here
		//TODO: when user quited , log its status (notify frontend the quit msg)
		cli.hub.unregister <- cli
		cli.conn.Close()
	}()
	//configurations
	cli.conn.SetReadLimit(ReadMaxLimit)
	cli.conn.SetReadDeadline(time.Now().Add(ReadDeadline))
	cli.conn.SetPongHandler(func(appData string) error {
		cli.conn.SetReadDeadline(time.Now().Add(ReadDeadline))
		return nil
	})
	//websocket coroutine here .
	for {
		_, msg, err := cli.conn.ReadMessage()
		if err != nil {
			//websocket connection error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				return errors.New("read error")
			}
		}
		//solve msg and sp characters
		//TODO: filter some dirty words ;
		//msg = bytes.Trim(msg, "\n")
		cli.hub.broadcast <- msg
	}

}

//WriteMsg write the msg from hub to the client
func (cli *Client) WriteMsg() error {
	t := time.NewTicker(WritePeriod)
	defer func() {
		t.Stop()
		cli.conn.Close()
	}()
	//handle writer  op
	for {
		select {
		//comma,ok . type reflection
		case msg, ok := <-cli.sendchannel:
			cli.conn.SetWriteDeadline(time.Now().Add(WriterWait))
			if !ok {
				//hub has closed connection
				cli.conn.WriteMessage(websocket.PingMessage, []byte{})
				return errors.New("connection was closed by hub")
			}

			w, err := cli.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return err
			}
			w.Write(msg)

			sendlength := len(cli.sendchannel)
			for i := 0; i < sendlength; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-cli.sendchannel)
			}

			if err := w.Close(); err != nil {
				return err
			}
		case <-t.C:
			cli.conn.SetWriteDeadline(time.Now().Add(WriterWait))
			//timeout
			if err := cli.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return err
			}
		}
	}

}

//ServeWs serve the main process ..
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	cli := new(Client)
	cli.conn = conn
	cli.hub = hub
	cli.sendchannel = make(chan []byte, 256)
	//registe client
	cli.hub.register <- cli
	//write msg
	go cli.WriteMsg()
	//read msg
	err = cli.ReadMsg()
	return err
}
