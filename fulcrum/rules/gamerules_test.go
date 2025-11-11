package gamerules

import (
	"context"
	"errors"
	"fmt"
	"fulcrum/ports"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockStore is a mock implementation of the Store interface for testing
type MockStore struct {
	currentPhase            int
	playerAssets            map[int]map[string]int
	playersCount            int
	nextPhaseError          error
	currentPhaseError       error
	playerAssetsError       error
	allGameDataError        error
	selectAssetsCalled      bool
	selectAssetsPlayerID    int
	selectAssetsPhaseID     int
	selectAssetsSelectedIDs []string
	completedPhases         map[int][]int
	completedPlayerPhases   map[int][]int
}

func (m *MockStore) MarkPhaseCompletion(ctx context.Context, phaseID int, playerID int) error {
	if m.completedPhases == nil {
		m.completedPhases = make(map[int][]int)
	}
	m.completedPhases[phaseID] = append(m.completedPhases[phaseID], playerID)
	return nil
}

func (m *MockStore) GetPhaseCompletion(ctx context.Context, phaseID int) (int, error) {
	if m.completedPhases == nil {
		return 0, nil
	}
	return len(m.completedPhases[phaseID]), nil
}

func (m *MockStore) GetPlayerPhaseCompletion(ctx context.Context, playerID int) ([]int, error) {
	if m.completedPlayerPhases == nil {
		return []int{}, nil
	}
	return m.completedPlayerPhases[playerID], nil
}

func (m *MockStore) RegisterPlayer(ctx context.Context, playerName, secret string) (int, error) {
	return 0, nil
}

func (m *MockStore) GetPlayersCount(ctx context.Context) (int, error) {
	return m.playersCount, nil
}

func (m *MockStore) VerifyPlayerSecret(ctx context.Context, playerID int, secret string) (bool, error) {
	return true, nil
}

func (m *MockStore) SelectMultipleAssets(ctx context.Context, playerID int, assetIDs []string, phaseID int) error {
	m.selectAssetsCalled = true
	m.selectAssetsPlayerID = playerID
	m.selectAssetsSelectedIDs = assetIDs
	m.selectAssetsPhaseID = phaseID
	return nil
}

func (m *MockStore) GetPlayerAssets(ctx context.Context, playerID int) (map[string]int, error) {
	if m.playerAssetsError != nil {
		return nil, m.playerAssetsError
	}
	if assets, ok := m.playerAssets[playerID]; ok {
		return assets, nil
	}
	return map[string]int{}, nil
}

func (m *MockStore) GetAllGameData(ctx context.Context) ([]ports.Player, error) {
	if m.allGameDataError != nil {
		return nil, m.allGameDataError
	}
	var players []ports.Player
	for playerID, assets := range m.playerAssets {
		completedPhases := []int{}
		if phases, ok := m.completedPlayerPhases[playerID]; ok {
			completedPhases = phases
		}
		players = append(players, ports.Player{
			ID:              playerID,
			Name:            fmt.Sprintf("Player%d", playerID),
			Assets:          assets,
			CompletedPhases: completedPhases,
		})
	}
	return players, nil
}

func (m *MockStore) SetPhase(ctx context.Context, phaseId int) (int, error) {
	if m.nextPhaseError != nil {
		return 0, m.nextPhaseError
	}
	m.currentPhase = phaseId
	return m.currentPhase, nil
}

func (m *MockStore) GetCurrentPhase(ctx context.Context) (int, error) {
	if m.currentPhaseError != nil {
		return 0, m.currentPhaseError
	}
	return m.currentPhase, nil
}

// Helper function to create a GameRules instance for testing
func createTestGameRules() GameRules {
	asset1 := Asset{ID: "RA", Text: "Test Asset 1", Cost: 3}
	asset2 := Asset{ID: "RB", Text: "Test Asset 2", Cost: 4}
	asset3 := Asset{ID: "RC", Text: "Test Asset 3", Cost: 5}
	asset4 := Asset{ID: "RD", Text: "Test Asset 4", Cost: 13}

	acceptedSet1 := NewSet[string]()
	acceptedSet1.Add(asset1.ID)

	acceptedSet2 := NewSet[string]()
	acceptedSet2.Add(asset2.ID)

	acceptedSet3 := NewSet[string]()
	acceptedSet3.Add(asset2.ID)
	acceptedSet3.Add(asset3.ID)

	acceptedSet4 := NewSet[string]()
	acceptedSet4.Add(asset1.ID)
	acceptedSet4.Add(asset4.ID)

	phases := map[int]Phase{
		1: {ID: 1, Type: "begin", Available: []string{}, Accepted: []Set[string]{}},
		2: {
			ID: 2, Type: "single",
			Available: []string{asset1.ID, asset2.ID},
			Accepted:  []Set[string]{acceptedSet1, acceptedSet2},
		},
		3: {
			ID: 3, Type: "single",
			Available: []string{asset1.ID, asset3.ID},
			Accepted:  []Set[string]{acceptedSet1, acceptedSet3},
		},
		4: {ID: 4, Type: "end", Available: []string{}, Accepted: []Set[string]{}},
		5: {ID: 5, Type: "begin", Available: []string{}, Accepted: []Set[string]{}},
		6: {
			ID: 6, Type: "multi",
			Available: []string{asset1.ID, asset2.ID, asset3.ID},
			Accepted:  []Set[string]{acceptedSet1, acceptedSet3},
		},
		7: {
			ID: 7, Type: "multi",
			Available: []string{asset4.ID},
			Accepted:  []Set[string]{acceptedSet4, acceptedSet3},
		},
		8: {ID: 8, Type: "end", Available: []string{}, Accepted: []Set[string]{}},
	}

	assets := map[string]Asset{
		"RA": asset1,
		"RB": asset2,
		"RC": asset3,
		"RD": asset4,
	}

	return MakeGameRules(assets, phases)
}

// Tests for AssetSelection
func TestAssetSelection_ValidSingleSelection(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
	}
	ctx := context.Background()
	playerID := 1
	phaseID := 2
	assetIDs := []string{"RA"}

	// Act
	err := rules.AssetSelection(playerID, phaseID, assetIDs, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.True(t, mockStore.selectAssetsCalled, "SelectMultipleAssets should have been called")
	assert.Equal(t, playerID, mockStore.selectAssetsPlayerID, "SelectMultipleAssets should be called with correct playerID")
	assert.Equal(t, phaseID, mockStore.selectAssetsPhaseID, "SelectMultipleAssets should be called with correct phaseID")
	assert.Equal(t, assetIDs, mockStore.selectAssetsSelectedIDs, "SelectMultipleAssets should be called with correct assetIDs")
}

func TestAssetSelection_ValidMultiSelection(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 6,
		playerAssets: map[int]map[string]int{
			1: {},
		},
	}
	ctx := context.Background()
	playerID := 1
	phaseID := 6
	assetIDs := []string{"RB", "RC"}

	// Act
	err := rules.AssetSelection(playerID, phaseID, assetIDs, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.True(t, mockStore.selectAssetsCalled, "SelectMultipleAssets should have been called")
	assert.Equal(t, playerID, mockStore.selectAssetsPlayerID, "SelectMultipleAssets should be called with correct playerID")
	assert.Equal(t, phaseID, mockStore.selectAssetsPhaseID, "SelectMultipleAssets should be called with correct phaseID")
	assert.Equal(t, assetIDs, mockStore.selectAssetsSelectedIDs, "SelectMultipleAssets should be called with correct assetIDs")
}

func TestAssetSelection_UnacceptableSelection(t *testing.T) {
	// Arrange
	rules := createTestGameRules()

	mockStore := &MockStore{
		currentPhase: 6,
		playerAssets: map[int]map[string]int{
			1: {},
		},
		completedPhases: make(map[int][]int),
	}
	ctx := context.Background()
	playerID := 1
	phaseID := 6
	assetIDs := []string{"RB"} // Is not accepted alone in phase 6

	// Act
	err := rules.AssetSelection(playerID, phaseID, assetIDs, mockStore, ctx)

	// Assert
	assert.Error(t, err)
	assert.NotContains(t, mockStore.completedPhases, phaseID, "Current phase should not be marked as completed")
	assert.NotContains(t, mockStore.completedPhases, phaseID+1, "Next phase should NOT be marked as completed")
}

func TestAssetSelection_MarksPhaseCompletion(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
		completedPhases: make(map[int][]int),
	}
	ctx := context.Background()
	playerID := 1
	phaseID := 2
	assetIDs := []string{"RA"}

	// Act
	err := rules.AssetSelection(playerID, phaseID, assetIDs, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, mockStore.completedPhases, phaseID, "Current phase should be marked as completed")
	assert.Contains(t, mockStore.completedPhases[phaseID], playerID, "Current phase should be completed for the player")
}

