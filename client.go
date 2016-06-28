package goirc

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strconv"
)

/*Client the main object that will hold information about what server to connect to
* If port isn't assigned, 6667 will be the default
* If SSL isn't assigned, false will be the default
 */
type Client struct {
	Server   string
	Port     int
	SSL      bool
	RecvData chan []byte

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
	c.RecvData = make(chan []byte)
	c.open = false

	raddrString := c.Server + ":" + strconv.Itoa(c.Port)

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

//ReadIt delet memememememem i dont know
// func (c *Client) ReadIt() chan []byte {
// 	return c.recvData
// }

//BeginListening reads from an open Client connection, returns data read or error
func (c *Client) BeginListening() {

	go func(recv chan []byte) {
		data := make([]byte, 1024)

		for c.open {
			read, err := c.connIO.Read(data)
			if err != nil {
				log.Printf("Error BeginListening: %v", err)
				break
			}
			recv <- data[:read]
		}
	}(c.RecvData)
}

//ConnectToChannel connects to a channel, returns an error if already connected
func (c *Client) ConnectToChannel(channel string) error {
	return errors.New("already connected to room ")
}

//DisconnectFromChannel disconnects from a channel, returns an error if not connected to that channel
func (c *Client) DisconnectFromChannel(channel string) error {
	return errors.New("what ain")
}
