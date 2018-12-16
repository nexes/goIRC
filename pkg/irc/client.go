package irc

import (
	"context"
	"errors"
	"time"
)

const (
	EventPing        = "PING"
	EventConnect     = "CONNECT"
	EventDisconnect  = "DISCONNECT"
	EventError       = "ERROR"
	EventRoomMessage = "ROOMMESSAGE"
	EventRoomLeave   = "ROOMLEAVE"
	EventMOTD        = "EVENTMOTD"
	EventMessage     = "EVENTMESSAGE"
)

type EventType struct {
	Message string
	Server  string
	Nick    string
	Room    string
	Code    int32
	Err     error
	Time    time.Time
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

	c.listenToChannels(cancel)
}

//JoinRoom join a room on the connected irc server
func (c *Client) JoinRoom(room string) {
	c.server.join(room)
}

//StopConnection closes and disconnects from the irc server. This will stop the blocking nature of
//StartConnection
func (c *Client) StopConnection() {
	c.server.close()
}

//this will block, listening for any data coming in from the server channels and send
//the data to the correct callback
func (c *Client) listenToChannels(cancel context.CancelFunc) {
	for {
		select {
		case line := <-c.server.recvChan:
			switch line.Code {
			case RPL_WELCOME, RPL_YOURHOST, RPL_CREATED, RPL_MYINFO, RPL_BOUNCE:
				if callback, ok := c.callbackHandlers[EventConnect]; ok {
					callback(EventType{
						Message: line.Message,
						Server:  line.ServerName,
						Code:    line.Code,
						Time:    line.Time,
					})
				}

			case RPL_MOTD, RPL_ENDOFMOTD:
				if callback, ok := c.callbackHandlers[EventMOTD]; ok {
					callback(EventType{
						Message: line.Message,
						Code:    line.Code,
						Server:  line.ServerName,
						Time:    line.Time,
					})
				}

			case RPL_TOPIC, RPL_NAMREPLY, RPL_ENDOFNAMES, RPL_FORWARDJOIN, RPL_ROOMJOIN, RPL_ROOMPART, RPL_ROOMQUIT:
				if callback, ok := c.callbackHandlers[EventRoomMessage]; ok {
					callback(EventType{
						Server:  line.ServerName,
						Code:    line.Code,
						Nick:    line.Nick,
						Room:    line.Room,
						Message: line.Message,
						Time:    line.Time,
					})
				}

			case RPL_ERRORJOIN:
				if callback, ok := c.callbackHandlers[EventError]; ok {
					callback(EventType{
						Server:  line.ServerName,
						Room:    line.Room,
						Code:    line.Code,
						Message: line.Message,
						Err:     errors.New(line.Message),
					})
				}

			case RPL_PRIVMSG:
				if callback, ok := c.callbackHandlers[EventMessage]; ok {
					callback(EventType{
						Server:  line.ServerName,
						Code:    line.Code,
						Nick:    line.Nick,
						Room:    line.Room,
						Message: line.Message,
						Time:    line.Time,
					})
				}
			}

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
