package main

import (
	"fmt"
	"game"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Player struct {
	Index      int
	ID         int
	Secret     string
	ServerUrl  string
	Client     *http.Client
	IsPaused   bool
	PhaseID    int
	LogChan    chan string
	StatusChan chan bool
	Rules      game.RulesData
	WebSocket  *WebSocketClient
	Available  []string
}

func (player *Player) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	player.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
	player.IsPaused = true
	player.PhaseID = 0
	player.Secret = ""
	defer player.WebSocket.Close()

	for {
		if !player.WebSocket.handleWebsocket(player.Log, player.HandlePhaseChange) {
			continue
		}

		if player.IsPaused {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		if player.PhaseID == 0 {
			continue
		}

		phase := player.Rules.Phases[player.PhaseID-1]
		switch phase.Type {
		case "begin":
			if player.Secret == "" {
				player.Wait()
				player.Register()
			}
			player.CompletePhase()
		case "end":
			player.CompletePhase()
		case "single", "multi":
			player.Wait()
			player.SelectAssets(phase)
			player.CompletePhase()
		}
	}
}

func GetWaitingRange() (int, int) {
	waitRange := strings.Split(os.Getenv("PLAYER_WAIT"), "-")
	if len(waitRange) != 2 {
		return 0, 0
	}
	minWait, err1 := strconv.Atoi(waitRange[0])
	maxWait, err2 := strconv.Atoi(waitRange[1])
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return minWait, maxWait
}

var minWait, maxWait = GetWaitingRange()

func (player *Player) Wait() {
	wait := RandomInt(minWait, maxWait)
	time.Sleep(time.Duration(wait) * time.Millisecond)
}

func (player *Player) Log(text string) {
	player.LogChan <- fmt.Sprintf("Player %d: %s", player.Index, text)
}

func (player *Player) LogError(err error) {
	player.Log(fmt.Sprintf("error: %v", err))
}

func (player *Player) HandlePhaseChange(phaseID int) {
	player.PhaseID = phaseID
	player.IsPaused = false
}

func (player *Player) CompletePhase() {
	player.IsPaused = true
	player.StatusChan <- true
	player.Log(fmt.Sprintf("completed phase %d", player.PhaseID))
}

func (player *Player) Register() {
	var query struct {
		Username string `json:"username"`
	}
	query.Username = fmt.Sprintf("player-%d", player.Index)
	var result struct {
		PlayerID int    `json:"player_id"`
		Secret   string `json:"secret"`
	}
	err := Post(player.Client, "/register", query, &result)
	if err != nil {
		player.LogError(err)
		return
	}
	player.ID = result.PlayerID
	player.Secret = result.Secret
	player.Log(fmt.Sprintf(
		"registration complete with ID %d and secret %s",
		player.ID, player.Secret,
	))
}

func (player *Player) SelectAssets(phase game.PhaseData) {
	var query struct {
		PlayerID int      `json:"player_id"`
		Secret   string   `json:"secret"`
		PhaseID  int      `json:"phase_id"`
		AssetIDs []string `json:"asset_ids"`
	}
	query.PlayerID = player.ID
	query.Secret = player.Secret
	query.PhaseID = player.PhaseID
	if player.Available == nil {
		player.Available = make([]string, 0)
	}
	for _, asset := range phase.Available {
		player.Available = append(player.Available, asset.ID)
	}
	player.Log(fmt.Sprintf("can choose among: %v", player.Available))
	candidates := ListRandomSubsets(player.Available)
	for _, candidate := range candidates {
		player.Log(fmt.Sprintf("trying: %v", candidate))
		query.AssetIDs = candidate
		var result struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		err := Post(player.Client, "/select", query, &result)
		if err == nil {
			player.Available = ListDifference(player.Available, candidate)
			break
		}
		player.LogError(err)
	}

}
