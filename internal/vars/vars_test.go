package vars

import (
	"reflect"
	"testing"
)

func TestSetGet(t *testing.T) {
	s := New()
	if _, ok := s.Get("target"); ok {
		t.Fatal("Get on empty store reported a binding")
	}
	s.Set("target", "10.0.0.5")
	if v, ok := s.Get("target"); !ok || v != "10.0.0.5" {
		t.Fatalf("Get(target) = %q,%v; want 10.0.0.5,true", v, ok)
	}
	s.Set("target", "10.0.0.6") // overwrite
	if v, _ := s.Get("target"); v != "10.0.0.6" {
		t.Fatalf("overwrite failed, Get(target) = %q", v)
	}
}

func TestNamesSorted(t *testing.T) {
	s := New()
	s.Set("port", "443")
	s.Set("host", "x")
	s.Set("user", "root")
	got := s.Names()
	want := []string{"host", "port", "user"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
}

func TestMissing(t *testing.T) {
	tests := []struct {
		name  string
		bound map[string]string
		tmpl  string
		want  []string
	}{
		{"none referenced", nil, "id -a", nil},
		{"all unbound, first-seen order", nil, "ssh <user>@<host> -p <port>", []string{"user", "host", "port"}},
		{"dedup unbound", nil, "curl <url> -o <url>.html", []string{"url"}},
		{"some bound", map[string]string{"user": "root"}, "ssh <user>@<host>", []string{"host"}},
		{"all bound", map[string]string{"target": "x"}, "nmap <target>", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			for k, v := range tt.bound {
				s.Set(k, v)
			}
			if got := s.Missing(tt.tmpl); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Missing(%q) = %v, want %v", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	s := New()
	s.Set("target", "10.0.0.5")
	s.Set("port", "22")

	got, err := s.Resolve("ssh -p <port> root@<target>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "ssh -p 22 root@10.0.0.5"; got != want {
		t.Fatalf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveUnboundErrors(t *testing.T) {
	s := New()
	s.Set("target", "10.0.0.5")

	got, err := s.Resolve("nmap <target> -w <wordlist>")
	if err == nil {
		t.Fatal("expected error for unbound <wordlist>, got nil")
	}
	if got != "" {
		t.Fatalf("on error Resolve must return empty, got %q (no half-filled commands)", got)
	}
}

func TestResolveNoVars(t *testing.T) {
	s := New()
	got, err := s.Resolve("id -a")
	if err != nil || got != "id -a" {
		t.Fatalf("Resolve(no vars) = %q,%v; want \"id -a\",nil", got, err)
	}
}
