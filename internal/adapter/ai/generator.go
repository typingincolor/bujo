package ai

import (
	"context"

	"github.com/typingincolor/bujo/internal/domain"
)

type SummaryGenerator interface {
	GenerateSummary(ctx context.Context, entries []domain.Entry, horizon domain.SummaryHorizon) (string, error)
}
