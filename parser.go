package goirc

import (
	"strings"
)

//IsPingRequest will return true if a PONG response is needed DELETE
//msg: send the data to see if its a ping
//server: the server you are connected to that the ping request would come from
func IsPingRequest(msg, server string) bool {
	return strings.Contains(msg, "PING : "+server)
}

//EndOfMOTD will return true if the Message Of The Day is finished DELETE
func EndOfMOTD(d string) bool {
	return strings.Contains(string(d), "End of /MOTD")
}

//FormatPrivMsg will format the data given to better print the privmsg
func FormatPrivMsg(d string) string {
	tokens := strings.Split(d, " ")
	user := tokens[0][:strings.Index(tokens[0], "!")]
	channel := tokens[2]
	msg := strings.Join(tokens[3:], " ")

	return channel + "| " + user + " > \t" + msg[1:]
}
