package irc

import (
	"context"
	"fmt"
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
	IRCServer        string
	UserName         string
	Pass             string
	server           Server
	callbackHandlers map[string]EventCallback
}

//NewClient new client object with a defaut server setup
func NewClient(nick, password, serverName string) *Client {
	return &Client{
		UserName:         nick,
		Pass:             password,
		IRCServer:        serverName,
		callbackHandlers: make(map[string]EventCallback),
		server:           NewIRCServer(serverName, false),
	}
}

//HandleEventFunc handle event callbacks
func (c *Client) HandleEventFunc(event string, cb EventCallback) {
	c.callbackHandlers[event] = cb
}

//StartConnection connect to the irc server supplied in the Client object
func (c *Client) StartConnection() {
	connectCtx, cancel := context.WithCancel(context.Background())

	if err := c.server.start(connectCtx, c.UserName, c.Pass); err != nil {
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
			// TODO
			fmt.Printf("%v\n", line)

		case ping := <-c.server.pingChan:
			if callback, ok := c.callbackHandlers[EventPing]; ok {
				callback(EventType{
					Message: ping,
				})
			}

		case err := <-c.server.errChan:
			if callback, ok := c.callbackHandlers[EventError]; ok {
				callback(EventType{
					Message: err.Error(),
					Err:     err,
				})
			}

		case <-c.server.closeChan:
			cancel()
			if callback, ok := c.callbackHandlers[EventDisconnect]; ok {
				callback(EventType{})
			}

			c.server.wg.Wait()
			return
		}
	}
}

func (c *Client) StopConnection() {
	c.server.close()
}