func TestAssetSelection_MarksNextPhaseCompletionWhenAccepted(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
		completedPhases: make(map[int][]int),
	}
	ctx := context.Background()
	playerID := 1
	phaseID := 2
	assetIDs := []string{"RA"} // This is accepted in phase 2 and phase 3

	// Act
	err := rules.AssetSelection(playerID, phaseID, assetIDs, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, mockStore.completedPhases, phaseID, "Current phase should be marked as completed")
	assert.Contains(t, mockStore.completedPhases[phaseID], playerID, "Current phase should be completed for the player")

	// Check that phase 3 is also marked as completed
	assert.Contains(t, mockStore.completedPhases, phaseID+1, "Next phase should be marked as completed")
	assert.Contains(t, mockStore.completedPhases[phaseID+1], playerID, "Next phase should be completed for the player")
}

func TestAssetSelection_DoesNotMarkNextPhaseWhenNotAccepted(t *testing.T) {
	// Arrange
	rules := createTestGameRules()

	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
		completedPhases: make(map[int][]int),
	}
	ctx := context.Background()
	playerID := 1
	phaseID := 2
	assetIDs := []string{"RB"} // Is not enough for phase 3

	// Act
	err := rules.AssetSelection(playerID, phaseID, assetIDs, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, mockStore.completedPhases, phaseID, "Current phase should be marked as completed")
	assert.Contains(t, mockStore.completedPhases[phaseID], playerID, "Current phase should be completed for the player")
	assert.NotContains(t, mockStore.completedPhases, phaseID+1, "Next phase should NOT be marked as completed")
}

