package ai

import (
	"context"
	"sort"
)

// Ranker wraps a Provider and guarantees graceful degradation: if the provider is
// nil, unimplemented, or errors, the caller's prior (fuzzy) ordering is preserved.
// Frontends call Rank unconditionally and never branch on "is AI configured".
type Ranker struct {
	provider Provider
}

// NewRanker returns a Ranker. A nil provider is valid and means "fuzzy only".
func NewRanker(p Provider) *Ranker { return &Ranker{provider: p} }

// Enabled reports whether a real provider backs this ranker.
func (r *Ranker) Enabled() bool { return r.provider != nil }

// Rank returns candidate IDs best-first. When no provider is available or the
// provider fails, candidates are returned ordered by their fuzzy Prior — never an
// error, because losing AI must never break search.
func (r *Ranker) Rank(ctx context.Context, intent string, candidates []Candidate) []Ranked {
	if r.provider != nil {
		if out, err := r.provider.Suggest(ctx, intent, candidates); err == nil && len(out) > 0 {
			return out
		}
	}
	return byPrior(candidates)
}

func byPrior(candidates []Candidate) []Ranked {
	out := make([]Ranked, len(candidates))
	for i, c := range candidates {
		out[i] = Ranked{ID: c.Command.ID, Score: float64(c.Prior)}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}
