// gamestore.go
package gamestore

import (
	"context"
	"fmt"
	"fulcrum/ports"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// GameStore handles Redis operations for the game
type GameStore struct {
	client *redis.Client
}

// NewGameStore creates a new game store with Redis
func NewGameStore(ctx context.Context, redisAddr string) (*GameStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Test connection using the provided context
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &GameStore{client: client}, nil
}

// RegisterPlayer registers a new player with the given name and secret, returning their numeric ID
func (gs *GameStore) RegisterPlayer(ctx context.Context, playerName, secret string) (int, error) {
	// Atomically increment and get the player counter
	playerID, err := gs.client.Incr(ctx, "PLAYER:counter").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to generate player ID: %w", err)
	}

	// Store the player name and secret
	key := fmt.Sprintf("PLAYER:%d:INFO", playerID)
	_, err = gs.client.HSet(ctx, key,
		"name", playerName,
		"secret", secret).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to store player info: %w", err)
	}

	return int(playerID), nil
}

// GetPlayersCount returns the current number of registered players
func (gs *GameStore) GetPlayersCount(ctx context.Context) (int, error) {
	count, err := gs.client.Get(ctx, "PLAYER:counter").Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get players count: %w", err)
	}
	return count, nil
}

// VerifyPlayerSecret checks if the provided secret matches the stored secret for the player
func (gs *GameStore) VerifyPlayerSecret(ctx context.Context, playerID int, secret string) (bool, error) {
	key := fmt.Sprintf("PLAYER:%d:INFO", playerID)
	storedSecret, err := gs.client.HGet(ctx, key, "secret").Result()
	if err != nil {
		if err == redis.Nil {
			return false, fmt.Errorf("player not found: %d", playerID)
		}
		return false, fmt.Errorf("failed to get player secret: %w", err)
	}

	return storedSecret == secret, nil
}

// GetPlayerName retrieves a player's name by ID
func (gs *GameStore) GetPlayerName(ctx context.Context, playerID int) (string, error) {
	key := fmt.Sprintf("PLAYER:%d:INFO", playerID)
	name, err := gs.client.HGet(ctx, key, "name").Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("player not found: %d", playerID)
		}
		return "", fmt.Errorf("failed to get player name: %w", err)
	}

	return name, nil
}

// SelectAsset records that a player has selected an asset in a specific phase
func (gs *GameStore) SelectAsset(ctx context.Context, playerID int, assetID string, phaseID int) error {
	key := fmt.Sprintf("PLAYER:%d:ASSETS", playerID)
	_, err := gs.client.HSet(ctx, key, assetID, phaseID).Result()
	if err != nil {
		return fmt.Errorf("failed to select asset: %w", err)
	}
	return nil
}

// SelectMultipleAssets records multiple asset selections for a player in a phase
func (gs *GameStore) SelectMultipleAssets(ctx context.Context, playerID int, assetIDs []string, phaseID int) error {
	key := fmt.Sprintf("PLAYER:%d:ASSETS", playerID)

	pipe := gs.client.Pipeline()
	for _, assetID := range assetIDs {
		pipe.HSet(ctx, key, assetID, phaseID)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to select multiple assets: %w", err)
	}

	return nil
}

// HasSelectedAsset checks if a player has selected a specific asset
func (gs *GameStore) HasSelectedAsset(ctx context.Context, playerID int, assetID string) (bool, error) {
	key := fmt.Sprintf("PLAYER:%d:ASSETS", playerID)
	exists, err := gs.client.HExists(ctx, key, assetID).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check asset selection: %w", err)
	}
	return exists, nil
}