func TestAssetSelection_MarksMultiplePhasesWhenAllAccepted(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
		completedPhases: make(map[int][]int),
	}
	ctx := context.Background()
	playerID := 1
	phaseID := 2
	assetIDs := []string{"RA"} // This is accepted in phases 2, 3, and 6

	// Act
	err := rules.AssetSelection(playerID, phaseID, assetIDs, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	// Check phase 2 (current)
	assert.Contains(t, mockStore.completedPhases, phaseID, "Current phase should be marked as completed")
	assert.Contains(t, mockStore.completedPhases[phaseID], playerID, "Current phase should be completed for the player")

	// Check phase 3
	assert.Contains(t, mockStore.completedPhases, phaseID+1, "Next phase (3) should be marked as completed")
	assert.Contains(t, mockStore.completedPhases[phaseID+1], playerID, "Next phase (3) should be completed for the player")

	// Check phase 6
	assert.Contains(t, mockStore.completedPhases, phaseID+4, "Next phase (6) should be marked as completed")
	assert.Contains(t, mockStore.completedPhases[phaseID+4], playerID, "Next phase (6) should be completed for the player")
}

func TestAssetSelection_PhaseMismatch(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
	}
	ctx := context.Background()

	// Act
	err := rules.AssetSelection(1, 1, []string{"RA"}, mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorIncorrectInput)
}

func TestAssetSelection_InvalidPhase(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 10,
		playerAssets: map[int]map[string]int{
			1: {},
		},
	}
	ctx := context.Background()

	// Act
	err := rules.AssetSelection(1, 10, []string{"RA"}, mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorSystemFailure)
}

func TestAssetSelection_GetPlayerAssetsFails(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
		playerAssetsError: errors.New("failed to get player assets"),
	}
	ctx := context.Background()

	// Act
	err := rules.AssetSelection(1, 2, []string{"RA"}, mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorSystemFailure)
}

func TestAssetSelection_InvalidSelection(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
	}
	ctx := context.Background()

	// Act
	err := rules.AssetSelection(1, 2, []string{"RD"}, mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorIncorrectInput)
}

// Tests for EnsureRegistrationIsAllowed
func TestEnsureRegistrationIsAllowed_AllowedInBeginPhase(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 1, // "begin" phase in our test setup
	}
	ctx := context.Background()

	// Act
	phaseId, err := rules.EnsureRegistrationIsAllowed(mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, phaseId)
}

