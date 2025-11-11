package ports

import (
	"context"
)

type Player struct {
	ID              int
	Name            string
	Assets          map[string]int // Map of asset ID to phase ID when selected
	CompletedPhases []int          // Array of phase IDs that the player has completed
}

// Store defines the interface for game data storage operations
type Store interface {
	// RegisterPlayer registers a new player with the given name and secret, returning their numeric ID
	RegisterPlayer(ctx context.Context, playerName, secret string) (int, error)

	// GetPlayersCount returns the current number of registered players
	GetPlayersCount(ctx context.Context) (int, error)

	// VerifyPlayerSecret checks if the provided secret matches the stored secret for the player
	VerifyPlayerSecret(ctx context.Context, playerID int, secret string) (bool, error)

	// SelectMultipleAssets records multiple asset selections for a player in a phase
	SelectMultipleAssets(ctx context.Context, playerID int, assetIDs []string, phaseID int) error

	// GetPlayerAssets gets all assets selected by a player
	GetPlayerAssets(ctx context.Context, playerID int) (map[string]int, error)

	// MarkPhaseCompletion records that a player has completed a phase
	MarkPhaseCompletion(ctx context.Context, phaseID int, playerID int) error

	// GetPhaseCompletion gets number of players having completed a phase
	GetPhaseCompletion(ctx context.Context, phaseID int) (int, error)

	// GetPlayerPhaseCompletion returns an array of phase IDs that a player has completed
	GetPlayerPhaseCompletion(ctx context.Context, playerID int) ([]int, error)

	// GetAllGameData collects all asset selections across all players
	GetAllGameData(ctx context.Context) ([]Player, error)

	// SetPhase sets the game phase counter and returns the new current value
	SetPhase(ctx context.Context, phaseID int) (int, error)

	// GetCurrentPhase returns the current value of the game phase counter
	GetCurrentPhase(ctx context.Context) (int, error)
}
