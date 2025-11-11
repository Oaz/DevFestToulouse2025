package gamerules

import (
	"fulcrum/ports"
	"game"
)

func MakeAsset(assetData game.AssetData) Asset {
	return Asset{
		ID:   assetData.ID,
		Text: assetData.Text,
		Cost: assetData.Cost,
	}
}

func GetRules() ports.Rules {
	return LoadRules()
}

func LoadRules() GameRules {
	rules := game.ReadRules()
	assetsDict := make(map[string]Asset, len(rules.Assets))
	phasesDict := make(map[int]Phase, len(rules.Phases))

	for _, assetData := range rules.Assets {
		assetsDict[assetData.ID] = MakeAsset(assetData)
	}

	for _, phaseData := range rules.Phases {
		phase := Phase{
			ID:        phaseData.ID,
			Type:      phaseData.Type,
			Available: make([]string, 0),
			Accepted:  make([]Set[string], 0),
		}

		for _, assetData := range phaseData.Available {
			phase.Available = append(phase.Available, assetData.ID)
		}

		for _, accepted := range phaseData.Accepted {
			set := make(Set[string])
			for _, assetData := range accepted.Set {
				set.Add(assetData.ID)
			}
			phase.Accepted = append(phase.Accepted, set)
		}

		phasesDict[phase.ID] = phase
	}

	return MakeGameRules(assetsDict, phasesDict)
}
