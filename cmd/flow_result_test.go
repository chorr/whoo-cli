package cmd

import (
	"strings"
	"testing"
)

func TestParseFlowGroupsMapShape(t *testing.T) {
	raw := []byte(`{
		"code": 200,
		"results": {
			"assets": {
				"total": {"from": 100, "to": 10, "margin": 90},
				"accounts": {
					"x1": {"from": 80, "to": 0, "margin": 80, "account_id": "x1"},
					"x2": {"from": 0, "to": 0, "margin": 0, "account_id": "x2"}
				}
			},
			"expenses": {
				"total": {"from": 0, "to": 50, "margin": -50},
				"accounts": {
					"x9": {"from": 0, "to": 50, "margin": -50, "account_id": "x9"}
				}
			}
		}
	}`)
	groups, err := parseFlowGroups(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) < 2 {
		t.Fatalf("groups=%d", len(groups))
	}
	lines := renderFlowGroupsLines(groups, nil)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "합계") {
		t.Fatal("missing total row")
	}
	if !strings.Contains(joined, "x1") {
		t.Fatal("missing account x1")
	}
	// zero margin account hidden
	if strings.Contains(joined, "x2") {
		t.Fatal("zero account should be hidden")
	}
}

func TestParseChangesMapShape(t *testing.T) {
	raw := []byte(`{
		"code": 200,
		"results": {
			"aggregate": {"in": 1000, "out": 200},
			"rows_type": "day",
			"rows": {
				"20260702": 100,
				"20260701": 50,
				"20260703": -20
			}
		}
	}`)
	v, err := parseChangesView(raw)
	if err != nil {
		t.Fatal(err)
	}
	if v.In != 1000 || v.Out != 200 {
		t.Fatalf("agg in=%v out=%v", v.In, v.Out)
	}
	if len(v.Rows) != 3 {
		t.Fatalf("rows=%d", len(v.Rows))
	}
	// sorted by date
	if v.Rows[0].Date != "20260701" {
		t.Fatalf("first=%s", v.Rows[0].Date)
	}
	lines := renderChangesLines(v)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "In") || !strings.Contains(joined, "2026-07-01") {
		t.Fatal(joined)
	}
}

func TestParseChangesArrayShape(t *testing.T) {
	raw := []byte(`{
		"code": 200,
		"results": {
			"aggregate": {"in": 1, "out": 0},
			"rows_type": "day",
			"rows": [
				{"date": "20260701", "money": 10},
				{"date": "20260702", "money": -5}
			]
		}
	}`)
	v, err := parseChangesView(raw)
	if err != nil || len(v.Rows) != 2 {
		t.Fatalf("err=%v rows=%v", err, v)
	}
}

func TestSliceViewport(t *testing.T) {
	lines := []string{"a", "b", "c", "d", "e"}
	view, off := sliceViewport(lines, 1, 2)
	if off != 1 || len(view) != 2 || view[0] != "b" {
		t.Fatalf("view=%v off=%d", view, off)
	}
	view, off = sliceViewport(lines, 99, 2)
	if off != 3 || view[0] != "d" {
		t.Fatalf("clamped view=%v off=%d", view, off)
	}
}
