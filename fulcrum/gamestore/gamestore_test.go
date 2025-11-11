package gamestore

import (
	"context"
	"fulcrum/ports"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupRedis creates a Redis container for testing
func setupRedis(t *testing.T) (string, func()) {
	ctx := context.Background()

	// Create Redis container with proper API usage
	redisContainer, err := redis.Run(ctx, "redis:7-alpine",
		redis.WithLogLevel("notice"), // Use notice log level to see "Ready to accept connections"
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(time.Second*30), // Increased timeout
		),
	)
	require.NoError(t, err, "Failed to start Redis container")

	// Get the connection details
	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err, "Failed to get Redis endpoint")

	// Return the endpoint and a cleanup function
	return endpoint, func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %s", err)
		}
	}
}

func TestGameStore_RegisterPlayer(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Test no player yet
	playerCount, err := store.GetPlayersCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, playerCount)

	// Register a player
	playerID, err := store.RegisterPlayer(ctx, "TestPlayer", "test-secret")
	require.NoError(t, err)
	assert.Equal(t, 1, playerID, "First player should have ID 1")

	// Register another player
	playerID2, err := store.RegisterPlayer(ctx, "TestPlayer2", "test-secret2")
	require.NoError(t, err)
	assert.Equal(t, 2, playerID2, "Second player should have ID 2")

	// Test retrieving player name
	name, err := store.GetPlayerName(ctx, playerID)
	require.NoError(t, err)
	assert.Equal(t, "TestPlayer", name)

	// Test retrieving player count
	newPlayerCount, err := store.GetPlayersCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, newPlayerCount)
}

func TestGameStore_PlayerSecret(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Register a player with a secret
	playerName := "SecretPlayer"
	playerSecret := "very-secret-key"
	playerID, err := store.RegisterPlayer(ctx, playerName, playerSecret)
	require.NoError(t, err)

	// Verify correct secret works
	valid, err := store.VerifyPlayerSecret(ctx, playerID, playerSecret)
	require.NoError(t, err)
	assert.True(t, valid, "Valid secret should be verified")

	// Verify incorrect secret fails
	valid, err = store.VerifyPlayerSecret(ctx, playerID, "wrong-secret")
	require.NoError(t, err)
	assert.False(t, valid, "Invalid secret should not be verified")

	// Verify non-existent player
	_, err = store.VerifyPlayerSecret(ctx, 999, "any-secret")
	assert.Error(t, err, "Non-existent player should cause error")
	assert.Contains(t, err.Error(), "player not found", "Error should indicate player not found")
}

func TestGameStore_AssetOperations(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Register a player
	playerID, err := store.RegisterPlayer(ctx, "TestPlayer", "test-secret")
	require.NoError(t, err)

	// Test selecting an asset
	err = store.SelectAsset(ctx, playerID, "asset1", 1)
	require.NoError(t, err)

	// Test has selected asset
	selected, err := store.HasSelectedAsset(ctx, playerID, "asset1")
	require.NoError(t, err)
	assert.True(t, selected, "Player should have selected asset1")

	selected, err = store.HasSelectedAsset(ctx, playerID, "asset2")
	require.NoError(t, err)
	assert.False(t, selected, "Player should not have selected asset2")

	// Test get asset phase
	phase, err := store.GetAssetPhase(ctx, playerID, "asset1")
	require.NoError(t, err)
	assert.Equal(t, 1, phase, "Asset1 should be selected in phase 1")

	// Test selecting multiple assets
	err = store.SelectMultipleAssets(ctx, playerID, []string{"asset2", "asset3"}, 2)
	require.NoError(t, err)

	// Test get player assets
	assets, err := store.GetPlayerAssets(ctx, playerID)
	require.NoError(t, err)
	assert.Equal(t, map[string]int{
		"asset1": 1,
		"asset2": 2,
		"asset3": 2,
	}, assets)
}

func TestGameStore_GetAllPlayerIDs(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Register multiple players
	id1, err := store.RegisterPlayer(ctx, "Player1", "test-secret1")
	require.NoError(t, err)
	id2, err := store.RegisterPlayer(ctx, "Player2", "test-secret2")
	require.NoError(t, err)
	id3, err := store.RegisterPlayer(ctx, "Player3", "test-secret3")
	require.NoError(t, err)

	// Get all player IDs
	ids, err := store.GetAllPlayerIDs(ctx)
	require.NoError(t, err)

	// Sort is not guaranteed, so we check if all IDs are present
	assert.ElementsMatch(t, []int{id1, id2, id3}, ids)
}