// GetPlayerAssets gets all assets selected by a player
func (gs *GameStore) GetPlayerAssets(ctx context.Context, playerID int) (map[string]int, error) {
	key := fmt.Sprintf("PLAYER:%d:ASSETS", playerID)
	assetsWithPhase, err := gs.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get player assets: %w", err)
	}

	// Convert phase IDs from string to int
	result := make(map[string]int, len(assetsWithPhase))
	for asset, phaseStr := range assetsWithPhase {
		phaseID, err := strconv.Atoi(phaseStr)
		if err != nil {
			return nil, fmt.Errorf("invalid phase ID format: %w", err)
		}
		result[asset] = phaseID
	}

	return result, nil
}

// MarkPhaseCompletion marks that a player has completed a specific phase
func (gs *GameStore) MarkPhaseCompletion(ctx context.Context, phaseID int, playerID int) error {
	key := fmt.Sprintf("PHASE:%d", phaseID)
	_, err := gs.client.SetBit(ctx, key, int64(playerID), 1).Result()
	if err != nil {
		return fmt.Errorf("failed to mark phase completion: %w", err)
	}
	playerPhasesKey := fmt.Sprintf("PLAYER:%d:PHASES", playerID)
	_, err = gs.client.SetBit(ctx, playerPhasesKey, int64(phaseID), 1).Result()
	if err != nil {
		return fmt.Errorf("failed to mark player's phase completion: %w", err)
	}
	return nil
}

// GetPhaseCompletion returns the count of players who have completed a specific phase
func (gs *GameStore) GetPhaseCompletion(ctx context.Context, phaseID int) (int, error) {
	key := fmt.Sprintf("PHASE:%d", phaseID)
	count, err := gs.client.BitCount(ctx, key, nil).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get phase completion count: %w", err)
	}
	return int(count), nil
}

// GetPlayerPhaseCompletion returns an array of phase IDs that a player has completed
func (gs *GameStore) GetPlayerPhaseCompletion(ctx context.Context, playerID int) ([]int, error) {
	key := fmt.Sprintf("PLAYER:%d:PHASES", playerID)

	// Get the bitmap and convert it to a list of completed phase IDs
	bitfield, err := gs.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// If the key doesn't exist, return an empty array (no phases completed)
			return []int{}, nil
		}
		return nil, fmt.Errorf("failed to get player's phase completion: %w", err)
	}

	// Convert the bitmap to a list of completed phase IDs
	completedPhases := []int{}
	for i := 0; i < len(bitfield)*8; i++ {
		// Get the bit value at position i by calculating the byte offset and bit position
		byteIndex := i / 8
		if byteIndex >= len(bitfield) {
			break
		}

		bitPos := uint(7 - (i % 8)) // Bits are stored in big-endian format
		bitVal := (bitfield[byteIndex] >> bitPos) & 1

		if bitVal == 1 {
			completedPhases = append(completedPhases, i)
		}
	}

	return completedPhases, nil
}

// GetAssetPhase gets the phase when a player selected a specific asset
func (gs *GameStore) GetAssetPhase(ctx context.Context, playerID int, assetID string) (int, error) {
	key := fmt.Sprintf("PLAYER:%d:ASSETS", playerID)
	phaseStr, err := gs.client.HGet(ctx, key, assetID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("asset not selected by player: %s", assetID)
		}
		return 0, fmt.Errorf("failed to get asset phase: %w", err)
	}

	phaseID, err := strconv.Atoi(phaseStr)
	if err != nil {
		return 0, fmt.Errorf("invalid phase ID format: %w", err)
	}

	return phaseID, nil
}

