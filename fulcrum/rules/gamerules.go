package gamerules

import (
	"context"
	"fmt"
	"fulcrum/ports"
	"sort"
)

type GameRules struct {
	Assets        map[string]Asset
	Phases        map[int]Phase
	BeginPhaseIDs map[int]int
	EndPhaseIDs   map[int]int
	MaxScore      map[int]int
}

func MakeGameRules(assetsDict map[string]Asset, phasesDict map[int]Phase) GameRules {
	phaseIDs := make([]int, 0, len(phasesDict))
	for k := range phasesDict {
		phaseIDs = append(phaseIDs, k)
	}
	sort.Ints(phaseIDs)
	maxScore := make(map[int]int)
	maxCost := 0
	beginPhaseIDs := make(map[int]int)
	beginPhaseID := 0
	for _, phaseID := range phaseIDs {
		if phasesDict[phaseID].Type == "begin" {
			beginPhaseID = phaseID
			maxCost = 0
		}
		beginPhaseIDs[phaseID] = beginPhaseID
		for _, assetId := range phasesDict[phaseID].Available {
			asset := assetsDict[assetId]
			maxCost += asset.Cost
		}
		maxScore[phaseID] = maxCost
	}
	sort.Sort(sort.Reverse(sort.IntSlice(phaseIDs)))
	endPhaseIDs := make(map[int]int)
	endPhaseID := 0
	for _, phaseID := range phaseIDs {
		if phasesDict[phaseID].Type == "end" {
			endPhaseID = phaseID
		}
		endPhaseIDs[phaseID] = endPhaseID
	}

	return GameRules{
		Assets:        assetsDict,
		Phases:        phasesDict,
		BeginPhaseIDs: beginPhaseIDs,
		EndPhaseIDs:   endPhaseIDs,
		MaxScore:      maxScore,
	}
}

func (r GameRules) AssetSelection(playerID int, phaseID int, assetsIDs []string, store ports.Store, ctxt context.Context) error {
	currentPhaseID, err := store.GetCurrentPhase(ctxt)
	if err != nil {
		return fmt.Errorf("%w: failed to get current phase", ports.ErrorSystemFailure)
	}
	if phaseID != currentPhaseID {
		return fmt.Errorf("%w: selection must be for the current phase", ports.ErrorIncorrectInput)
	}
	currentPhase, exists := r.Phases[currentPhaseID]
	if !exists {
		return fmt.Errorf("%w: current phase configuration not found", ports.ErrorSystemFailure)
	}

	existingAssets, err := store.GetPlayerAssets(ctxt, playerID)
	if err != nil {
		return fmt.Errorf("%w: failed to retrieve player's assets", ports.ErrorSystemFailure)
	}

	newAssets := MakeSet(assetsIDs)
	for assetID, _ := range existingAssets {
		newAssets.Add(assetID)
	}
	if !currentPhase.Accept(newAssets) {
		return fmt.Errorf("%w: selection doesn't meet phase requirements", ports.ErrorIncorrectInput)
	}
	if err := store.SelectMultipleAssets(ctxt, playerID, assetsIDs, phaseID); err != nil {
		return fmt.Errorf("%w: failed to update player assets", ports.ErrorIncorrectInput)
	}
	markPhaseID := phaseID
	for {
		if err := store.MarkPhaseCompletion(ctxt, markPhaseID, playerID); err != nil {
			return fmt.Errorf("%w: failed to mark phase completion", ports.ErrorSystemFailure)
		}
		markPhaseID++
		nextPhase, exists := r.Phases[markPhaseID]
		if !exists {
			break
		}
		if !nextPhase.Accept(newAssets) {
			break
		}
	}
	return nil
}

func (r GameRules) EnsureRegistrationIsAllowed(store ports.Store, ctxt context.Context) (int, error) {
	phaseID, err := store.GetCurrentPhase(ctxt)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to get current phase", ports.ErrorSystemFailure)
	}
	if phaseID == 0 {
		return 0, fmt.Errorf("%w: game has not started yet", ports.ErrorIncorrectInput)
	}
	phase, exists := r.Phases[phaseID]
	if !exists {
		return 0, fmt.Errorf("%w: invalid phase", ports.ErrorSystemFailure)
	}
	if phase.Type != "begin" {
		return 0, fmt.Errorf("%w: registration is currently closed", ports.ErrorIncorrectInput)
	}
	return phaseID, nil
}

