package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nexes/goirc"
)

func userInput(client *goirc.Client, chn *goirc.Channel) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		cmd := strings.SplitN(scanner.Text(), " ", 2)

		if cmd[0] == "/write" {
			chn.SendMessage(cmd[1])

		} else if cmd[0] == "/quit" {
			client.LeaveChannel(chn, "Adios mis amigos")
			client.CloseConnection("Brought to you by goIRC")
		}
	}
}

func main() {
	client := &goirc.Client{
		Server: "irc.freenode.net",
		Port:   6667,
		SSL:    false,
		Nick:   "conJinx_",
	}

	err := client.ConnectToServer()
	if err != nil {
		fmt.Printf("Error connecting %s\n", err.Error())
		return
	}

	client.Listen()

	chn, err := client.JoinChannel("#programming")
	if err != nil {
		fmt.Printf("error connecting to a channel %s\n", err.Error())
	}

	//listen to input from the CLI
	go userInput(client, chn)

	for client.IsOpen() {
		select {
		case fromServer := <-client.RecvServerMessage():
			//get the actual server host name, e.g may not be irc.freenode.net
			if strings.EqualFold(fromServer["IDName"], "rpl_welcome") {
				client.Server = fromServer["Host"]
				fmt.Printf("Server Name %s\n", client.Server)
			}
			//check if we get a ping, send a pong response
			if strings.EqualFold(fromServer["IDName"], "ping") {
				client.SendPongResponse()
			}

		case fromChannel := <-client.RecvChannelMessage():
			//check for channel name change, e.g an extrea leading "#"
			if strings.EqualFold(fromChannel["IDName"], "rpl_channelname") {
				chn.Name = fromChannel["NewName"]
				fmt.Printf("Joined channel %s\n", chn.Name)
			}
			fmt.Printf("channel %s \n", fromChannel["IDName"])

		case fromMessage := <-client.RecvPrivMessage():
			if strings.EqualFold(fromMessage["IDName"], "rpl_privmsg") {
				fmt.Printf(goirc.FormatPrivMsg(fromMessage["Channel"], fromMessage["Nick"], fromMessage["MSG"]))

			} else {
				fmt.Println(fromMessage["IDName"])
			}
		}
	}

	fmt.Println("Closing goIRC")
}