func TestEnsureRegistrationIsAllowed_GameNotStarted(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 0,
	}
	ctx := context.Background()

	// Act
	_, err := rules.EnsureRegistrationIsAllowed(mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorIncorrectInput)
}

func TestEnsureRegistrationIsAllowed_ClosedInNonBeginPhase(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2, // "single" phase in our test setup
	}
	ctx := context.Background()

	// Act
	_, err := rules.EnsureRegistrationIsAllowed(mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorIncorrectInput)
}

func TestEnsureRegistrationIsAllowed_InvalidPhase(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 10, // phase doesn't exist in our test setup
	}
	ctx := context.Background()

	// Act
	_, err := rules.EnsureRegistrationIsAllowed(mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorSystemFailure)
}

func TestEnsureRegistrationIsAllowed_StoreError(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase:      1,
		currentPhaseError: errors.New("failed to get current phase"),
	}
	ctx := context.Background()

	// Act
	_, err := rules.EnsureRegistrationIsAllowed(mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorSystemFailure)
}

// Tests for GetPlayerStatus
func TestGetPlayerStatus_NoRelevantAssets(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 6,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3},
		},
	}
	ctx := context.Background()

	// Act
	assets, scores, err := rules.GetPlayerStatus(6, 1, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, assets)
	assert.Equal(t, 2, len(scores))
	assert.Equal(t, 9, scores[0])
	assert.Equal(t, rules.MaxScore[8], scores[1])
}

func TestGetPlayerStatus_SomeRelevantAssets(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 6,
		playerAssets: map[int]map[string]int{
			1: {"RA": 2, "RB": 6, "RC": 6},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3, 6},
		},
	}
	ctx := context.Background()

	// Act
	assets, scores, err := rules.GetPlayerStatus(6, 1, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, assets, 2)
	assert.Contains(t, assets, "RB")
	assert.Contains(t, assets, "RC")
	assert.NotContains(t, assets, "RA")
	assert.Equal(t, 2, len(scores))
	assert.Equal(t, 3, scores[0])
	assert.Equal(t, 9, scores[1])
}

func TestGetPlayerStatus_AllRelevantAssets(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 4,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3},
		},
	}
	ctx := context.Background()

	// Act
	assets, scores, err := rules.GetPlayerStatus(4, 1, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, assets, 2)
	assert.Contains(t, assets, "RB")
	assert.Contains(t, assets, "RC")
	assert.Equal(t, 1, len(scores))
	assert.Equal(t, 9, scores[0])
}

func TestGetPlayerStatus_Scores(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 8,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3, "RA": 6, "RD": 7},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3, 6, 7},
		},
	}
	ctx := context.Background()

	// Act
	_, scores, err := rules.GetPlayerStatus(8, 1, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, 9, scores[0])
	assert.Equal(t, 16, scores[1])
}

func TestGetPlayerStatus_ScoresWhenPlayerHasMissedPhase(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 8,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RA": 6, "RD": 7},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 6, 7},
		},
	}
	ctx := context.Background()

	// Act
	_, scores, err := rules.GetPlayerStatus(8, 1, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, rules.MaxScore[4], scores[0])
	assert.Equal(t, 16, scores[1])
}

func TestGetPlayerStatus_ScoresWhenPlayerHasMissedAllPhasesOfFirstSequence(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 8,
		playerAssets: map[int]map[string]int{
			1: {"RA": 6, "RD": 7},
		},
		completedPlayerPhases: map[int][]int{
			1: {6, 7},
		},
	}
	ctx := context.Background()

	// Act
	_, scores, err := rules.GetPlayerStatus(8, 1, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, rules.MaxScore[4], scores[0])
	assert.Equal(t, 16, scores[1])
}

func TestGetPlayerStatus_ScoresWhenRemainingSequence(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 4,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3},
		},
	}
	ctx := context.Background()

	// Act
	_, scores, err := rules.GetPlayerStatus(4, 1, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, scores, 1)
	assert.Equal(t, 9, scores[0])
}

