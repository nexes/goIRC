package goirc

import (
	"bufio"
	"fmt"
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
	Nick     string
	Password string

	RecvFromServer  chan []byte //receving data from a channle, e.g PRIVMSG
	RecvFromChannel chan []byte //receving data from the server, e.g CONNECT/DISCONNECT

	open        bool
	ircChannels []Channel

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
	c.ircChannels = make([]Channel, 0, 5)
	c.RecvFromChannel = make(chan []byte)
	c.RecvFromServer = make(chan []byte)
	c.open = false

	raddrString := net.JoinHostPort(c.Server, strconv.Itoa(c.Port))

	raddr, err := net.ResolveTCPAddr("tcp", raddrString)
	if err != nil {
		log.Printf("resolve tcp for %s: %s", raddrString, err.Error())
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
func (c *Client) CloseConnection(msg string) {
	c.open = false

	_, err := c.connIO.Writer.WriteString("QUIT :" + msg + "\r\n")
	if err != nil {
		log.Printf("Error closing: %s", err.Error())
	}
	c.connIO.Writer.Flush()
}

//SendPongResponse will send a PONG response when a PING request was recieved
func (c *Client) SendPongResponse() {
	_, err := c.connIO.Writer.WriteString("PONG " + c.Server + "\r\n")
	if err != nil {
		//return this duh
		log.Println(err.Error())
	}

	if c.connIO.Writer.Buffered() > 0 {
		c.connIO.Writer.Flush()
	}
}

//Listen reads from an open Client connection, returns data read or error
func (c *Client) Listen() {
	var userMsg string

	if c.Password != "" {
		userMsg = fmt.Sprintf("PASS %s\r\nNICK %s\r\nUSER %s 0 * :goirc bot\r\n", c.Password, c.Nick, c.Nick)
	} else {
		userMsg = fmt.Sprintf("NICK %s\r\nUSER %s 0 * :goirc bot\r\n", c.Nick, c.Nick)
	}

	_, err := c.connIO.Write([]byte(userMsg))
	if err != nil {
		log.Printf("Error writing to IRC: %s", err.Error())
	}
	if c.connIO.Writer.Buffered() > 0 {
		c.connIO.Flush()
	}

	//this will handle reading from the connection.
	go func(rSrv, rChn chan []byte) {
		defer c.conn.Close()

		for c.open {
			read, err := c.connIO.Reader.ReadString('\n')
			if err != nil {
				log.Printf("Error ConnectAndListen: %s", err.Error())
				break
			}

			//gets rid of the leading ":", easier to parse without this char
			if read[0] == ':' {
				read = read[1:]
			}

			//refactor these to their own space
			if strings.Contains(read, "001 "+c.Nick) {
				host := read[:strings.Index(read, " ")]
				c.Server = host

			} else if strings.Contains(read, c.Server+" 470 "+c.Nick) {
				ch := read[len(c.Server+" 470 "+c.Nick):strings.LastIndex(read, ":")]
				chlist := strings.Split(strings.TrimSpace(ch), " ")

				for i, v := range c.ircChannels {
					if v.chName == chlist[0] {
						c.ircChannels[i].chName = chlist[len(chlist)-1]
						break
					}
				}

			} else if strings.Contains(read, c.Server+" 332 "+c.Nick) {
				topic := read[strings.Index(read, ":")+1:]
				ch := read[strings.Index(read, "#"):strings.Index(read, ":")]

				for i, v := range c.ircChannels {
					if strings.Contains(ch, v.chName) {
						c.ircChannels[i].topic = topic
						break
					}
				}

			} else if strings.Contains(read, c.Server+" 353 "+c.Nick) {
				nick := read[strings.Index(read, ":")+1:]
				nicklist := strings.Split(nick, " ")
				chn := &c.ircChannels[len(c.ircChannels)-1] //get the last one added

				chn.nicks = append(chn.nicks, nicklist...)
			}

			//this is a cheap hack, you need a better way to know if its a prvmsg
			if strings.Contains(read, "PRIVMSG #") {
				msg := FormatPrivMsg(read)
				rChn <- []byte(msg)
			} else {
				rSrv <- []byte(read)
			}
		}
	}(c.RecvFromServer, c.RecvFromChannel)
}

//JoinChannel connects to a channel, returns an error if already connected
func (c *Client) JoinChannel(chName string) (*Channel, error) {
	if chName[0] != '#' {
		chName = "#" + chName
	}

	nc := Channel{
		chName:    strings.TrimSpace(chName),
		username:  c.Nick,
		connected: false,
		active:    false,
		nicks:     make([]string, 0, 200),
	}

	err := nc.connect(c.connIO)
	if err != nil {
		return nil, err
	}

	c.ircChannels = append(c.ircChannels, nc)
	return &nc, nil
}

//LeaveChannel disconnects from a channel, returns an error if not connected to that channel
func (c *Client) LeaveChannel(ch *Channel, msg string) error {
	ch.connected = false

	_, err := c.connIO.Writer.WriteString("PART " + ch.chName + " :" + msg + "\r\n")
	if err != nil {
		return err
	}
	c.connIO.Writer.Flush()

	for index, v := range c.ircChannels {
		if v.chName == ch.chName {
			if index > 0 {
				c.ircChannels = append(c.ircChannels[:index], c.ircChannels[index+1:]...)

			} else if index == 0 {
				c.ircChannels = []Channel{}
			}
		}
	}
	return nil
}

//GetChannel will return the channel object if one exists for the channel name given
func (c *Client) GetChannel(name string) *Channel {
	for _, ch := range c.ircChannels {
		if strings.Contains(ch.chName, strings.TrimSpace(name)) {
			return &ch
		}
	}
	return nil
}

//ChangeNick change your current NICK
func (c *Client) ChangeNick(nick string) error {
	if c.Nick != nick {
		c.Nick = strings.TrimSpace(nick)
	}

	//update the new nick with all open channels
	for _, v := range c.ircChannels {
		v.username = c.Nick
	}

	_, err := c.connIO.Writer.WriteString("NICK " + c.Nick + "\r\n")
	if err != nil {
		return err
	}
	if c.connIO.Writer.Buffered() > 0 {
		c.connIO.Writer.Flush()
	}
	return nil
}

//IdentifyNick sends the NickServ identify command to the server to register the nick
func (c *Client) IdentifyNick(pass string) error {
	msg := "/msg NickServ identify " + pass + "\r\n"

	_, err := c.connIO.Writer.WriteString(msg)
	if err != nil {
		return err
	}

	if c.connIO.Writer.Buffered() > 0 {
		c.connIO.Writer.Flush()
	}
	return nil
}

//QueryWho return information about the nick passed. Mask is defaulted to 0 if left blank
func (c *Client) QueryWho(nick, mask string) error {
	if mask == "" {
		mask = "0"
	}

	_, err := c.connIO.Writer.WriteString("WHO " + nick + "\r\n")
	if err != nil {
		return err
	}

	if c.connIO.Writer.Buffered() > 0 {
		c.connIO.Writer.Flush()
	}

	return nil
}

//QueryWhoWas returns the recent nicks the user has used
func (c *Client) QueryWhoWas(nick string) error {
	_, err := c.connIO.Writer.WriteString("WHOWAS " + nick + "\r\n")
	if err != nil {
		return err
	}

	if c.connIO.Writer.Buffered() > 0 {
		c.connIO.Writer.Flush()
	}

	return nil
}
