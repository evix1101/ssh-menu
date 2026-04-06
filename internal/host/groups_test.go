package host

import (
	"testing"
)

func TestAssignMenuNumbers_ExplicitNumbers(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", DescText: "desc", MenuNumber: 3},
		{ShortName: "b", DescText: "desc", MenuNumber: 1},
	}
	result, err := AssignMenuNumbers(hosts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].ShortName != "b" || result[0].MenuNumber != 1 {
		t.Errorf("expected first host to be 'b' with number 1, got %s with %d", result[0].ShortName, result[0].MenuNumber)
	}
	if result[1].ShortName != "a" || result[1].MenuNumber != 3 {
		t.Errorf("expected second host to be 'a' with number 3, got %s with %d", result[1].ShortName, result[1].MenuNumber)
	}
}

func TestAssignMenuNumbers_AutoAssign(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", DescText: "desc", MenuNumber: 2},
		{ShortName: "b", DescText: "desc", MenuNumber: 0},
		{ShortName: "c", DescText: "desc", MenuNumber: 0},
	}
	result, err := AssignMenuNumbers(hosts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	numbers := map[string]int{}
	for _, h := range result {
		numbers[h.ShortName] = h.MenuNumber
	}
	if numbers["a"] != 2 {
		t.Errorf("expected a=2, got %d", numbers["a"])
	}
	if numbers["b"] != 1 {
		t.Errorf("expected b=1, got %d", numbers["b"])
	}
	if numbers["c"] != 3 {
		t.Errorf("expected c=3, got %d", numbers["c"])
	}
}

func TestAssignMenuNumbers_DuplicateError(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", DescText: "desc", MenuNumber: 1},
		{ShortName: "b", DescText: "desc", MenuNumber: 1},
	}
	_, err := AssignMenuNumbers(hosts)
	if err == nil {
		t.Fatal("expected error for duplicate menu numbers")
	}
}

func TestGetAllGroups_SortedUngroupedLast(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", Groups: []string{"Zebra"}},
		{ShortName: "b", Groups: []string{"Alpha"}},
		{ShortName: "c", Groups: []string{}},
	}
	groups := GetAllGroups(hosts)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d: %v", len(groups), groups)
	}
	if groups[0] != "Alpha" {
		t.Errorf("expected first group Alpha, got %s", groups[0])
	}
	if groups[1] != "Zebra" {
		t.Errorf("expected second group Zebra, got %s", groups[1])
	}
	if groups[2] != "Ungrouped" {
		t.Errorf("expected last group Ungrouped, got %s", groups[2])
	}
}

func TestSortWithPins_PinnedFirst(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", MenuNumber: 1, Pinned: false},
		{ShortName: "b", MenuNumber: 2, Pinned: true},
		{ShortName: "c", MenuNumber: 3, Pinned: false},
	}
	result := SortWithPins(hosts)
	if result[0].ShortName != "b" {
		t.Errorf("expected pinned host first, got %s", result[0].ShortName)
	}
	if result[1].ShortName != "a" {
		t.Errorf("expected a second, got %s", result[1].ShortName)
	}
	if result[2].ShortName != "c" {
		t.Errorf("expected c third, got %s", result[2].ShortName)
	}
}
