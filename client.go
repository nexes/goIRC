package goirc

import (
	"bufio"
	"errors"
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

	RecvChanel chan []byte //receving data from a channle, e.g PRIVMSG
	RecvServer chan []byte //receving data from the server, e.g CONNECT/DISCONNECT

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
	c.ircChannels = make([]channel, 5)
	c.RecvServer = make(chan []byte)
	c.RecvChanel = make(chan []byte)
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
	c.open = false
	c.conn.Close()
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

//BeginListening reads from an open Client connection, returns data read or error
func (c *Client) BeginListening() {
	var userMsg string

	if c.Password != "" {
		userMsg = "PASS " + c.Password + "\r\nNICK " + c.Nick + "\r\nUSER " + c.Nick + " 0 * :goirc bot\r\n"
	} else {
		userMsg = "NICK " + c.Nick + "\r\nUSER " + c.Nick + " 0 * :goirc bot\r\n"
	}

	go func(rSrv, rChn chan []byte) {
		data := make([]byte, 1024)

		// tell IRC who we are
		_, err := c.connIO.Write([]byte(userMsg))
		if err != nil {
			log.Printf("Error writing to IRC: %v", err)
		}
		if c.connIO.Writer.Buffered() > 0 {
			c.connIO.Flush()
		}

		for c.open {
			read, err := c.connIO.Reader.Read(data)
			if err != nil {
				log.Printf("Error BeginListening: %v", err)
				break
			}

			if strings.Contains(string(data[:read]), "PRIVMSG") {
				rChn <- data[:read]
			} else {
				rSrv <- data[:read]
			}
		}
	}(c.RecvServer, c.RecvChanel)
}

//ConnectToChannel connects to a channel, returns an error if already connected
func (c *Client) ConnectToChannel(chName string) error {
	if chName[0] != '#' {
		chName = "#" + chName
	}

	nc := channel{
		name:      chName,
		connected: false,
		active:    false,
	}

	err := nc.connect(c.connIO)
	if err != nil {
		return err
	}

	c.ircChannels = append(c.ircChannels, nc)
	return nil
}

//DisconnectFromChannel disconnects from a channel, returns an error if not connected to that channel
func (c *Client) DisconnectFromChannel(channel string) error {
	return errors.New("what ain")
}
