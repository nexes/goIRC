package goirc

import (
	"strings"
)

//IsPingRequest will return true if a PONG response is needed
func IsPingRequest(d []byte) bool {
	return strings.Contains(string(d), "PING :")
}

//EndOfMOTD will return true if the Message Of The Day is finished
func EndOfMOTD(d []byte) bool {
	return strings.Contains(string(d), "End of /MOTD")
}
