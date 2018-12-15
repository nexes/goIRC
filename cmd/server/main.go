package main

import (
	"fmt"
	"strings"

	"github.com/nexes/goIRC/pkg/irc"
)

func main() {
	client := irc.NewClient("hilljoh", "", "irc.freenode.net")

	client.HandleEventFunc(irc.EventConnect, func(event irc.EventType) {
		fmt.Println("connection connect happened")
	})

	client.HandleEventFunc(irc.EventDisconnect, func(event irc.EventType) {
		fmt.Println("disconnect event happened")
	})

	client.HandleEventFunc(irc.EventPing, func(event irc.EventType) {
		fmt.Println("ping event happened ", event.Message)
	})

	client.HandleEventFunc(irc.EventError, func(event irc.EventType) {
		fmt.Printf("Error event happened %v\n", event.Err)
	})

	// capture user input
	go func() {
	loop:
		for {
			var input string

			_, err := fmt.Scanln(&input)
			if err != nil {
				fmt.Println("error reading stdio ", err)
			}

			switch strings.ToLower(input) {
			case "/quit":
				client.StopConnection()
				break loop
			}
		}
	}()

	client.StartConnection()
	fmt.Println("at the end of main")
}
