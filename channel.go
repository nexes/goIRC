package goirc

import (
	"bufio"
	"errors"
)

type channel struct {
	name      string
	connected bool
	active    bool //this the active channel the user is typing in (needed)
}

func (c *channel) connect(io *bufio.ReadWriter) error {
	joincmd := "JOIN " + c.name + "\r\n"

	_, err := io.Writer.Write([]byte(joincmd))
	if err != nil {
		errmsg := "error joining " + c.name + ": " + err.Error()
		return errors.New(errmsg)
	}

	//flush out the buffer if needed
	if io.Writer.Buffered() > 0 {
		io.Writer.Flush()
	}

	c.connected = true
	return nil
}

//SendMessage send a message to the channel
func (c *channel) SendMessage(msg string) {

}

func parseNickList(data []byte) {

}
