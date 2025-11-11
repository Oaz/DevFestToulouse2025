package main

import (
	"game"
	"log"
	"net/http"
	"time"
)

type Master struct {
	AdminPassword string
	Players       []*Player
	LogChan       chan string
	StatusChan    chan bool
	Phase         int
	Working       int
	Rules         game.RulesData
	Client        *http.Client
	WebSocket     *WebSocketClient
}

func (master *Master) Run(done chan bool) {
	master.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
	master.Phase = 0

	go func() {
		for msg := range master.LogChan {
			log.Printf(msg)
		}
	}()

	go func() {
		for range master.StatusChan {
			master.Working--
			c, t := master.GetCompletion(master.Phase)
			if master.Working > 0 {
				continue
			}
			if master.Phase == len(master.Rules.Phases) {
				done <- true
				return
			}
			c, t = master.GetCompletion(master.Phase + 1)
			log.Printf("Master: %d/%d players completed phase %d which has not started yet", c, t, master.Phase+1)
			master.NextPhase()
		}
	}()

	defer master.WebSocket.Close()
	master.NextPhase()
	for {
		master.WebSocket.handleWebsocket(master.Log, master.HandlePhaseChange)
	}
}

func (master *Master) HandlePhaseChange(phaseId int) {
	master.Phase = phaseId
}

func (master *Master) NextPhase() {
	master.Working = len(master.Players)
	if master.AdminPassword == "" {
		return
	}
	var query struct {
		AdminPassword string `json:"admin_password"`
		PhaseID       int    `json:"phase_id"`
	}
	query.AdminPassword = master.AdminPassword
	query.PhaseID = master.Phase + 1
	log.Printf("Master pushing to phase %d", query.PhaseID)
	var result struct {
		Message  string `json:"message"`
		NewPhase int    `json:"phase_id"`
	}
	err := Post(master.Client, "/forward", query, &result)
	if err != nil {
		log.Printf("Master error: %v", err)
	}
}

func (master *Master) GetCompletion(phaseID int) (int, int) {
	var query struct {
		PhaseID int `json:"phase_id"`
	}
	query.PhaseID = phaseID
	var result struct {
		CompletedCount int `json:"completed_count"`
		TotalPlayers   int `json:"total_players"`
	}
	err := Post(master.Client, "/completion", query, &result)
	if err != nil {
		log.Printf("Master error: %v", err)
		return 0, 0
	}
	return result.CompletedCount, result.TotalPlayers
}

func (master *Master) Log(text string) {
	master.LogChan <- text
}