func (r GameRules) GetPlayerStatus(phaseID int, playerId int, store ports.Store, ctxt context.Context) ([]string, []int, error) {
	beginPhaseID, exists := r.BeginPhaseIDs[phaseID]
	if !exists {
		return nil, nil, fmt.Errorf("%w: invalid phase", ports.ErrorSystemFailure)
	}
	endPhaseID, exists := r.EndPhaseIDs[phaseID]
	if !exists {
		return nil, nil, fmt.Errorf("%w: invalid phase", ports.ErrorSystemFailure)
	}
	allPlayerAssets, err := store.GetPlayerAssets(ctxt, playerId)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to retrieve player's assets", ports.ErrorSystemFailure)
	}
	completedPlayerPhases, err := store.GetPlayerPhaseCompletion(ctxt, playerId)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to retrieve player's phase completion", ports.ErrorSystemFailure)
	}
	assetIDs, scoreByEnd := r.ComputePlayerScores(phaseID, beginPhaseID, endPhaseID, allPlayerAssets, completedPlayerPhases)
	scores := make([]int, len(scoreByEnd))
	keys := make([]int, 0, len(scoreByEnd))
	for end := range scoreByEnd {
		keys = append(keys, end)
	}
	sort.Ints(keys)
	for i, k := range keys {
		scores[i] = scoreByEnd[k]
	}
	return assetIDs, scores, nil
}

func (r GameRules) GetAllPlayersStatuses(phaseID int, store ports.Store, ctxt context.Context) ([]ports.PlayerResult, error) {
	beginPhaseID, exists := r.BeginPhaseIDs[phaseID]
	if !exists {
		return nil, fmt.Errorf("%w: invalid phase", ports.ErrorSystemFailure)
	}
	endPhaseID, exists := r.EndPhaseIDs[phaseID]
	if !exists {
		return nil, fmt.Errorf("%w: invalid phase", ports.ErrorSystemFailure)
	}

	allGameData, err := store.GetAllGameData(ctxt)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve all game data", ports.ErrorSystemFailure)
	}

	var result []ports.PlayerResult
	for _, player := range allGameData {
		// Filter assets to only include those relevant to the current phase
		relevantAssets := make(map[string]int)
		for assetID, assetPhase := range player.Assets {
			if assetPhase >= beginPhaseID && assetPhase <= endPhaseID {
				relevantAssets[assetID] = assetPhase
			}
		}

		_, scoreByEnd := r.ComputePlayerScores(phaseID, beginPhaseID, endPhaseID, player.Assets, player.CompletedPhases)
		result = append(result, ports.PlayerResult{
			ID:     player.ID,
			Name:   player.Name,
			Assets: relevantAssets,
			Score:  scoreByEnd[endPhaseID],
		})

	}

	return result, nil
}

func (r GameRules) ComputePlayerScores(phaseID int, beginPhaseID int, endPhaseID int, allPlayerAssets map[string]int, completedPlayerPhases []int) ([]string, map[int]int) {
	assetIDs := make([]string, 0, len(r.Assets))
	scoreByEnd := make(map[int]int)

	for _, phase := range r.Phases {
		if phase.ID <= endPhaseID && r.EndPhaseIDs[phase.ID] == phase.ID {
			scoreByEnd[phase.ID] = 0
		}
	}
	for assetID, assetPhase := range allPlayerAssets {
		if assetPhase >= beginPhaseID {
			assetIDs = append(assetIDs, assetID)
		}
		end := r.EndPhaseIDs[assetPhase]
		scoreByEnd[end] += r.Assets[assetID].Cost
	}
	completed := MakeSet(completedPlayerPhases)
	for end := range scoreByEnd {
		begin := r.BeginPhaseIDs[end]
		for phaseId := begin + 1; phaseId < min(end, phaseID+1); phaseId++ {
			if !completed.Contains(phaseId) {
				scoreByEnd[end] = r.MaxScore[end]
			}
		}
	}
	return assetIDs, scoreByEnd
}

var _ ports.Rules = (*GameRules)(nil)
