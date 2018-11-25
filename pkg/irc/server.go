package irc

import (
	"bufio"
	"context"
	"io"
	"log"
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

	wg       sync.WaitGroup
	recvChan chan string
	pingChan chan string
}

//NewIRCServer create a new irc server
func NewIRCServer(server string, useTLS bool) Server {
	return Server{
		ServerName: server,
		Port:       6667,
		UseTSL:     useTLS,
		running:    false,
		Timeout:    time.Second * 60,
		PingFreq:   time.Minute * 2,
		recvChan:   make(chan string),
		pingChan:   make(chan string),
	}
}

func (s *Server) start(ctx context.Context, rw *bufio.ReadWriter) {
	if !s.running {
		s.running = true

		s.wg.Add(3)
		go s.recv(ctx, rw)
		go s.send(ctx, rw)
		go s.sendPingResponse(ctx, rw)
	}
}

func (s *Server) recv(ctx context.Context, rw *bufio.ReadWriter) {
	defer s.wg.Done()

	for {
		data, err := rw.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("eof error, closing connection ", err)
				ctx.Done()
				return
			}
			// TODO
			log.Println("recv error ", err)
			break
		}

		s.recvChan <- data
	}
}

func (s *Server) send(ctx context.Context, rw *bufio.ReadWriter) {
	defer s.wg.Done()
}

func (s *Server) sendPingResponse(ctx context.Context, rw *bufio.ReadWriter) {
	ticker := time.NewTicker(s.PingFreq)
	defer s.wg.Done()

	for {
		select {
		case <-ticker.C:
			// TODO
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
}