func TestGameStore_ComputeGameResults(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Register players and select assets
	id1, err := store.RegisterPlayer(ctx, "Player1", "test-secret1")
	require.NoError(t, err)
	id2, err := store.RegisterPlayer(ctx, "Player2", "test-secret2")
	require.NoError(t, err)

	err = store.SelectAsset(ctx, id1, "weapon1", 1)
	require.NoError(t, err)
	err = store.SelectAsset(ctx, id1, "armor1", 2)
	require.NoError(t, err)
	err = store.SelectAsset(ctx, id2, "weapon2", 1)
	require.NoError(t, err)

	// Mark some phase completions
	err = store.MarkPhaseCompletion(ctx, 1, id1)
	require.NoError(t, err)
	err = store.MarkPhaseCompletion(ctx, 2, id1)
	require.NoError(t, err)
	err = store.MarkPhaseCompletion(ctx, 1, id2)
	require.NoError(t, err)

	// Compute game results
	results, err := store.GetAllGameData(ctx)
	require.NoError(t, err)

	// Create expected results (order is not guaranteed)
	expected := []ports.Player{
		{
			ID:   id1,
			Name: "Player1",
			Assets: map[string]int{
				"weapon1": 1,
				"armor1":  2,
			},
			CompletedPhases: []int{1, 2},
		},
		{
			ID:   id2,
			Name: "Player2",
			Assets: map[string]int{
				"weapon2": 1,
			},
			CompletedPhases: []int{1},
		},
	}

	// Check results
	assert.Len(t, results, 2)

	// Create a map to easily find players by ID
	resultMap := make(map[int]ports.Player)
	for _, p := range results {
		resultMap[p.ID] = p
	}

	// Verify each expected player
	for _, exp := range expected {
		actual, found := resultMap[exp.ID]
		assert.True(t, found, "Player ID %d not found in results", exp.ID)
		assert.Equal(t, exp.Name, actual.Name)
		assert.Equal(t, exp.Assets, actual.Assets)
		assert.ElementsMatch(t, exp.CompletedPhases, actual.CompletedPhases)
	}
}

func TestGameStore_ClearPlayerData(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Register a player and select assets
	playerID, err := store.RegisterPlayer(ctx, "TestPlayer", "test-secret")
	require.NoError(t, err)
	err = store.SelectAsset(ctx, playerID, "asset1", 1)
	require.NoError(t, err)

	// Clear player data
	err = store.ClearPlayerData(ctx, playerID)
	require.NoError(t, err)

	// Verify player data is cleared
	_, err = store.GetPlayerName(ctx, playerID)
	assert.Error(t, err, "Player name should not exist after clearing")

	assets, err := store.GetPlayerAssets(ctx, playerID)
	require.NoError(t, err)
	assert.Empty(t, assets, "Player should have no assets after clearing")
}

func TestGameStore_ResetGame(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Register players and select assets
	id1, err := store.RegisterPlayer(ctx, "Player1", "test-secret1")
	require.NoError(t, err)
	id2, err := store.RegisterPlayer(ctx, "Player2", "test-secret2")
	require.NoError(t, err)

	err = store.SelectAsset(ctx, id1, "weapon1", 1)
	require.NoError(t, err)
	err = store.SelectAsset(ctx, id2, "weapon2", 1)
	require.NoError(t, err)

	// Reset game
	err = store.ResetGame(ctx)
	require.NoError(t, err)

	// Verify game is reset
	players, err := store.GetAllPlayerIDs(ctx)
	require.NoError(t, err)
	assert.Empty(t, players, "No players should exist after reset")

	// Register a new player after reset - should start from ID 1 again
	newID, err := store.RegisterPlayer(ctx, "NewPlayer", "test-secret")
	require.NoError(t, err)
	assert.Equal(t, 1, newID, "Player ID should start from 1 after reset")
}

func TestGameStore_DuplicatePlayerNames(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Register two players with the same name
	id1, err := store.RegisterPlayer(ctx, "SameName", "test-secret")
	require.NoError(t, err)
	id2, err := store.RegisterPlayer(ctx, "SameName", "test-secret2")
	require.NoError(t, err)

	// IDs should be different
	assert.NotEqual(t, id1, id2, "Players have same name but should have different IDs anyway")

	// Both names should be retrievable
	name1, err := store.GetPlayerName(ctx, id1)
	require.NoError(t, err)
	name2, err := store.GetPlayerName(ctx, id2)
	require.NoError(t, err)

	assert.Equal(t, "SameName", name1)
	assert.Equal(t, "SameName", name2)

	// Select different assets
	err = store.SelectAsset(ctx, id1, "asset1", 1)
	require.NoError(t, err)
	err = store.SelectAsset(ctx, id2, "asset2", 2)
	require.NoError(t, err)

	// Each player should have their own assets
	assets1, err := store.GetPlayerAssets(ctx, id1)
	require.NoError(t, err)
	assets2, err := store.GetPlayerAssets(ctx, id2)
	require.NoError(t, err)

	assert.Equal(t, map[string]int{"asset1": 1}, assets1)
	assert.Equal(t, map[string]int{"asset2": 2}, assets2)
}

