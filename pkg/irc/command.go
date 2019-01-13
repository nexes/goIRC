package irc

import (
	"fmt"
	"strings"
)

type Command struct {
	Action string
	Args   []string
}

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

func (s *Server) join(room ...string) {
	if len(room) < 1 {
		return
	}

	rooms := strings.Join(room, ",")
	fmt.Println("join ", rooms)

	_, e := s.readWriter.WriteString(fmt.Sprintf("JOIN %s\r\n", rooms))
	if e != nil {
		s.errChan <- e
		return
	}

	s.readWriter.Flush()
}

func (s *Server) part(room, message string) {
	_, e := s.readWriter.WriteString(fmt.Sprintf("PART %s :%s\r\n", room, message))
	if e != nil {
		s.errChan <- e
		return
	}

	s.readWriter.Flush()
}

func (s *Server) invite(nick, room string) {
	_, e := s.readWriter.WriteString(fmt.Sprintf("INVITE %s %s\r\n", nick, room))
	if e != nil {
		s.errChan <- e
		return
	}

	s.readWriter.Flush()
}

func (s *Server) kick(user, room, message string) {
	_, e := s.readWriter.WriteString(fmt.Sprintf("KICK %s %s :%s\r\n", room, user, message))
	if e != nil {
		s.errChan <- e
		return
	}

	s.readWriter.Flush()
}

func (s *Server) privMessage(target, message string) {
	_, e := s.readWriter.WriteString(fmt.Sprintf("PRIVMSG %s :%s\r\n", target, message))
	if e != nil {
		s.errChan <- e
	}

	s.readWriter.Flush()
}

func (s *Server) list(scope ...string) {
	if len(scope) < 1 {
		return
	}

	scopes := strings.Join(scope, ",")

	_, e := s.readWriter.WriteString(fmt.Sprintf("LIST %s\r\n", scopes))
	if e != nil {
		s.errChan <- e
	}

	s.readWriter.Flush()
}

func (s *Server) name(scope ...string) {
	if len(scope) < 1 {
		return
	}

	scopes := strings.Join(scope, ",")

	_, e := s.readWriter.WriteString(fmt.Sprintf("NAMES %s\r\n", scopes))
	if e != nil {
		s.errChan <- e
	}

	s.readWriter.Flush()
}
