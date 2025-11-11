package gamerules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssets(t *testing.T) {
	assets := LoadRules().Assets

	assert.Equal(t, 29, len(assets))

	assert.Contains(t, assets, "RA", "asset RA must be present")
	ra := assets["RA"]
	assert.Equal(t, "Quand je reçois la balle, je l'envoie à Alice", ra.Text)
	assert.Equal(t, 3, ra.Cost)

	assert.Contains(t, assets, "RB", "asset RB must be present")
	rb := assets["RB"]
	assert.Equal(t, "Quand je reçois la balle, je l'envoie à Bob", rb.Text)
	assert.Equal(t, 3, rb.Cost)
}

func TestPhase2(t *testing.T) {
	phases := LoadRules().Phases

	assert.Equal(t, 18, len(phases))

	p, ok := phases[2]
	assert.True(t, ok, "phase must exist")
	assert.Equal(t, "single", p.Type)
	expectedAvail := MakeSet([]string{"RB", "RC"})
	avail := MakeSet(p.Available)
	for a := range avail.Difference(expectedAvail) {
		t.Fatalf("phase has unexpected asset %s in Available", a)
	}
	for a := range expectedAvail.Difference(avail) {
		t.Fatalf("phase has missing asset %s in Available", a)
	}
	assert.Equal(t, 2, len(p.Accepted))
}

func TestPhase10(t *testing.T) {
	phases := LoadRules().Phases

	p, ok := phases[10]
	assert.True(t, ok, "phase must exist")
	assert.Equal(t, "multi", p.Type)
	expectedAvail := MakeSet([]string{"SA", "SB", "SC", "SD", "SE"})
	avail := MakeSet(p.Available)
	for a := range avail.Difference(expectedAvail) {
		t.Fatalf("phase has unexpected asset %s in Available", a)
	}
	for a := range expectedAvail.Difference(avail) {
		t.Fatalf("phase has missing asset %s in Available", a)
	}
	expectedAccepted := []Set[string]{
		{
			"SC": {},
		},
		{
			"SD": {},
		},
		{
			"SE": {},
		},
	}
	for i, expected := range expectedAccepted {
		actual := p.Accepted[i]
		for a := range actual.Difference(expected) {
			t.Fatalf("phase has unexpected asset %s in Accepted", a)
		}
		for a := range expected.Difference(actual) {
			t.Fatalf("phase has missing asset %s in Accepted", a)
		}
	}
}
