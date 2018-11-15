package irc

import (
	"time"
)

//Server describe the irc server
type Server struct {
	ServerName string
	Port       int32
	UseTSL     bool
	Timeout    time.Duration
}
