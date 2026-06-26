package ai

import (
	"context"
	"testing"

	"github.com/jessn-dev/nock/pkg/format"
)

func cands() []Candidate {
	return []Candidate{
		{Command: format.Command{ID: "low"}, Prior: 1},
		{Command: format.Command{ID: "high"}, Prior: 9},
		{Command: format.Command{ID: "mid"}, Prior: 5},
	}
}

func TestRankNilProviderFallsBackToPrior(t *testing.T) {
	r := NewRanker(nil)
	if r.Enabled() {
		t.Fatal("Enabled() should be false with nil provider")
	}
	got := r.Rank(context.Background(), "anything", cands())
	want := []string{"high", "mid", "low"}
	for i, w := range want {
		if got[i].ID != w {
			t.Fatalf("position %d = %q, want %q (full: %+v)", i, got[i].ID, w, got)
		}
	}
}

// failingProvider always errors, simulating a down backend or missing key.
type failingProvider struct{}

func (failingProvider) Name() string { return "failing" }
func (failingProvider) Suggest(context.Context, string, []Candidate) ([]Ranked, error) {
	return nil, ErrNotImplemented
}

func TestRankProviderErrorFallsBackToPrior(t *testing.T) {
	r := NewRanker(failingProvider{})
	if !r.Enabled() {
		t.Fatal("Enabled() should be true with a non-nil provider")
	}
	got := r.Rank(context.Background(), "x", cands())
	if got[0].ID != "high" {
		t.Fatalf("expected fuzzy fallback ordering, got %+v", got)
	}
}
