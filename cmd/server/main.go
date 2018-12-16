package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nexes/goIRC/pkg/irc"
)

func main() {
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
		fmt.Printf("Error event: %v\n", event.Err)
	})

	client.HandleEventFunc(irc.EventMOTD, func(event irc.EventType) {
		fmt.Printf("%s", event.Message)
	})

	client.HandleEventFunc(irc.EventMessage, func(event irc.EventType) {
		fmt.Printf("[%s]: %s - %s", event.Room, event.Nick, event.Message)
	})

	client.HandleEventFunc(irc.EventRoomMessage, func(event irc.EventType) {
		switch event.Code {
		case irc.RPL_ROOMJOIN:
			fmt.Printf("%s Joined %s", event.Nick, event.Room)
		case irc.RPL_ROOMPART:
			fmt.Printf("%s Parted %s", event.Nick, event.Room)
		case irc.RPL_ROOMQUIT:
			fmt.Printf("%s Quit %s", event.Nick, event.Room)
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

			args := strings.Split(input, " ")
			fmt.Printf("args[0] = %s\n", args[0])

			switch strings.ToLower(args[0]) {
			case "/quit":
				client.StopConnection()
				break loop

			case "/join":
				client.JoinRoom(args[1])
			}
		}
	}()

	client.StartConnection()
}
