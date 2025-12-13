package shared

// BDDThen represents BDD Then step results for testing
type BDDThen struct {
	Description string `json:"description"`
	Result      string `json:"result"`
	Passed      bool   `json:"passed"`
}

// NewBDDThen creates a new BDD Then step
func NewBDDThen(description, result string, passed bool) *BDDThen {
	return &BDDThen{
		Description: description,
		Result:      result,
		Passed:      passed,
	}
}

// IsSuccess returns true if the BDD step passed
func (bt *BDDThen) IsSuccess() bool {
	return bt.Passed
}
