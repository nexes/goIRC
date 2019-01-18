# goIRC

 A go implementation of connecting to an IRC server. To run this in the command line do the following.

### Install
```go
go get github.com/nexes/goIRC
```

### Example
```go
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
  client := irc.NewClient("yourNick", "yourPassIfYouHaveOne", "irc.freenode.net")

  // listen to the callbacks you're interested in

  client.HandleEventFunc(irc.EventConnect, func(event irc.EventType) {
    fmt.Printf("connection made message: %s\n", event.Message)
  })

  client.HandleEventFunc(irc.EventDisconnect, func(event irc.EventType) {
    fmt.Println("disconnect event")
  })

  client.HandleEventFunc(irc.EventError, func(event irc.EventType) {
    fmt.Printf("Error code %d: msg: %s\n", event.Code, event.Err.Error())
  })

  client.HandleEventFunc(irc.EventMessage, func(event irc.EventType) {
    // received a message
    fmt.Printf("[%s]: %s - %s\n", event.Room, event.Nick, event.Message)
  })

  client.HandleEventFunc(irc.EventChannelMessage, func(event irc.EventType) {
    switch event.Code {
    case irc.RPL_FORWARDJOIN:
      // sometimes the room will forward to a different named room, e.g #programming -> ##programming
      fmt.Printf("Room forwared to %s. message: %s\n", event.Room, event.Message)
      currentRoom = event.Room
    }
  })

  client.HandleEventFunc(irc.EventRoomMessage, func(event irc.EventType) {
    switch event.Code {
    case irc.RPL_ROOMJOIN:
      fmt.Printf("\t%s Joined %s\n", event.Nick, event.Room)
    case irc.RPL_ROOMQUIT:
      fmt.Printf("\t%s Quit %s\n", event.Nick, event.Room)
    }
  })

  // capture user input, send commands or write to a room
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
          Args:   args[1:],
        })
      case "/part":
        client.Command(irc.Command{
          Action: "part",
          Args:   args[1:],
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

  // start the connection, this block here
  client.StartConnection()
}
```

### Todo
* Better documentation for the api
* ssl/tls
* complete response codes

## LICENSE (MIT)
Copyright (c) 2016-2019 Joe Berria
