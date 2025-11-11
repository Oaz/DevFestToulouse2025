package main

import (
	"fmt"
	"game"
	"os"
	"strconv"
	"sync"
)

func main() {
	fmt.Println("Starting...")

	numPlayers := 500
	rules := game.ReadRules()
	logChan := make(chan string, numPlayers*100)
	statusChan := make(chan bool, numPlayers*100)

	master := &Master{
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		Players:       make([]*Player, numPlayers),
		Rules:         rules,
		LogChan:       logChan,
		StatusChan:    statusChan,
		WebSocket:     CreateWebSocket("master"),
	}

	var wg sync.WaitGroup
	for i := 0; i < numPlayers; i++ {
		ws := CreateWebSocket(strconv.Itoa(i))

		player := &Player{
			Index:      i,
			Rules:      rules,
			LogChan:    logChan,
			StatusChan: statusChan,
			WebSocket:  ws,
		}

		master.Players[i] = player

		wg.Add(1)
		go player.Run(&wg)
	}
	fmt.Println("All players started.")

	fmt.Println("Starting master.")
	done := make(chan bool)
	go master.Run(done)
	<-done
	fmt.Println("This is the end.")
}
