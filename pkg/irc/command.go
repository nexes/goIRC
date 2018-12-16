package irc

import (
	"fmt"
)

func (s *Server) ping() {
	s.readWriter.WriteString(fmt.Sprintf("PONG %s", s.ServerName))
	s.readWriter.Flush()
}

func (s *Server) pass(password string) {
	if len(password) > 1 {
		s.readWriter.WriteString(fmt.Sprintf("PASS %s\r\n", password))
		s.readWriter.Flush()
	}
}

func (s *Server) user(username string) {
	s.readWriter.WriteString(fmt.Sprintf("NICK %s\r\n", username))
	s.readWriter.WriteString(fmt.Sprintf("USER %s %d * :GoIRC bot\r\n", username, 0))
	s.readWriter.Flush()
}

func (s *Server) join(room string) {
	if room[0] != '#' {
		room = "#" + room
	}
	_, e := s.readWriter.WriteString(fmt.Sprintf("JOIN %s\r\n", room))
	if e != nil {
		fmt.Println("Join error ", e)
	}
	s.readWriter.Flush()
}
