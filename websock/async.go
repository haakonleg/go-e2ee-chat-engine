package websock

import "golang.org/x/net/websocket"

// Packet is a websocket message with an included error message which might
// come from a closed connection
type Packet struct {
	Msg Message
	Err error
}

// AsyncConn is an asynchronous wrapper for a websocket connection
type AsyncConn struct {
	ws   *websocket.Conn
	send chan Packet
	wait chan struct{}
	stop chan struct{}
}

// NewAsyncConn returns a new asynchronous websocket connection
func NewAsyncConn(ws *websocket.Conn) *AsyncConn {
	return &AsyncConn{
		ws,
		make(chan Packet, 1),
		make(chan struct{}, 1),
		make(chan struct{}, 1),
	}
}

// Run the given async connection in a separate goroutine
func (conn *AsyncConn) Run() {
	var pk Packet
	for {
		select {
		// Only continue when were asked too. Other handlers may
		// send/receive on the same websocket, so we do not want to
		// intervene
		case <-conn.wait:
			pk.Err = Msg.Receive(conn.ws, &pk.Msg)
			conn.send <- pk

			// TODO does an error always mean that the connection is closed? If
			// not this should be done differently
			if pk.Err != nil {
				return
			}
		case <-conn.stop:
			return
		}
	}
}

// Get fetches a packet from the connection through a channel
func (conn *AsyncConn) Get() <-chan Packet {
	conn.wait <- struct{}{}
	return conn.send
}

// Conn returns the inner websocket connection
func (conn *AsyncConn) Conn() *websocket.Conn {
	return conn.ws
}

// Close shuts the internal getter down and closes the websocket connection
func (conn *AsyncConn) Close() {
	conn.stop <- struct{}{}
	conn.ws.Close()
}
