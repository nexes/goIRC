package irc

import (
	"context"
	"log"
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
	connectCtx, cancel := context.WithCancel(ctx)

	err := c.server.start(connectCtx)
	if err != nil {
		if callback, ok := c.callbackHandlers[EventError]; ok {
			callback(EventType{
				Err: err,
			})
		}

		if callback, ok := c.callbackHandlers[EventDisconnect]; ok {
			callback(EventType{
				Message: "Error from initial connect attempt",
				Err:     err,
			})
		}

		cancel()
		return
	}

	if callback, ok := c.callbackHandlers[EventConnect]; ok {
		callback(EventType{})
	}

	for {
		select {
		case line := <-c.server.recvChan:
			log.Printf("recvChan: %s", line)

		case ping := <-c.server.pingChan:
			log.Println("ping received ", ping)
			if callback, ok := c.callbackHandlers[EventPing]; ok {
				callback(EventType{})
			}

		case err := <-c.server.errChan:
			log.Println("server error: ", err.Error())
			if callback, ok := c.callbackHandlers[EventError]; ok {
				callback(EventType{
					Err: err,
				})
			}

		case <-ctx.Done():
			log.Println("client connect context done()")
			if callback, ok := c.callbackHandlers[EventDisconnect]; ok {
				callback(EventType{})
			}

			c.server.close()
			cancel()
			return
		}
	}
}