func TestGetPlayerStatus_ScoresWhenPlayerHasMissedAllPhasesOfLastSequence(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 8,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3},
		},
	}
	ctx := context.Background()

	// Act
	_, scores, err := rules.GetPlayerStatus(8, 1, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, 9, scores[0])
	assert.Equal(t, rules.MaxScore[8], scores[1])
}

func TestGetPlayerStatus_InvalidPhase(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 10,
		playerAssets: map[int]map[string]int{
			1: {"RA": 1, "RB": 2},
		},
	}
	ctx := context.Background()

	// Act
	assets, scores, err := rules.GetPlayerStatus(10, 1, mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorSystemFailure)
	assert.Nil(t, assets)
	assert.Nil(t, scores)
}

func TestGetPlayerStatus_StoreError(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 2,
		playerAssets: map[int]map[string]int{
			1: {},
		},
		playerAssetsError: errors.New("failed to get player assets"),
	}
	ctx := context.Background()

	// Act
	assets, scores, err := rules.GetPlayerStatus(2, 1, mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorSystemFailure)
	assert.Nil(t, assets)
	assert.Nil(t, scores)
}

// Tests for GetAllPlayersStatuses
func TestGetAllPlayersStatuses_EmptyGame(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase:          6,
		playerAssets:          map[int]map[string]int{},
		completedPlayerPhases: map[int][]int{},
	}
	ctx := context.Background()

	// Act
	players, err := rules.GetAllPlayersStatuses(6, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, players)
}

func TestGetAllPlayersStatuses_SinglePlayer(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 6,
		playerAssets: map[int]map[string]int{
			1: {"RA": 2, "RB": 6, "RC": 6},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3, 6},
		},
	}
	ctx := context.Background()

	// Act
	players, err := rules.GetAllPlayersStatuses(6, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, players, 1)

	player := players[0]
	assert.Equal(t, 1, player.ID)
	assert.Len(t, player.Assets, 2)
	assert.Equal(t, 6, player.Assets["RB"])
	assert.Equal(t, 6, player.Assets["RC"])
	_, hasRA := player.Assets["RA"]
	assert.False(t, hasRA, "RA should not be included as it was selected before begin phase")
	assert.Equal(t, 9, player.Score)
}

func TestGetAllPlayersStatuses_MultiplePlayers(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 8,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3, "RA": 6, "RD": 7},
			2: {"RA": 2, "RB": 6},
			3: {"RC": 3},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3, 6, 7},
			2: {2, 6},
			3: {3},
		},
	}
	ctx := context.Background()

	// Act
	players, err := rules.GetAllPlayersStatuses(8, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, players, 3)

	// Find players by ID for assertions
	playerMap := make(map[int]ports.PlayerResult)
	for _, player := range players {
		playerMap[player.ID] = player
	}

	// Player 1 - has all assets and completed all phases
	player1 := playerMap[1]
	assert.Equal(t, 1, player1.ID)
	assert.Len(t, player1.Assets, 2) // Only assets from after begin phase (5)
	assert.Equal(t, 6, player1.Assets["RA"])
	assert.Equal(t, 7, player1.Assets["RD"])
	_, hasRB := player1.Assets["RB"]
	assert.False(t, hasRB, "RB should not be included as it was selected before begin phase")
	_, hasRC := player1.Assets["RC"]
	assert.False(t, hasRC, "RC should not be included as it was selected before begin phase")
	assert.Equal(t, 16, player1.Score) // RA(3) + RD(13) from phases 6-8

	// Player 2 - has some assets, missed some phases
	player2 := playerMap[2]
	assert.Equal(t, 2, player2.ID)
	assert.Len(t, player2.Assets, 1) // Only RB from phase 6
	assert.Equal(t, 6, player2.Assets["RB"])
	_, hasRA := player2.Assets["RA"]
	assert.False(t, hasRA, "RA should not be included as it was selected before begin phase")
	assert.Equal(t, rules.MaxScore[8], player2.Score) // Missed phase 7

	// Player 3 - minimal assets
	player3 := playerMap[3]
	assert.Equal(t, 3, player3.ID)
	assert.Empty(t, player3.Assets)                   // RC was selected before begin phase (5)
	assert.Equal(t, rules.MaxScore[8], player3.Score) // Missed phases 6 and 7
}