// GetAllPlayerIDs gets all registered player IDs
func (gs *GameStore) GetAllPlayerIDs(ctx context.Context) ([]int, error) {
	// Get all keys matching the pattern "PLAYER:<id>:INFO"
	keys, err := gs.client.Keys(ctx, "PLAYER:*:INFO").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get player keys: %w", err)
	}

	ids := make([]int, 0, len(keys))
	for _, key := range keys {
		// Extract the numeric ID from the key name
		// Format is "PLAYER:<id>:INFO"
		// We need to extract just the <id> part
		idStr := key[7 : len(key)-5] // Remove "PLAYER:" prefix and ":INFO" suffix
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue // Skip invalid formats
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetAllGameData collects all asset selections across all players
func (gs *GameStore) GetAllGameData(ctx context.Context) ([]ports.Player, error) {
	// Get all player IDs
	playerIDs, err := gs.GetAllPlayerIDs(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]ports.Player, 0, len(playerIDs))

	// For each player, get their name, asset selections, and phase completions
	for _, playerID := range playerIDs {
		name, err := gs.GetPlayerName(ctx, playerID)
		if err != nil {
			return nil, err
		}

		assets, err := gs.GetPlayerAssets(ctx, playerID)
		if err != nil {
			return nil, err
		}

		completedPhases, err := gs.GetPlayerPhaseCompletion(ctx, playerID)
		if err != nil {
			return nil, err
		}

		player := ports.Player{
			ID:              playerID,
			Name:            name,
			Assets:          assets,
			CompletedPhases: completedPhases,
		}

		results = append(results, player)
	}

	return results, nil
}

// ClearPlayerData removes all data for a player
func (gs *GameStore) ClearPlayerData(ctx context.Context, playerID int) error {
	// Keys to delete
	infoKey := fmt.Sprintf("PLAYER:%d:INFO", playerID)
	assetsKey := fmt.Sprintf("PLAYER:%d:ASSETS", playerID)

	_, err := gs.client.Del(ctx, infoKey, assetsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to clear player data: %w", err)
	}
	return nil
}

// ResetGame clears all game data (for testing or game reset)
func (gs *GameStore) ResetGame(ctx context.Context) error {
	// Get all player keys
	infoKeys, err := gs.client.Keys(ctx, "PLAYER:*:INFO").Result()
	if err != nil {
		return fmt.Errorf("failed to get player info keys: %w", err)
	}

	assetKeys, err := gs.client.Keys(ctx, "PLAYER:*:ASSETS").Result()
	if err != nil {
		return fmt.Errorf("failed to get player asset keys: %w", err)
	}

	playerPhasesKeys, err := gs.client.Keys(ctx, "PLAYER:*:PHASES").Result()
	if err != nil {
		return fmt.Errorf("failed to get player phases keys: %w", err)
	}

	phaseKeys, err := gs.client.Keys(ctx, "PHASE:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get phase completion keys: %w", err)
	}

	// Combine all keys to delete
	allKeys := append(infoKeys, assetKeys...)
	allKeys = append(allKeys, playerPhasesKeys...)
	allKeys = append(allKeys, phaseKeys...)
	allKeys = append(allKeys, "PLAYER:counter", "GAME:phase")

	// Delete all keys
	if len(allKeys) > 0 {
		_, err = gs.client.Del(ctx, allKeys...).Result()
		if err != nil {
			return fmt.Errorf("failed to delete game data: %w", err)
		}
	}

	// Reset player counter
	_, err = gs.client.Set(ctx, "PLAYER:counter", 0, 0).Result()
	if err != nil {
		return fmt.Errorf("failed to reset player counter: %w", err)
	}

	return nil
}

// SetPhase sets the game phase counter and returns the new current value
func (gs *GameStore) SetPhase(ctx context.Context, phaseID int) (int, error) {
	_, err := gs.client.Set(ctx, "GAME:phase", phaseID, 0).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment game phase: %w", err)
	}

	return phaseID, nil
}

// GetCurrentPhase returns the current value of the game phase counter
func (gs *GameStore) GetCurrentPhase(ctx context.Context) (int, error) {
	// Get the current phase value
	phaseStr, err := gs.client.Get(ctx, "GAME:phase").Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist yet, which means the game is in phase 0
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current game phase: %w", err)
	}

	// Convert the phase from string to int
	phase, err := strconv.Atoi(phaseStr)
	if err != nil {
		return 0, fmt.Errorf("invalid phase format: %w", err)
	}

	return phase, nil
}

// Ensure GameStore implements the Store interface
var _ ports.Store = (*GameStore)(nil)
