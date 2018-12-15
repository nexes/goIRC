package irc

import (
	"fmt"
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
	RPL_MOTDSTART     = 375
	RPL_MOTD          = 372
	RPL_ENDOFMOTD     = 376
)

type IncomingData struct {
	Code       int32
	CodeName   string
	ServerName string
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

	data := IncomingData{
		Time: time.Now(),
	}
	segments := strings.Split(line, " ")

	if responseCode, err := strconv.Atoi(segments[1]); err == nil {
		data.ServerName = segments[0][1:]
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
		default:
			fmt.Println("default switch: ", line)
		}
	} else {
		// TODO
		fmt.Println("NON NUMERICAL: ", line)
	}

	return data, true
}