func TestGameStore_NonExistentPlayer(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Try to get name of non-existent player
	_, err = store.GetPlayerName(ctx, 999)
	assert.Error(t, err, "Getting name of non-existent player should error")

	// Try to get assets of non-existent player
	assets, err := store.GetPlayerAssets(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, assets, "Non-existent player should have no assets")

	// Try to get asset phase of non-existent player
	_, err = store.GetAssetPhase(ctx, 999, "asset")
	assert.Error(t, err, "Getting asset phase for non-existent player should error")
}

func TestGameStore_GamePhase(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Test initial phase
	phase, err := store.GetCurrentPhase(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, phase, "Initial game phase should be 0")

	// Test NextPhase
	phase, err = store.SetPhase(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, phase, "Game phase should be 1 after first increment")

	// Verify phase using GetCurrentPhase
	phase, err = store.GetCurrentPhase(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, phase, "Current game phase should be 1")

	// Increment again
	phase, err = store.SetPhase(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, phase, "Game phase should be 2 after second increment")

	// Verify phase again
	phase, err = store.GetCurrentPhase(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, phase, "Current game phase should be 2")

	// Test phase reset
	err = store.ResetGame(ctx)
	require.NoError(t, err)

	phase, err = store.GetCurrentPhase(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, phase, "Game phase should be reset to 0")
}

func TestGameStore_PhaseCompletion(t *testing.T) {
	ctx := context.Background()
	redisAddr, cleanup := setupRedis(t)
	defer cleanup()

	store, err := NewGameStore(ctx, redisAddr)
	require.NoError(t, err)

	// Register some test players
	player1, err := store.RegisterPlayer(ctx, "Player1", "secret1")
	require.NoError(t, err)
	player2, err := store.RegisterPlayer(ctx, "Player2", "secret2")
	require.NoError(t, err)
	player3, err := store.RegisterPlayer(ctx, "Player3", "secret3")
	require.NoError(t, err)

	// Test initial phase completion count
	phaseID := 1
	count, err := store.GetPhaseCompletion(ctx, phaseID)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Initial phase completion count should be zero")

	// Test initial player phase completion
	phases, err := store.GetPlayerPhaseCompletion(ctx, player1)
	require.NoError(t, err)
	assert.Empty(t, phases, "Player should have no completed phases initially")

	// Mark player1 as completed phase 1
	err = store.MarkPhaseCompletion(ctx, phaseID, player1)
	require.NoError(t, err)

	// Verify phase completion count after one player completed
	count, err = store.GetPhaseCompletion(ctx, phaseID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Phase completion count should be 1 after one player completes")

	// Verify player1's completed phases
	phases, err = store.GetPlayerPhaseCompletion(ctx, player1)
	require.NoError(t, err)
	assert.Equal(t, []int{1}, phases, "Player1 should have completed phase 1")

	// Mark player2 as completed phase 1
	err = store.MarkPhaseCompletion(ctx, phaseID, player2)
	require.NoError(t, err)

	// Verify phase completion count after two players completed
	count, err = store.GetPhaseCompletion(ctx, phaseID)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "Phase completion count should be 2 after two players complete")

	// Check a different phase that no one has completed
	anotherPhaseID := 2
	count, err = store.GetPhaseCompletion(ctx, anotherPhaseID)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "No players should have completed phase 2")

	// Mark a player as completed the second phase
	err = store.MarkPhaseCompletion(ctx, anotherPhaseID, player3)
	require.NoError(t, err)

	// Verify completion count for each phase
	count1, err := store.GetPhaseCompletion(ctx, phaseID)
	require.NoError(t, err)
	count2, err := store.GetPhaseCompletion(ctx, anotherPhaseID)
	require.NoError(t, err)

	assert.Equal(t, 2, count1, "Two players should have completed phase 1")
	assert.Equal(t, 1, count2, "One player should have completed phase 2")

	// Verify player3's completed phases
	phases, err = store.GetPlayerPhaseCompletion(ctx, player3)
	require.NoError(t, err)
	assert.Equal(t, []int{2}, phases, "Player3 should have completed phase 2")

	// Mark player3 as also completing phase 1
	err = store.MarkPhaseCompletion(ctx, phaseID, player3)
	require.NoError(t, err)

	// Verify player3 has now completed both phases
	phases, err = store.GetPlayerPhaseCompletion(ctx, player3)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 2}, phases, "Player3 should have completed phases 1 and 2")

	// Test marking completion again doesn't change the count
	err = store.MarkPhaseCompletion(ctx, phaseID, player1) // player1 already completed
	require.NoError(t, err)

	count, err = store.GetPhaseCompletion(ctx, phaseID)
	require.NoError(t, err)
	assert.Equal(t, 3, count, "Phase 1 should now have 3 completions")

	// Test ResetGame also clears phase completion data
	err = store.ResetGame(ctx)
	require.NoError(t, err)

	count, err = store.GetPhaseCompletion(ctx, phaseID)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Phase completion count should be reset to 0")

	// Verify player phase completion is also reset
	phases, err = store.GetPlayerPhaseCompletion(ctx, player1)
	require.NoError(t, err)
	assert.Empty(t, phases, "Player phase completion should be reset")
}
