package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ForwardTestCase struct {
	adminPassword  string
	requestPhaseID int
	setupFakes     func(*FakeStore, *FakeNotifier)
	expectedFakes  func(*testing.T, *FakeStore, *FakeNotifier)
	expectedStatus int
	expectedBody   map[string]interface{}
}

func TestHandleForward_Success(t *testing.T) {
	runForwardTest(t, ForwardTestCase{
		adminPassword:  "verysecretpassword",
		requestPhaseID: 2,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
			store.GetCurrentPhaseFunc = func(ctx context.Context) (int, error) {
				return 1, nil
			}
			store.SetPhaseFunc = func(ctx context.Context, phaseId int) (int, error) {
				return phaseId, nil
			}
		},
		expectedFakes: func(t *testing.T, store *FakeStore, notifier *FakeNotifier) {
			assert.Equal(t, 1, len(notifier.BroadcastMessages), "phase change should have been broadcast")
			assert.Equal(t, map[string]interface{}{
				"type":     "phase_change",
				"phase_id": 2,
			}, notifier.BroadcastMessages[0], "Incorrect broadcast message")

		},
		expectedStatus: http.StatusOK,
		expectedBody: map[string]interface{}{
			"message":  "Phase updated successfully",
			"phase_id": float64(2),
		},
	})
}

func TestHandleForward_InvalidPassword(t *testing.T) {
	runForwardTest(t, ForwardTestCase{
		adminPassword:  "wrong-password",
		requestPhaseID: 2,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
		},
		expectedFakes: func(t *testing.T, store *FakeStore, notifier *FakeNotifier) {
		},
		expectedStatus: http.StatusUnprocessableEntity,
		expectedBody: map[string]interface{}{
			"error": "Invalid admin password",
		},
	})
}

func TestHandleForward_CannotSkipPhases(t *testing.T) {
	runForwardTest(t, ForwardTestCase{
		adminPassword:  "verysecretpassword",
		requestPhaseID: 3,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
			store.GetCurrentPhaseFunc = func(ctx context.Context) (int, error) {
				return 1, nil
			}
		},
		expectedFakes:  func(t *testing.T, store *FakeStore, notifier *FakeNotifier) {},
		expectedStatus: http.StatusUnprocessableEntity,
		expectedBody: map[string]interface{}{
			"error": "Cannot skip phases",
		},
	})
}

func TestHandleForward_AlreadyInPhase(t *testing.T) {
	runForwardTest(t, ForwardTestCase{
		adminPassword:  "verysecretpassword",
		requestPhaseID: 2,
		setupFakes: func(store *FakeStore, notifier *FakeNotifier) {
			store.GetCurrentPhaseFunc = func(ctx context.Context) (int, error) {
				return 2, nil
			}
		},
		expectedFakes:  func(t *testing.T, store *FakeStore, notifier *FakeNotifier) {},
		expectedStatus: http.StatusOK,
		expectedBody: map[string]interface{}{
			"message":  "Already in the requested phase",
			"phase_id": float64(2),
		},
	})
}

func runForwardTest(t *testing.T, testCase ForwardTestCase) {
	// Arrange
	api, fakes := createTestAPI(t, []byte{})
	testCase.setupFakes(fakes.store, fakes.notifier)

	// Act
	rec := CallAPI(t, "/forward", ForwardRequest{
		AdminPassword: testCase.adminPassword,
		PhaseID:       testCase.requestPhaseID,
	}, api.handleForward)

	// Assert
	assert.Equal(t, testCase.expectedStatus, rec.Code)
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, testCase.expectedBody, response)
	testCase.expectedFakes(t, fakes.store, fakes.notifier)
}
