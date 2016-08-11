package goirc

import (
	"bufio"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
)

//Channel data on connected channel
type Channel struct {
	Name  string
	Nicks []string

	topic     string
	username  string
	connected bool
	active    bool //this the active channel the user is typing in (needed?)

	nickListLock *sync.Mutex
	writer       *bufio.Writer
}

func (c *Channel) connect(io *bufio.ReadWriter) error {
	joincmd := "JOIN " + c.Name + "\r\n"

	_, err := io.Writer.WriteString(joincmd)
	if err != nil {
		errmsg := "error joining " + c.Name + ": " + err.Error()
		return errors.New(errmsg)
	}

	//flush out the buffer if needed
	if io.Writer.Buffered() > 0 {
		io.Writer.Flush()
	}

	c.writer = io.Writer
	c.connected = true
	return nil
}

//SendMessage send a message to the channel
func (c *Channel) SendMessage(msg string) {
	chat := "PRIVMSG " + c.Name + " :" + msg

	_, err := c.writer.WriteString(chat)
	if err != nil {
		log.Printf("Channel writing error %s", err.Error())
	}

	if c.writer.Buffered() > 0 {
		c.writer.Flush()
	}
}

//SendMessageToUser send a message to a user. if the nick is not found in the nick list the message wont be sent.
func (c *Channel) SendMessageToUser(nick, msg string) error {
	if !sort.StringsAreSorted(c.Nicks) {
		sort.Strings(c.Nicks)
	}

	index := sort.SearchStrings(c.Nicks, nick)
	if index < len(c.Nicks) && strings.EqualFold(nick, c.Nicks[index]) {

		_, err := c.writer.WriteString("PRIVMSG " + nick + " :" + msg)
		if err != nil {
			return err
		}

		if c.writer.Buffered() > 0 {
			c.writer.Flush()
		}
		return nil
	}
	return errors.New("Nick " + nick + " wasn't found to send a message to")
}

//Topic returns the channel topic if one was set
func (c *Channel) Topic() string {
	if c.topic != "" {
		return c.topic
	}
	return "No Topic was set"
}

//you will need to update this list when users join/quit
func (c *Channel) updateNickList(data string) {
	c.nickListLock.Lock()
	defer c.nickListLock.Unlock()

	nicks := strings.Split(data, " ")
	c.Nicks = append(c.Nicks, nicks[0:]...)
	sort.Strings(c.Nicks)
}
