package goirc

import (
	"bufio"
	// "errors"
	"log"
	"net"
	"strconv"
	"strings"
)

/*Client the main object that will hold information about what server to connect to
* If port isn't assigned, 6667 will be the default
* If SSL isn't assigned, false will be the default
 */
type Client struct {
	Server   string
	Port     int
	SSL      bool
	Nick     string //user refactor to its own obj?
	Password string //user refactor to its own obj?

	RecvFromServer  chan []byte //receving data from a channle, e.g PRIVMSG
	RecvFromChannel chan []byte //receving data from the server, e.g CONNECT/DISCONNECT

	open        bool
	ircChannels []channel

	conn   *net.TCPConn
	connIO *bufio.ReadWriter
}

//ConnectToServer connects to the server:port described by the Client object
func (c *Client) ConnectToServer() error {
	if c.Port == 0 {
		c.Port = 6667
	}
	if c.SSL {
		c.Port = 6697
	}
	c.ircChannels = make([]channel, 0, 5)
	c.RecvFromChannel = make(chan []byte)
	c.RecvFromServer = make(chan []byte)
	c.open = false

	raddrString := net.JoinHostPort(c.Server, strconv.Itoa(c.Port))

	raddr, err := net.ResolveTCPAddr("tcp", raddrString)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return err
	}
	c.conn = conn
	c.connIO = bufio.NewReadWriter(
		bufio.NewReader(c.conn),
		bufio.NewWriter(c.conn),
	)

	c.open = true
	c.conn.SetKeepAlive(true)
	return nil
}

//IsOpen will return a bool showing if the connection to the server is still open
func (c *Client) IsOpen() bool {
	//make this better obviously
	return c.open
}

//CloseConnection closes the TCP connection to the server, closes any irc channels that may be left
func (c *Client) CloseConnection() {
	log.Println("closed called")
	c.open = false
	c.connIO.Writer.Flush()
	c.connIO.Reader.Discard(c.connIO.Reader.Buffered())

	c.conn.Close()
	//channels take care of them selfs?
}

//SendPongResponse will send a PONG response when a PING request was recieved
func (c *Client) SendPongResponse() {
	_, err := c.connIO.Writer.Write([]byte("PONG " + c.Server + "\r\n"))
	if err != nil {
		//return this duh
		log.Println(err.Error)
	}

	if c.connIO.Writer.Buffered() > 0 {
		c.connIO.Writer.Flush()
	}
}

//ConnectAndListen reads from an open Client connection, returns data read or error
func (c *Client) ConnectAndListen() {
	var userMsg string

	if c.Password != "" {
		userMsg = "PASS " + c.Password + "\r\nNICK " + c.Nick + "\r\nUSER " + c.Nick + " 0 * :goirc bot\r\n"
	} else {
		userMsg = "NICK " + c.Nick + "\r\nUSER " + c.Nick + " 0 * :goirc bot\r\n"
	}

	//need to do error handling, 400's
	go func(rSrv, rChn chan []byte) {
		// tell IRC who we are
		_, err := c.connIO.Write([]byte(userMsg))
		if err != nil {
			log.Printf("Error writing to IRC: %v", err)
		}
		if c.connIO.Writer.Buffered() > 0 {
			c.connIO.Flush()
		}

		for c.open {
			read, err := c.connIO.Reader.ReadString('\n')
			if err != nil {
				log.Printf("Error ConnectAndListen: %v", err)
				break
			}

			//this is a cheap hack, you need a better way to know if its a prvmsg
			if strings.Contains(read, "PRIVMSG #") {
				rChn <- []byte(read)
			} else {
				rSrv <- []byte(read)
			}
		}
	}(c.RecvFromServer, c.RecvFromChannel)
}

//ConnectToChannel connects to a channel, returns an error if already connected
func (c *Client) ConnectToChannel(chName string) error {
	if chName[0] != '#' {
		chName = "#" + chName
	}

	nc := channel{
		chName:    chName,
		username:  c.Nick,
		connected: false,
		active:    false,
		nicks:     make([]string, 0, 100),
	}

	err := nc.connect(c.RecvFromChannel, c.connIO)
	if err != nil {
		return err
	}

	c.ircChannels = append(c.ircChannels, nc)
	return nil
}

//DisconnectFromChannel disconnects from a channel, returns an error if not connected to that channel
func (c *Client) DisconnectFromChannel(ch *channel, msg string) error {
	if msg == "" {
		msg = "goirc by nexes" //temp right now
	}
	ch.connected = false

	_, err := c.connIO.Writer.Write([]byte("PART " + ch.chName + " :" + msg))
	if err != nil {
		return err
	}
	if c.connIO.Writer.Buffered() > 0 {
		c.connIO.Writer.Flush()
	}

	index := -1
	for i, v := range c.ircChannels {
		if v.chName == ch.chName {
			index = i
			break
		}
	}

	//make sure this is working
	if index >= 0 {
		c.ircChannels = append(c.ircChannels[:index], c.ircChannels[index+1:]...)
	}

	return nil
}

//GetChannel will return the channel object if one exists for the channel name given
func (c *Client) GetChannel(name string) *channel {
	for _, ch := range c.ircChannels {
		if strings.Contains(ch.chName, name) {
			return &ch
		}
	}
	return nil
}

//ChangeNick change your current NICK
func (c *Client) ChangeNick(nick string) error {
	return nil
}
