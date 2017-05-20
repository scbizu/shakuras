package kara

//Hub is the client hub .
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

//NewHub  init the hub instance .
func NewHub() *Hub {

	return &Hub{
		clients:    make(map[*Client]bool, 0),
		broadcast:  make(chan []byte, 0),
		register:   make(chan *Client, 0),
		unregister: make(chan *Client, 0),
	}
}

//Run will run the hub instance and serve .
func (h *Hub) Run() error {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			client.sendchannel <- []byte("Nace > 来了一个新的基佬")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.sendchannel <- []byte("Nace > 一个基佬走了..")
				close(client.sendchannel)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.sendchannel <- message:
				default:
					delete(h.clients, client)
					close(client.sendchannel)
				}
			}
		}
	}
}
