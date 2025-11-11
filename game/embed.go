package game

import (
	_ "embed"
	"log"

	"gopkg.in/yaml.v3"
)

//go:embed rules.yaml
var YAMLData []byte

type AssetData struct {
	ID   string `yaml:"id"`
	Text string `yaml:"text"`
	Cost int    `yaml:"cost"`
}

type PhaseData struct {
	ID        int         `yaml:"id"`
	Type      string      `yaml:"type"`
	Available []AssetData `yaml:"available"`
	Accepted  []struct {
		Set []AssetData `yaml:"set"`
	} `yaml:"accepted,omitempty"`
}

type RulesData struct {
	Assets []AssetData `yaml:"assets"`
	Phases []PhaseData `yaml:"phases"`
}

func ReadRules() RulesData {
	var rules RulesData
	if err := yaml.Unmarshal(YAMLData, &rules); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}
	return rules
}