func TestGetAllPlayersStatuses_MultiplePlayersOldPhase(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 8,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3, "RA": 6, "RD": 7},
			2: {"RA": 2, "RB": 6},
			3: {"RC": 3},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3, 6, 7},
			2: {2, 6},
			3: {2, 3},
		},
	}
	ctx := context.Background()

	// Act
	players, err := rules.GetAllPlayersStatuses(4, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, players, 3)

	// Find players by ID for assertions
	playerMap := make(map[int]ports.PlayerResult)
	for _, player := range players {
		playerMap[player.ID] = player
	}

	// Player 1 - has all assets and completed all phases
	player1 := playerMap[1]
	assert.Equal(t, 1, player1.ID)
	assert.Len(t, player1.Assets, 2)
	assert.Equal(t, 2, player1.Assets["RB"])
	assert.Equal(t, 3, player1.Assets["RC"])
	_, hasRA := player1.Assets["RA"]
	assert.False(t, hasRA, "RA should not be included as it was selected after end phase")
	_, hasRD := player1.Assets["RD"]
	assert.False(t, hasRD, "RD should not be included as it was selected after end phase")
	assert.Equal(t, 9, player1.Score) // RB(4) + RC(5) from phases 2-4

	// Player 2 - has some assets, missed some phases
	player2 := playerMap[2]
	assert.Equal(t, 2, player2.ID)
	assert.Len(t, player2.Assets, 1)
	assert.Equal(t, 2, player2.Assets["RA"])
	_, hasRB := player2.Assets["RB"]
	assert.False(t, hasRB, "RB should not be included as it was selected after end phase")
	assert.Equal(t, rules.MaxScore[4], player2.Score) // Missed phase 3

	// Player 3 - minimal assets
	player3 := playerMap[3]
	assert.Equal(t, 3, player3.ID)
	assert.Len(t, player3.Assets, 1)
	assert.Equal(t, 3, player3.Assets["RC"])
	assert.Equal(t, 5, player3.Score) // RC(5) from phase 3
}

func TestGetAllPlayersStatuses_NoRelevantAssets(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 6,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3}, // Assets from before begin phase (5)
			2: {},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3},
			2: {},
		},
	}
	ctx := context.Background()

	// Act
	players, err := rules.GetAllPlayersStatuses(6, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, players, 2)

	// Both players should have empty assets maps but different scores
	playerMap := make(map[int]ports.PlayerResult)
	for _, player := range players {
		playerMap[player.ID] = player
	}

	player1 := playerMap[1]
	assert.Empty(t, player1.Assets)
	assert.Equal(t, rules.MaxScore[8], player1.Score) // Missed phase 6

	player2 := playerMap[2]
	assert.Empty(t, player2.Assets)
	assert.Equal(t, rules.MaxScore[8], player2.Score) // Missed phase 6
}

func TestGetAllPlayersStatuses_InvalidPhase(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 10,
		playerAssets: map[int]map[string]int{
			1: {"RA": 1, "RB": 2},
		},
	}
	ctx := context.Background()

	// Act
	players, err := rules.GetAllPlayersStatuses(10, mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorSystemFailure)
	assert.Nil(t, players)
}

func TestGetAllPlayersStatuses_StoreError(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 6,
		playerAssets: map[int]map[string]int{
			1: {"RA": 2, "RB": 6},
		},
		allGameDataError: errors.New("failed to get all game data"),
	}

	ctx := context.Background()

	// Act
	players, err := rules.GetAllPlayersStatuses(6, mockStore, ctx)

	// Assert
	assert.ErrorIs(t, err, ports.ErrorSystemFailure)
	assert.Nil(t, players)
}

func TestGetAllPlayersStatuses_DifferentPhases(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 4,
		playerAssets: map[int]map[string]int{
			1: {"RB": 2, "RC": 3},
			2: {"RA": 2},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 3},
			2: {2, 3},
		},
	}
	ctx := context.Background()

	// Act - Test with phase 4 (end of first cycle)
	players, err := rules.GetAllPlayersStatuses(4, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, players, 2)

	playerMap := make(map[int]ports.PlayerResult)
	for _, player := range players {
		playerMap[player.ID] = player
	}

	player1 := playerMap[1]
	assert.Equal(t, 2, len(player1.Assets))
	assert.Equal(t, 2, player1.Assets["RB"])
	assert.Equal(t, 3, player1.Assets["RC"])
	assert.Equal(t, 9, player1.Score) // RB(4) + RC(5)

	player2 := playerMap[2]
	assert.Equal(t, 1, len(player2.Assets))
	assert.Equal(t, 2, player2.Assets["RA"])
	assert.Equal(t, 3, player2.Score) // RA(3)
}

