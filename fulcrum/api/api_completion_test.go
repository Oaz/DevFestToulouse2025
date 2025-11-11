package api

import (
	"context"
	"encoding/json"
	"fulcrum/ports"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CompletionTestCase struct {
	name           string
	requestPhaseID int
	setupFakes     func(*FakeStore, *FakeNotifier)
	expectedStatus int
	expectedBody   map[string]interface{}
}

func TestHandleCompletion_Success(t *testing.T) {
	runCompletionTest(t, CompletionTestCase{
		name:           "Success with some players completed",
		requestPhaseID: 2,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
			store.GetPhaseCompletionFunc = func(ctx context.Context, phaseID int) (int, error) {
				assert.Equal(t, 2, phaseID)
				return 3, nil
			}
			store.GetPlayersCountFunc = func(ctx context.Context) (int, error) {
				return 5, nil
			}
		},
		expectedStatus: http.StatusOK,
		expectedBody: map[string]interface{}{
			"completed_count": float64(3),
			"total_players":   float64(5),
		},
	})
}

func TestHandleCompletion_ZeroCompleted(t *testing.T) {
	runCompletionTest(t, CompletionTestCase{
		name:           "Success with zero players completed",
		requestPhaseID: 1,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
			store.GetPhaseCompletionFunc = func(ctx context.Context, phaseID int) (int, error) {
				assert.Equal(t, 1, phaseID)
				return 0, nil
			}
			store.GetPlayersCountFunc = func(ctx context.Context) (int, error) {
				return 10, nil
			}
		},
		expectedStatus: http.StatusOK,
		expectedBody: map[string]interface{}{
			"completed_count": float64(0),
			"total_players":   float64(10),
		},
	})
}

func TestHandleCompletion_AllCompleted(t *testing.T) {
	runCompletionTest(t, CompletionTestCase{
		name:           "Success with all players completed",
		requestPhaseID: 3,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
			store.GetPhaseCompletionFunc = func(ctx context.Context, phaseID int) (int, error) {
				assert.Equal(t, 3, phaseID)
				return 7, nil
			}
			store.GetPlayersCountFunc = func(ctx context.Context) (int, error) {
				return 7, nil
			}
		},
		expectedStatus: http.StatusOK,
		expectedBody: map[string]interface{}{
			"completed_count": float64(7),
			"total_players":   float64(7),
		},
	})
}

func TestHandleCompletion_CompletionError(t *testing.T) {
	runCompletionTest(t, CompletionTestCase{
		name:           "Error getting phase completion",
		requestPhaseID: 4,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
			store.GetPhaseCompletionFunc = func(ctx context.Context, phaseID int) (int, error) {
				return 0, ports.ErrorSystemFailure
			}
		},
		expectedStatus: http.StatusInternalServerError,
		expectedBody: map[string]interface{}{
			"error": "Failed to get phase completion data",
		},
	})
}

func TestHandleCompletion_PlayersCountError(t *testing.T) {
	runCompletionTest(t, CompletionTestCase{
		name:           "Error getting players count",
		requestPhaseID: 2,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
			store.GetPhaseCompletionFunc = func(ctx context.Context, phaseID int) (int, error) {
				return 5, nil
			}
			store.GetPlayersCountFunc = func(ctx context.Context) (int, error) {
				return 0, ports.ErrorSystemFailure
			}
		},
		expectedStatus: http.StatusInternalServerError,
		expectedBody: map[string]interface{}{
			"error": "Failed to get total players count",
		},
	})
}

func runCompletionTest(t *testing.T, testCase CompletionTestCase) {
	// Arrange
	api, fakes := createTestAPI(t, []byte{})
	testCase.setupFakes(fakes.store, fakes.notifier)

	// Act
	rec := CallAPI(t, "/completion", CompletionRequest{
		PhaseID: testCase.requestPhaseID,
	}, api.handleCompletion)

	// Assert
	assert.Equal(t, testCase.expectedStatus, rec.Code, "Test case: %s", testCase.name)
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err, "Test case: %s", testCase.name)
	assert.Equal(t, testCase.expectedBody, response, "Test case: %s", testCase.name)
}
