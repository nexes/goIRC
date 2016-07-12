package goirc

import (
	"bufio"
	"errors"
	"log"
	"sort"
	"strings"
)

type channel struct {
	chName   string
	topic    string
	username string
	nicks    []string

	connected bool
	active    bool //this the active channel the user is typing in (needed?)

	writer *bufio.Writer
}

func (c *channel) connect(recv chan []byte, io *bufio.ReadWriter) error {
	joincmd := "JOIN " + c.chName + "\r\n"

	_, err := io.Writer.Write([]byte(joincmd))
	if err != nil {
		errmsg := "error joining " + c.chName + ": " + err.Error()
		return errors.New(errmsg)
	}

	c.writer = io.Writer

	//flush out the buffer if needed
	if io.Writer.Buffered() > 0 {
		io.Writer.Flush()
	}

	//read until we hit the end of the name list, we parse nicks here
	for {
		read, err := io.Reader.ReadString('\n')
		if err != nil || strings.Contains(read, "366 "+c.username) {
			break
		}
		if strings.Contains(read, ":Forwarding to another channel") {
			c.chName = "#" + c.chName

		} else if strings.Contains(read, "332 "+c.username) {
			index := strings.Index(read, c.chName) + len(c.chName)
			c.topic = read[index:]

		} else if strings.Contains(read, "353 "+c.username) {
			index := strings.Index(read, c.chName+" :")
			c.updateNickList(read[index:])
		}
	}

	c.connected = true
	return nil
}

//SendMessage send a message to the channel
func (c *channel) SendMessage(msg string) {
	chat := "PRIVMSG " + c.chName + " :" + msg

	_, err := c.writer.Write([]byte(chat))
	if err != nil {
		log.Printf("Channel writing error %s", err.Error())
	}

	if c.writer.Buffered() > 0 {
		c.writer.Flush()
	}
}

//SendMessageToUser send a message to a user. if the nick is not found in the nick list the message wont be sent.
func (c *channel) SendMessageToUser(nick, msg string) {
	//can you send a message to a user outside of the channel?

}

//ChannelName returns the channel name
func (c *channel) ChannelName() string {
	return c.chName
}

//ChannelNicksList returns the nick list fo the channel
func (c *channel) ChannelNicksList() []string {
	return c.nicks
}

//ChannelTopic returns the channel topic if one was set
func (c *channel) ChannelTopic() string {
	if c.topic != "" {
		return c.topic
	}
	return "No Topic was set"
}

//you will need to update this list when users join/quit
func (c *channel) updateNickList(data string) {
	nicks := strings.Split(data, " ")

	for _, nick := range nicks {
		c.nicks = append(c.nicks, nick)
	}
	//check for duplicates
	sort.Strings(c.nicks)
}
