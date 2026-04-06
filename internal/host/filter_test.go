package host

import (
	"testing"
)

func TestScore_ExactSubstring(t *testing.T) {
	s := Score("prod", "web-prod-01")
	if s <= 0 {
		t.Errorf("expected positive score for substring match, got %d", s)
	}
}

func TestScore_NoMatch(t *testing.T) {
	s := Score("xyz", "web-prod-01")
	if s != 0 {
		t.Errorf("expected 0 for no match, got %d", s)
	}
}

func TestScore_ExactBeatsFuzzy(t *testing.T) {
	exact := Score("prod", "prod-server")
	fuzzy := Score("prod", "p-r-o-d-server")
	if fuzzy >= exact {
		t.Errorf("exact substring (%d) should beat scattered match (%d)", exact, fuzzy)
	}
}

func TestScore_BoundaryBeatsMiddle(t *testing.T) {
	boundary := Score("prod", "web-prod")
	middle := Score("prod", "reproduce")
	if middle >= boundary {
		t.Errorf("boundary match (%d) should beat mid-word match (%d)", boundary, middle)
	}
}

func TestScore_EmptyQuery(t *testing.T) {
	s := Score("", "anything")
	if s != 0 {
		t.Errorf("expected 0 for empty query, got %d", s)
	}
}

func TestMatch_SearchesAllFields(t *testing.T) {
	h := Host{
		ShortName: "web-01",
		DescText:  "production server",
		LongName:  "web01.example.com",
		IP:        "10.0.1.5",
		Groups:    []string{"Servers"},
	}
	s := Match("production", h)
	if s <= 0 {
		t.Errorf("expected match on description, got %d", s)
	}
	s = Match("10.0", h)
	if s <= 0 {
		t.Errorf("expected match on IP, got %d", s)
	}
	s = Match("server", h)
	if s <= 0 {
		t.Errorf("expected match on group, got %d", s)
	}
}

func TestFilterHosts_NumericPrefixMatchesMenuNumber(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", MenuNumber: 12, DescText: "desc"},
		{ShortName: "b", MenuNumber: 3, DescText: "desc"},
		{ShortName: "c", MenuNumber: 123, DescText: "desc"},
	}
	result := FilterHosts("12", hosts)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
}

func TestFilterHosts_EmptyQueryReturnsAll(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", MenuNumber: 1},
		{ShortName: "b", MenuNumber: 2},
	}
	result := FilterHosts("", hosts)
	if len(result) != 2 {
		t.Errorf("expected all hosts returned, got %d", len(result))
	}
}
