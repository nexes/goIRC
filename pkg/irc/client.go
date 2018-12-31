package irc

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	EventPing           = "PING"
	EventConnect        = "CONNECT"
	EventDisconnect     = "DISCONNECT"
	EventError          = "ERROR"
	EventRoomMessage    = "ROOMMESSAGE"
	EventChannelMessage = "CHANNELMESSAGE"
	EventMOTD           = "EVENTMOTD"
	EventMessage        = "EVENTMESSAGE"
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

//WriteToTarget will send the message to the target, e.g the room, a user etc
func (c *Client) WriteToTarget(target string, message string) {
	c.server.privMessage(target, message)
}

//StopConnection closes and disconnects from the irc server. This will stop the blocking nature of
func (c *Client) StopConnection() {
	c.server.close()
}

func (c *Client) Command(command Command) {
	// TODO
	if len(command.Action) > 0 {
		command.Action = strings.ToLower(strings.TrimSpace(command.Action))
		command.Args = strings.ToLower(strings.TrimSpace(command.Args))

		c.server.sendChan <- command
	}
}

//this will block, listening for any data coming in from the server channels and send the data to the correct callback
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

			case RPL_LIST, RPL_LISTEND, RPL_FORWARDJOIN, RPL_NAMREPLY, RPL_ENDOFNAMES:
				if callback, ok := c.callbackHandlers[EventChannelMessage]; ok {
					callback(EventType{
						Server:  line.ServerName,
						Code:    line.Code,
						Nick:    line.Nick,
						Room:    line.Room,
						Time:    line.Time,
						Message: line.Message,
					})
				}

			case RPL_TOPIC, RPL_ROOMJOIN, RPL_ROOMPART, RPL_ROOMQUIT:
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