func TestGetAllPlayersStatuses_AssetPhaseFiltering(t *testing.T) {
	// Arrange
	rules := createTestGameRules()
	mockStore := &MockStore{
		currentPhase: 6,
		playerAssets: map[int]map[string]int{
			1: {"RA": 2, "RC": 6, "RD": 7},
		},
		completedPlayerPhases: map[int][]int{
			1: {2, 6, 7},
		},
	}
	ctx := context.Background()

	// Act
	players, err := rules.GetAllPlayersStatuses(6, mockStore, ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, players, 1)

	player := players[0]
	assert.Equal(t, 1, player.ID)
	// Should only include assets selected after begin phase (5)
	assert.Len(t, player.Assets, 2)
	assert.Equal(t, 6, player.Assets["RC"])
	assert.Equal(t, 7, player.Assets["RD"])
	// Should not include assets from before or at begin phase
	_, hasRA := player.Assets["RA"]
	assert.False(t, hasRA, "RA should not be included (phase 2 < 5)")
}

func TestMakeGameRules_MaxScore(t *testing.T) {
	// Arrange
	asset1 := Asset{ID: "RA", Text: "Test Asset 1", Cost: 3}
	asset2 := Asset{ID: "RB", Text: "Test Asset 2", Cost: 4}
	asset3 := Asset{ID: "RC", Text: "Test Asset 3", Cost: 5}
	asset4 := Asset{ID: "RD", Text: "Test Asset 4", Cost: 13}
	asset5 := Asset{ID: "SA", Text: "Test Asset 5", Cost: 7}
	asset6 := Asset{ID: "SB", Text: "Test Asset 6", Cost: 2}
	asset7 := Asset{ID: "SC", Text: "Test Asset 7", Cost: 55}
	asset8 := Asset{ID: "SD", Text: "Test Asset 8", Cost: 1}

	assets := map[string]Asset{
		"RA": asset1,
		"RB": asset2,
		"RC": asset3,
		"RD": asset4,
		"SA": asset5,
		"SB": asset6,
		"SC": asset7,
		"SD": asset8,
	}

	phases := map[int]Phase{
		1: {ID: 1, Type: "begin", Available: []string{}, Accepted: []Set[string]{}},
		2: {ID: 2, Type: "single", Available: []string{"RA", "RB"}, Accepted: []Set[string]{}},
		3: {ID: 3, Type: "single", Available: []string{"RC", "RD"}, Accepted: []Set[string]{}},
		4: {ID: 4, Type: "end", Available: []string{}, Accepted: []Set[string]{}},
		5: {ID: 5, Type: "begin", Available: []string{}, Accepted: []Set[string]{}},
		6: {ID: 6, Type: "multi", Available: []string{"SA", "SB"}, Accepted: []Set[string]{}},
		7: {ID: 7, Type: "multi", Available: []string{"SC", "SD"}, Accepted: []Set[string]{}},
		8: {ID: 8, Type: "end", Available: []string{}, Accepted: []Set[string]{}},
	}

	// Act
	rules := MakeGameRules(assets, phases)

	// Assert
	assert.NotNil(t, rules.MaxScore, "MaxScore should be initialized")
	assertMaxScore(t, rules, 1, 0)
	assertMaxScore(t, rules, 2, 7)
	assertMaxScore(t, rules, 3, 25)
	assertMaxScore(t, rules, 4, 25)
	assertMaxScore(t, rules, 5, 0)
	assertMaxScore(t, rules, 6, 9)
	assertMaxScore(t, rules, 7, 65)
	assertMaxScore(t, rules, 8, 65)
}

func assertMaxScore(t *testing.T, rules GameRules, phaseID int, expected int) {
	maxScore, exists := rules.MaxScore[phaseID]
	assert.True(t, exists, "MaxScore should exist for phase %d", phaseID)
	assert.Equal(t, expected, maxScore, "MaxScore for phase %d should equal total asset cost", phaseID)
}
