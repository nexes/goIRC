package irc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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
	recvChan chan IncomingData
	errChan  chan error
	pingChan chan string
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

		recvChan: make(chan IncomingData),
		errChan:  make(chan error),
		pingChan: make(chan string),
	}
}

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

func (s *Server) recv(ctx context.Context) {
	defer s.wg.Done()

	for {
		data, err := s.readWriter.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("EOF: closing connection ", err)
				s.errChan <- err
				return
			}

			s.errChan <- err
		}

		if recData, ok := parseRawInput(data); ok {
			s.recvChan <- recData
		}
	}
}

func (s *Server) send(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("send ctx.Done called")
			return
		}
	}
}

func (s *Server) sendPingResponse(ctx context.Context) {
	ticker := time.NewTicker(s.PingFreq)
	defer s.wg.Done()

	for {
		select {
		case <-ticker.C:
			log.Println("sending ping")
			s.pingChan <- "all done with ping"

		case <-ctx.Done():
			log.Println("ping ctx.Done called")
			ticker.Stop()
			return
		}
	}
}

func (s *Server) close() {
	// todo
	close(s.pingChan)
	close(s.recvChan)
	close(s.errChan)
}
