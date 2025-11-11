package gamerules

type Asset struct {
	ID   string
	Text string
	Cost int
}

type Phase struct {
	ID        int
	Type      string
	Available []string
	Accepted  []Set[string]
}

func (phase Phase) Accept(assets Set[string]) bool {
	if phase.Type == "begin" || phase.Type == "end" {
		return true
	}
	isValid := false
	for _, accepted := range phase.Accepted {
		if accepted.IsSubset(assets) {
			isValid = true
			break
		}
	}
	return isValid
}
