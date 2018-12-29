package irc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

//Server describe the irc server
type Server struct {
	ServerName string
	Port       int32
	UseTSL     bool
	running    bool

	Timeout  time.Duration
	PingFreq time.Duration

	readWriter *bufio.ReadWriter
	conn       net.Conn

	wg sync.WaitGroup
	//TODO: these will neeed to be a custom struct to handle more data; make buffered
	recvChan  chan IncomingData
	sendChan  chan Command
	errChan   chan error
	pingChan  chan string
	closeChan chan struct{}
}

//NewIRCServer create a new irc server
func NewIRCServer(server string, useTLS bool) Server {
	return Server{
		ServerName: server,
		Port:       6667,
		UseTSL:     useTLS,
		running:    false,

		Timeout:  time.Minute * 3,
		PingFreq: time.Minute * 2,

		recvChan:  make(chan IncomingData),
		sendChan:  make(chan Command),
		errChan:   make(chan error),
		pingChan:  make(chan string),
		closeChan: make(chan struct{}),
	}
}

//start will make the initial irc connection and start the needed go routines if no errors occured
func (s *Server) start(ctx context.Context, username, password string) error {
	if !s.running {
		s.running = true

		dialer := net.Dialer{
			Timeout: s.Timeout,
		}

		conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", s.ServerName, s.Port))
		if err != nil {
			s.running = false
			return err
		}

		s.conn = conn
		s.readWriter = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

		if len(password) > 1 {
			s.pass(password)
		}
		s.user(username)

		s.wg.Add(3)
		go s.recv(ctx)
		go s.send(ctx)
		go s.sendPingResponse(ctx)

		return nil
	}

	return errors.New("Calling start on a running server")
}

//block waiting to receive anything from the connected irc server, if a I/O error happens, the connection
//will be cloased and the server disconnected
func (s *Server) recv(ctx context.Context) {
	defer s.wg.Done()

	for {
		data, err := s.readWriter.ReadString('\n')
		if err != nil {
			s.errChan <- err
			s.closeChan <- struct{}{}

			if s.running {
				s.running = false
				s.conn.Close()
			}
			break
		}

		if recData, ok := parseRawInput(data); ok {
			s.recvChan <- recData
		}
	}

	<-ctx.Done()
}

//send will send user commands to the connected irc server
func (s *Server) send(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case command := <-s.sendChan:
			switch command.Action {
			case "join":
				s.join(command.Args)
			case "list":
				s.list(command.Args)
			}

		case <-ctx.Done():
			return
		}
	}
}

//send our automatic ping/pong responses
func (s *Server) sendPingResponse(ctx context.Context) {
	ticker := time.NewTicker(s.PingFreq)
	defer s.wg.Done()

	for {
		select {
		case <-ticker.C:
			s.ping()
			s.pingChan <- "PONG response to PING"

		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

//close will closed the server connection and close the active go routines
func (s *Server) close() {
	if s.running {
		s.running = false
		s.conn.SetReadDeadline(time.Now())
	}

	s.wg.Wait()
	close(s.pingChan)
	close(s.recvChan)
	close(s.errChan)
	close(s.closeChan)
}
