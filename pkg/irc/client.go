package irc

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
)

const (
	EventPing       = "PING"
	EventConnect    = "CONNECT"
	EventDisconnect = "DISCONNECT"
	EventError      = "ERROR"
	EventRoomJoin   = "ROOMJOIN"
	EventRoomLeave  = "ROOMLEAVE"
)

type EventType struct {
	Message string
	Err     error
}

type EventCallback func(EventType)

//Client object describing the irc connection
type Client struct {
	UserName         string
	IRCServer        string
	server           Server
	callbackHandlers map[string]EventCallback

	readWriter *bufio.ReadWriter
	connection net.Conn
}

//NewClient new client object with a defaut server setup
func NewClient(nick, serverName string) *Client {
	return &Client{
		UserName:         nick,
		IRCServer:        serverName,
		callbackHandlers: make(map[string]EventCallback),
		server:           NewIRCServer(serverName, false),
	}
}

//HandleEventFunc handle event callbacks
func (c *Client) HandleEventFunc(event string, cb EventCallback) {
	c.callbackHandlers[event] = cb
}

//Connect connect to the irc server supplied in the Client object
func (c *Client) Connect(ctx context.Context) {
	dialer := net.Dialer{}
	dialer.Timeout = c.server.Timeout
	connectCtx, cancel := context.WithCancel(ctx)

	conn, err := dialer.DialContext(connectCtx, "tcp", fmt.Sprintf("%s:%d", c.server.ServerName, c.server.Port))
	if err != nil {
		callback, ok := c.callbackHandlers[EventDisconnect]
		if ok {
			callback(EventType{
				Err: err,
			})
		}

		c.server.close()
		cancel()
		return
	}

	c.connection = conn
	c.readWriter = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	callback, ok := c.callbackHandlers[EventConnect]
	if ok {
		callback(EventType{})
	}

	c.server.start(connectCtx, c.readWriter)
	for {
		select {
		case line := <-c.server.recvChan:
			log.Printf("recvChan: %s", line)

		case ping := <-c.server.pingChan:
			log.Println("ping received ", ping)
			callback, ok := c.callbackHandlers[EventPing]
			if ok {
				callback(EventType{})
			}

		case err := <-c.server.errChan:
			log.Println("server error: ", err.Error())
			callback, ok := c.callbackHandlers[EventError]
			if ok {
				callback(EventType{
					Err: err,
				})
			}

		case <-ctx.Done():
			log.Println("client connect context done()")
			callback, ok := c.callbackHandlers[EventDisconnect]
			if ok {
				callback(EventType{})
			}

			c.server.close()
			cancel()
			return
		}
	}
}
