package irc

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"time"
)

const (
	EventPing       = "PING"
	EventConnect    = "CONNECT"
	EventDisconnect = "DISCONNECT"
	EventRoomJoin   = "ROOMJOIN"
	EventRoomLeave  = "ROOMLEAVE"

	eventCount = 5
)

type eventFunc func()

//Client object describing the irc connection
type Client struct {
	UserName         string
	server           Server
	callbackHandlers map[string]eventFunc

	readWriter *bufio.ReadWriter
	connection net.Conn
}

//NewClientWithServer new client object
func NewClientWithServer(nick string, server Server) *Client {
	return &Client{
		UserName:         nick,
		server:           server,
		callbackHandlers: make(map[string]eventFunc, eventCount),
	}
}

//NewClient new client object with a defaut server setup
func NewClient(nick, serverName string) *Client {
	return &Client{
		UserName:         nick,
		callbackHandlers: make(map[string]eventFunc, eventCount),
		server: Server{
			ServerName: serverName,
			Port:       6667,
			UseTSL:     false,
			Timeout:    time.Second * 60,
		},
	}
}

//HandleEventFunc handle event callbacks
func (c *Client) HandleEventFunc(event string, f eventFunc) {
	c.callbackHandlers[event] = f
}

//Connect connect to the irc server supplied in the Client object
func (c *Client) Connect(ctx context.Context) error {
	dialer := net.Dialer{}
	dialer.Timeout = c.server.Timeout

	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", c.server.ServerName, c.server.Port))
	if err != nil {
		return err
	}

	c.connection = conn
	c.readWriter = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	callback, ok := c.callbackHandlers[EventConnect]
	if ok {
		callback()
	}

	return nil
}
