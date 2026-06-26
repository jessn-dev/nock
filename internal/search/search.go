// Package search implements nock's zero-dependency fuzzy matcher. This is the
// Milestone-1 ranking path: instant, offline, no AI. The AI ranker (Milestone 4)
// layers on top of — and reuses — these Match scores as a prior.
//
// The algorithm is a subsequence match with positional bonuses, the same family
// fzf uses: query characters must appear in order within the target, and matches
// that are contiguous, at word boundaries, or near the start score higher.
package search

import (
	"sort"
	"strings"
)

// Item is anything searchable: an opaque ID plus the haystack text to match against.
type Item struct {
	ID   string
	Text string
}

// Result pairs an item ID with its match score. Higher is better.
type Result struct {
	ID    string
	Score int
}

// Rank returns items matching query, best first. An empty query returns all items
// in their original order (score 0). Non-matching items are dropped.
func Rank(query string, items []Item) []Result {
	if strings.TrimSpace(query) == "" {
		out := make([]Result, len(items))
		for i, it := range items {
			out[i] = Result{ID: it.ID, Score: 0}
		}
		return out
	}
	q := strings.ToLower(query)
	var out []Result
	for _, it := range items {
		if score, ok := match(q, strings.ToLower(it.Text)); ok {
			out = append(out, Result{ID: it.ID, Score: score})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

// match returns a score if every rune of query appears in target in order.
func match(query, target string) (int, bool) {
	const (
		bonusContiguous = 8
		bonusBoundary   = 10
		bonusStart      = 12
		penaltyGap      = 1
	)
	score := 0
	ti := 0
	prevMatched := -2
	tr := []rune(target)
	for _, qr := range query {
		found := false
		for ; ti < len(tr); ti++ {
			if tr[ti] != qr {
				continue
			}
			switch {
			case ti == 0:
				score += bonusStart
			case isBoundary(tr[ti-1]):
				score += bonusBoundary
			case ti == prevMatched+1:
				score += bonusContiguous
			default:
				score -= penaltyGap
			}
			prevMatched = ti
			ti++
			found = true
			break
		}
		if !found {
			return 0, false
		}
	}
	return score, true
}

func isBoundary(r rune) bool {
	return r == ' ' || r == '-' || r == '_' || r == '/' || r == '.' || r == ':'
}
