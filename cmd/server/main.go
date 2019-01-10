package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nexes/goIRC/pkg/irc"
)

func main() {
	var currentRoom string
	client := irc.NewClient("hilljoh", "", "irc.freenode.net")

	client.HandleEventFunc(irc.EventConnect, func(event irc.EventType) {
		fmt.Printf("connection = codeName: %d message: %s\n", event.Code, event.Message)
	})

	client.HandleEventFunc(irc.EventDisconnect, func(event irc.EventType) {
		fmt.Println("disconnect event")
	})

	client.HandleEventFunc(irc.EventPing, func(event irc.EventType) {
		fmt.Println("ping event: ", event.Message)
	})

	client.HandleEventFunc(irc.EventError, func(event irc.EventType) {
		fmt.Printf("Error event code %d: msg: %s\n", event.Code, event.Err.Error())
	})

	client.HandleEventFunc(irc.EventMOTD, func(event irc.EventType) {
		fmt.Printf("%s\n", event.Message)
	})

	client.HandleEventFunc(irc.EventMessage, func(event irc.EventType) {
		fmt.Printf("[%s]: %s - %s\n", event.Room, event.Nick, event.Message)
	})

	client.HandleEventFunc(irc.EventChannelMessage, func(event irc.EventType) {
		switch event.Code {
		case irc.RPL_FORWARDJOIN:
			// sometimes the room will forward to a different named room, e.g #programming -> ##programming
			fmt.Printf("Room forwared to %s. message: %s\n", event.Room, event.Message)
			currentRoom = event.Room
		case irc.RPL_LIST:
			fmt.Printf("List: Room = %s. Topic =  %s\n", event.Room, event.Message)
		case irc.RPL_NAMREPLY:
			fmt.Printf("Name for %s: %s\n", event.Room, event.Message)
		}
	})

	client.HandleEventFunc(irc.EventRoomMessage, func(event irc.EventType) {
		switch event.Code {
		case irc.RPL_ROOMJOIN:
			fmt.Printf("\t%s Joined %s\n", event.Nick, event.Room)
		case irc.RPL_ROOMPART:
			fmt.Printf("\t%s Parted %s\n", event.Nick, event.Room)
		case irc.RPL_ROOMQUIT:
			fmt.Printf("\t%s Quit %s\n", event.Nick, event.Room)
		case irc.RPL_TOPIC:
			fmt.Printf("TOPIC: %s\n\n", event.Message)
		}
	})

	// capture user input
	go func() {
	loop:
		for {
			stdin := bufio.NewReader(os.Stdin)
			input, err := stdin.ReadString('\n')
			if err != nil {
				fmt.Println("error reading stdio ", err)
			}

			args := strings.Split(strings.TrimSpace(input), " ")

			switch strings.ToLower(args[0]) {
			case "/join":
				client.Command(irc.Command{
					Action: "join",
					Args:   args[1],
				})
			case "/list":
				client.Command(irc.Command{
					Action: "list",
					Args:   args[1],
				})
			case "/names":
				client.Command(irc.Command{
					Action: "names",
					Args:   args[1],
				})
			case "/quit":
				client.StopConnection()
				break loop

			default:
				message := strings.Join(args, " ")
				client.WriteToTarget(currentRoom, message)
			}
		}
	}()

	client.StartConnection()
}
