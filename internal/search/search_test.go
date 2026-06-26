package search

import (
	"reflect"
	"testing"
)

func ids(rs []Result) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.ID
	}
	return out
}

func TestRankEmptyQueryReturnsAllInOrder(t *testing.T) {
	items := []Item{{ID: "a", Text: "alpha"}, {ID: "b", Text: "bravo"}, {ID: "c", Text: "charlie"}}
	got := Rank("  ", items)
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(ids(got), want) {
		t.Fatalf("empty query order = %v, want %v", ids(got), want)
	}
	for _, r := range got {
		if r.Score != 0 {
			t.Fatalf("empty query should score 0, got %d for %s", r.Score, r.ID)
		}
	}
}

func TestRankDropsNonMatches(t *testing.T) {
	items := []Item{{ID: "hit", Text: "nmap service scan"}, {ID: "miss", Text: "gobuster dir"}}
	got := Rank("nmap", items)
	if want := []string{"hit"}; !reflect.DeepEqual(ids(got), want) {
		t.Fatalf("Rank(nmap) = %v, want %v (non-match dropped)", ids(got), want)
	}
}

func TestRankSubsequenceMatches(t *testing.T) {
	// "nsv" is a subsequence of "nmap service version" but not "gobuster".
	items := []Item{{ID: "svc", Text: "nmap service version"}, {ID: "gob", Text: "gobuster"}}
	got := Rank("nsv", items)
	if want := []string{"svc"}; !reflect.DeepEqual(ids(got), want) {
		t.Fatalf("subsequence Rank(nsv) = %v, want %v", ids(got), want)
	}
}

func TestRankOrdersBetterMatchFirst(t *testing.T) {
	// "scan" matches both, but contiguously/at a boundary in "scan ports" and
	// only as a scattered subsequence in "s...c...a...n". Boundary+start wins.
	items := []Item{
		{ID: "scattered", Text: "service cache analysis network"},
		{ID: "exact", Text: "scan ports"},
	}
	got := Rank("scan", items)
	if got[0].ID != "exact" {
		t.Fatalf("expected 'exact' first, got order %v", ids(got))
	}
	if got[0].Score <= got[1].Score {
		t.Fatalf("expected exact (%d) to outscore scattered (%d)", got[0].Score, got[1].Score)
	}
}

func TestRankCaseInsensitive(t *testing.T) {
	items := []Item{{ID: "a", Text: "NMAP Service Scan"}}
	if got := Rank("nmap", items); len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("case-insensitive match failed: %v", ids(got))
	}
}
