package engine

import (
	"testing"

	"github.com/jessn-dev/nock/internal/vars"
	"github.com/jessn-dev/nock/pkg/format"
)

func fixture() []format.Cheatsheet {
	return []format.Cheatsheet{{
		Name: "recon",
		Commands: []format.Command{
			{ID: "nmap-sv", Name: "nmap service scan", Command: "nmap -sV <target>",
				Intent: "identify open ports and service versions", Tags: []string{"scan", "nmap"}, Risk: format.RiskLow},
			{ID: "gobuster-dir", Name: "gobuster directory brute", Command: "gobuster dir -u <url> -w <wordlist>",
				Intent: "discover hidden web directories", Tags: []string{"web", "brute"}, Risk: format.RiskMedium},
		},
	}}
}

func TestSearchFindsByIntentAndName(t *testing.T) {
	e := New(fixture(), vars.New())
	if got := e.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}

	res := e.Search("service scan")
	if len(res) == 0 || res[0].Command.ID != "nmap-sv" {
		t.Fatalf("expected nmap-sv first, got %+v", res)
	}

	res = e.Search("directories")
	if len(res) == 0 || res[0].Command.ID != "gobuster-dir" {
		t.Fatalf("expected gobuster-dir first, got %+v", res)
	}
}

func TestSearchFindsByID(t *testing.T) {
	e := New(fixture(), vars.New())
	res := e.Search("nmap-sv")
	if len(res) == 0 || res[0].Command.ID != "nmap-sv" {
		t.Fatalf("expected nmap-sv first when searching by id, got %+v", res)
	}
}

func TestSearchEmptyReturnsAll(t *testing.T) {
	e := New(fixture(), vars.New())
	if got := len(e.Search("")); got != 2 {
		t.Fatalf("empty search = %d results, want 2", got)
	}
}

func TestResolveFillsVars(t *testing.T) {
	e := New(fixture(), vars.New())
	e.Vars().Set("target", "10.0.0.5")

	cmd, ok := e.Get("nmap-sv")
	if !ok {
		t.Fatal("nmap-sv not found")
	}
	got, err := e.Resolve(cmd)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := "nmap -sV 10.0.0.5"; got != want {
		t.Fatalf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveErrorsOnMissingVar(t *testing.T) {
	e := New(fixture(), vars.New())
	cmd, _ := e.Get("gobuster-dir")
	if _, err := e.Resolve(cmd); err == nil {
		t.Fatal("expected error for unresolved vars, got nil")
	}
}
