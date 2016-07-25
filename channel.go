package goirc

import (
	"bufio"
	"errors"
	"log"
	"sort"
	"strings"
)

//Channel data on connected channel
type Channel struct {
	chName   string
	topic    string
	username string
	nicks    []string

	connected bool
	active    bool //this the active channel the user is typing in (needed?)

	writer *bufio.Writer
}

func (c *Channel) connect(io *bufio.ReadWriter) error {
	joincmd := "JOIN " + c.chName + "\r\n"

	_, err := io.Writer.WriteString(joincmd)
	if err != nil {
		errmsg := "error joining " + c.chName + ": " + err.Error()
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
	chat := "PRIVMSG " + c.chName + " :" + msg

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
	if !sort.StringsAreSorted(c.nicks) {
		sort.Strings(c.nicks)
	}
	index := sort.SearchStrings(c.nicks, nick)
	if index < len(c.nicks) && strings.EqualFold(nick, c.nicks[index]) {

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

//Name returns the channel name
func (c *Channel) Name() string {
	return c.chName
}

//NickList returns the nick list fo the channel
func (c *Channel) NickList() []string {
	return c.nicks
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
	nicks := strings.Split(data, " ")
	c.nicks = append(c.nicks, nicks[0:]...)

	//check for duplicates
	sort.Strings(c.nicks)
}
