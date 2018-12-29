package irc

import (
	"strconv"
	"strings"
	"time"
)

const (
	RPL_WELCOME       = 1
	RPL_YOURHOST      = 2
	RPL_CREATED       = 3
	RPL_MYINFO        = 4
	RPL_BOUNCE        = 5
	RPL_LUSERCLIENT   = 251
	RPL_LUSEROP       = 252
	RPL_LUSERUNKNOWN  = 253
	RPL_LUSERCHANNELS = 254
	RPL_LUSERME       = 255
	RPL_LIST          = 322
	RPL_LISTEND       = 323
	RPL_TOPIC         = 332
	RPL_NAMREPLY      = 353
	RPL_ENDOFNAMES    = 366
	RPL_MOTDSTART     = 375
	RPL_MOTD          = 372
	RPL_ENDOFMOTD     = 376
	RPL_FORWARDJOIN   = 470

	RPL_ERRORJOIN = 477

	RPL_ROOMJOIN = 999
	RPL_ROOMPART = 998
	RPL_ROOMQUIT = 997
	RPL_PRIVMSG  = 996
)

type IncomingData struct {
	Code       int32
	CodeName   string
	ServerName string
	Room       string
	Count      int
	Nick       string
	Message    string
	Time       time.Time
}

// parse the input from the server to determine the incoming message type.
// Return false if this couldn't be done
func parseRawInput(line string) (IncomingData, bool) {
	if len(line) == 0 {
		return IncomingData{}, false
	}

	segments := strings.Split(strings.TrimSpace(line), " ")
	responseCode, err := strconv.Atoi(segments[1])

	if err != nil {
		return parseNonNumericReply(segments)
	} else {
		return parseNumericReply(responseCode, segments)
	}
}

func parseNonNumericReply(segments []string) (IncomingData, bool) {
	data := IncomingData{}
	data.Time = time.Now()

	if index := strings.Index(segments[0], "!"); index != -1 {
		data.Nick = segments[0][1:index]
	}

	switch strings.ToLower(segments[1]) {
	case "join":
		data.Code = RPL_ROOMJOIN
		data.CodeName = "RPL_ROOMJOIN"
		data.Room = segments[2]
	case "part":
		data.Code = RPL_ROOMPART
		data.CodeName = "RPL_ROOMPART"
		data.Room = segments[2]
	case "quit":
		data.Code = RPL_ROOMQUIT
		data.CodeName = "RPL_ROOMQUIT"
		data.Message = strings.Join(segments[2:], " ")
	case "privmsg":
		data.Code = RPL_PRIVMSG
		data.CodeName = "RPL_PRIVMSG"
		data.Room = segments[2]
		data.Message = strings.Join(segments[3:], " ")
	}

	return data, true
}

func parseNumericReply(responseCode int, segments []string) (IncomingData, bool) {
	data := IncomingData{}

	data.ServerName = segments[0][1:]
	data.Time = time.Now()
	data.Nick = segments[2]
	data.Message = strings.Join(segments[3:], " ")

	if data.Message[0] == ':' {
		data.Message = data.Message[1:]
	}

	switch responseCode {
	case RPL_WELCOME:
		data.Code = RPL_WELCOME
		data.CodeName = "RPL_WELCOME"
	case RPL_YOURHOST:
		data.Code = RPL_YOURHOST
		data.CodeName = "RPL_YOURHOST"
	case RPL_CREATED:
		data.Code = RPL_CREATED
		data.CodeName = "RPL_CREATED"
	case RPL_MYINFO:
		data.Code = RPL_MYINFO
		data.CodeName = "RPL_MYINFO"
	case RPL_BOUNCE:
		data.Code = RPL_BOUNCE
		data.CodeName = "RPL_BOUNCE"
	case RPL_LUSERCLIENT:
		data.Code = RPL_LUSERCLIENT
		data.CodeName = "RPL_LUSERCLIENT"
	case RPL_LUSEROP:
		data.Code = RPL_LUSEROP
		data.CodeName = "RPL_LUSEROP"
	case RPL_LUSERUNKNOWN:
		data.Code = RPL_LUSERUNKNOWN
		data.CodeName = "RPL_LUSERUNKNOWN"
	case RPL_LUSERCHANNELS:
		data.Code = RPL_LUSERCHANNELS
		data.CodeName = "RPL_LUSERCHANNELS"
	case RPL_LUSERME:
		data.Code = RPL_LUSERME
		data.CodeName = "RPL_LUSERME"
	case RPL_TOPIC:
		data.Code = RPL_TOPIC
		data.CodeName = "RPL_TOPIC"
		data.Room = segments[3]
		data.Message = strings.Join(segments[4:], " ")
	case RPL_NAMREPLY:
		data.Code = RPL_NAMREPLY
		data.CodeName = "RPL_NAMREPLY"
		data.Room = segments[4]
		data.Message = strings.Join(segments[5:], ",")
	case RPL_ENDOFNAMES:
		data.Code = RPL_ENDOFNAMES
		data.CodeName = "RPL_ENDOFNAMES"
		data.Room = segments[3]
		data.Message = strings.Join(segments[4:], " ")
	case RPL_MOTDSTART:
		data.Code = RPL_MOTDSTART
		data.CodeName = "RPL_MOTDSTART"
		data.Message = strings.Join(segments[4:], " ")
	case RPL_MOTD:
		data.Code = RPL_MOTD
		data.CodeName = "RPL_MOTD"
		data.Message = strings.Join(segments[4:], " ")
	case RPL_ENDOFMOTD:
		data.Code = RPL_ENDOFMOTD
		data.CodeName = "RPL_ENDOFMOTD"
	case RPL_FORWARDJOIN:
		data.Code = RPL_FORWARDJOIN
		data.CodeName = "RPL_FORWARDJOIN"
		data.Message = strings.Join(segments[5:], " ")
		data.Room = segments[4]
	case RPL_ERRORJOIN:
		data.Code = RPL_ERRORJOIN
		data.CodeName = "RPL_ERRORJOIN"
		data.Room = segments[3]
		data.Message = strings.Join(segments[4:], " ")
	case RPL_LIST:
		data.Code = RPL_LIST
		data.CodeName = "RPL_LIST"
		data.Room = segments[3]
		data.Message = strings.Join(segments[5:], " ")
		data.Count = 0
		if count, err := strconv.Atoi(segments[4]); err == nil {
			data.Count = count
		}
	case RPL_LISTEND:
		data.Code = RPL_LISTEND
		data.CodeName = "RPL_LISTEND"
		data.Message = strings.Join(segments[3:], " ")
	}

	return data, true
}
