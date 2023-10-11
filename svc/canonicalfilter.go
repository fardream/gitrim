package svc

import (
	"fmt"

	"github.com/fardream/gitrim"
)

func NewCanonicalFilter(rawtext string) (*Filter, error) {
	lines, err := gitrim.LoadPatternStringFromString(rawtext, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the filter text: %w", err)
	}

	return &Filter{
		RawText:          rawtext,
		CanonicalFilters: lines,
	}, nil
}
