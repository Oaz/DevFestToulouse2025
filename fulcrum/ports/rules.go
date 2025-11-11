package ports

import "context"

type PlayerResult struct {
	ID     int
	Name   string
	Assets map[string]int
	Score  int
}

type Rules interface {
	EnsureRegistrationIsAllowed(store Store, ctxt context.Context) (int, error)
	GetPlayerStatus(phaseID int, playerId int, store Store, ctxt context.Context) ([]string, []int, error)
	GetAllPlayersStatuses(phaseID int, store Store, ctxt context.Context) ([]PlayerResult, error)
	AssetSelection(playerID int, phaseID int, assetsIDs []string, store Store, ctxt context.Context) error
}
