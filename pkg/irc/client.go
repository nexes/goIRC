package irc

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
)

type eventFunc func()

const (
	EventPing       = "PING"
	EventConnect    = "CONNECT"
	EventDisconnect = "DISCONNECT"
	EventRoomJoin   = "ROOMJOIN"
	EventRoomLeave  = "ROOMLEAVE"

	eventCount = 5
)

//Client object describing the irc connection
type Client struct {
	UserName         string
	IRCServer        string
	server           Server
	callbackHandlers map[string]eventFunc

	readWriter *bufio.ReadWriter
	connection net.Conn
}

//NewClient new client object with a defaut server setup
func NewClient(nick, serverName string) *Client {
	return &Client{
		UserName:         nick,
		IRCServer:        serverName,
		callbackHandlers: make(map[string]eventFunc, eventCount),
		server:           NewIRCServer(serverName, false),
	}
}

//HandleEventFunc handle event callbacks
func (c *Client) HandleEventFunc(event string, f eventFunc) {
	c.callbackHandlers[event] = f
}

//Connect connect to the irc server supplied in the Client object
func (c *Client) Connect(ctx context.Context) {
	dialer := net.Dialer{}
	dialer.Timeout = c.server.Timeout

	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", c.server.ServerName, c.server.Port))
	if err != nil {
		callback, ok := c.callbackHandlers[EventDisconnect]
		if ok {
			callback()
		}

		c.server.close()
		ctx.Done()
		return
	}

	c.connection = conn
	c.readWriter = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	callback, ok := c.callbackHandlers[EventConnect]
	if ok {
		callback()
	}

	c.server.start(ctx, c.readWriter)
	for {
		select {
		case line := <-c.server.recvChan:
			log.Printf("recvChan: %s", line)

		case ping := <-c.server.pingChan:
			log.Println("ping received ", ping)
			callback, ok := c.callbackHandlers[EventPing]
			if ok {
				callback()
			}

		case <-ctx.Done():
			// todo cleanup
			log.Println("context done()")
			c.server.close()

			callback, ok := c.callbackHandlers[EventDisconnect]
			if ok {
				callback()
			}

			return
		}
	}
}
