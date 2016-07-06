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

//FormatPrivMsg will format the data given to better print the privmsg
func FormatPrivMsg(d []byte) string {
	tokens := strings.Split(string(d), " ")
	user := tokens[0][1:strings.Index(tokens[0], "!")]
	channel := tokens[2]
	msg := strings.Join(tokens[3:], " ")

	return "[" + channel + "] " + user + "\t" + msg
}
