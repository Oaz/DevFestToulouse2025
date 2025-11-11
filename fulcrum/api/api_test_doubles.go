package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fulcrum/ports"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CallAPI(t *testing.T, route string, body any, handler func(w http.ResponseWriter, r *http.Request)) *httptest.ResponseRecorder {
	reqBytes, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, route, bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

type FakeStore struct {
	RegisterPlayerFunc       func(ctx context.Context, playerName, secret string) (int, error)
	GetPlayersCountFunc      func(ctx context.Context) (int, error)
	VerifyPlayerSecretFunc   func(ctx context.Context, playerID int, secret string) (bool, error)
	SelectMultipleAssetsFunc func(ctx context.Context, playerID int, assetIDs []string, phaseID int) error
	MarkPhaseCompletionFunc  func(ctx context.Context, phaseID int, playerID int) error
	GetPhaseCompletionFunc   func(ctx context.Context, phaseID int) (int, error)
	GetPlayerAssetsFunc      func(ctx context.Context, playerID int) (map[string]int, error)
	GetAllGameDataFunc       func(ctx context.Context) ([]ports.Player, error)
	SetPhaseFunc             func(ctx context.Context, phaseId int) (int, error)
	GetCurrentPhaseFunc      func(ctx context.Context) (int, error)
}

func (m *FakeStore) MarkPhaseCompletion(ctx context.Context, phaseID int, playerID int) error {
	return m.MarkPhaseCompletionFunc(ctx, phaseID, playerID)
}

func (m *FakeStore) GetPhaseCompletion(ctx context.Context, phaseID int) (int, error) {
	return m.GetPhaseCompletionFunc(ctx, phaseID)
}

func (m *FakeStore) GetPlayerPhaseCompletion(ctx context.Context, playerID int) ([]int, error) {
	return nil, nil
}

func (m *FakeStore) SetPhase(ctx context.Context, phaseId int) (int, error) {
	return m.SetPhaseFunc(ctx, phaseId)
}

func (m *FakeStore) GetCurrentPhase(ctx context.Context) (int, error) {
	return m.GetCurrentPhaseFunc(ctx)
}

func (m *FakeStore) RegisterPlayer(ctx context.Context, playerName, secret string) (int, error) {
	return m.RegisterPlayerFunc(ctx, playerName, secret)
}

func (m *FakeStore) GetPlayersCount(ctx context.Context) (int, error) {
	return m.GetPlayersCountFunc(ctx)
}

func (m *FakeStore) VerifyPlayerSecret(ctx context.Context, playerID int, secret string) (bool, error) {
	return m.VerifyPlayerSecretFunc(ctx, playerID, secret)
}

func (m *FakeStore) SelectMultipleAssets(ctx context.Context, playerID int, assetIDs []string, phaseID int) error {
	return m.SelectMultipleAssetsFunc(ctx, playerID, assetIDs, phaseID)
}

func (m *FakeStore) GetPlayerAssets(ctx context.Context, playerID int) (map[string]int, error) {
	return m.GetPlayerAssetsFunc(ctx, playerID)
}

func (m *FakeStore) GetAllGameData(ctx context.Context) ([]ports.Player, error) {
	return m.GetAllGameDataFunc(ctx)
}

// Ensure FakeStore implements the Store interface
var _ ports.Store = (*FakeStore)(nil)

// FakeNotifier implements the signaling.Notifier interface for testing
type FakeNotifier struct {
	HandleWebSocketFunc func(w http.ResponseWriter, r *http.Request)
	BroadcastMessages   []any
}

func (m *FakeNotifier) Run() {
}

func (m *FakeNotifier) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	m.HandleWebSocketFunc(w, r)
}

func (m *FakeNotifier) BroadcastJson(message any) error {
	if m.BroadcastMessages == nil {
		m.BroadcastMessages = make([]any, 0)
	}
	m.BroadcastMessages = append(m.BroadcastMessages, message)
	return nil
}

var _ ports.Notifier = (*FakeNotifier)(nil)

type FakeRules struct {
	CurrentPhase        int
	RegistrationError   error
	AssetSelectionError error
	RelevantAssets      []string
	Scores              []int
	StatusError         error
	PlayerResults       []ports.PlayerResult
}

func (r FakeRules) EnsureRegistrationIsAllowed(store ports.Store, ctxt context.Context) (int, error) {
	return r.CurrentPhase, r.RegistrationError
}

func (r FakeRules) GetPlayerStatus(phaseID int, playerId int, store ports.Store, ctxt context.Context) ([]string, []int, error) {
	return r.RelevantAssets, r.Scores, r.StatusError
}

func (r FakeRules) GetAllPlayersStatuses(phaseID int, store ports.Store, ctxt context.Context) ([]ports.PlayerResult, error) {
	return r.PlayerResults, r.StatusError
}

func (r FakeRules) AssetSelection(playerID int, phaseID int, assetsIDs []string, store ports.Store, ctxt context.Context) error {
	return r.AssetSelectionError
}

var _ ports.Rules = (*FakeRules)(nil)

// StubRandomGenerator allows controlling random number generation in tests
type StubRandomGenerator struct {
	Value []byte
}

func (m *StubRandomGenerator) Read(p []byte) (n int, err error) {
	copy(p, m.Value)
	return len(p), nil
}

type Fakes struct {
	store    *FakeStore
	notifier *FakeNotifier
	rules    *FakeRules
}

func createTestAPI(t *testing.T, randomData []byte) (*API, Fakes) {
	store := &FakeStore{}
	notifier := &FakeNotifier{
		HandleWebSocketFunc: func(w http.ResponseWriter, r *http.Request) {},
	}
	mockRandom := &StubRandomGenerator{
		Value: randomData,
	}
	rules := &FakeRules{
		RegistrationError:   nil,
		AssetSelectionError: nil,
		RelevantAssets:      nil,
		StatusError:         nil,
	}
	api := NewAPI(store, notifier, rules, "verysecretpassword", mockRandom)

	return api, Fakes{
		store:    store,
		notifier: notifier,
		rules:    rules,
	}
}
