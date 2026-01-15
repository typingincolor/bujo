package domain

import (
	"errors"
	"strings"
)

type PromptType string

const (
	PromptTypeSummaryDaily  PromptType = "summary-daily"
	PromptTypeSummaryWeekly PromptType = "summary-weekly"
	PromptTypeAsk           PromptType = "ask"
)

var validPromptTypes = map[PromptType]bool{
	PromptTypeSummaryDaily:  true,
	PromptTypeSummaryWeekly: true,
	PromptTypeAsk:           true,
}

func (pt PromptType) String() string {
	return string(pt)
}

func (pt PromptType) IsValid() bool {
	return validPromptTypes[pt]
}

type PromptTemplate struct {
	Type     PromptType
	Content  string
	Filename string
}

func (t PromptTemplate) Validate() error {
	if !t.Type.IsValid() {
		return errors.New("invalid prompt type")
	}
	if strings.TrimSpace(t.Content) == "" {
		return errors.New("prompt content cannot be empty")
	}
	return nil
}

func PromptTypeFromHorizon(horizon SummaryHorizon) PromptType {
	switch horizon {
	case SummaryHorizonDaily:
		return PromptTypeSummaryDaily
	case SummaryHorizonWeekly:
		return PromptTypeSummaryWeekly
	default:
		return PromptTypeSummaryDaily
	}
}
