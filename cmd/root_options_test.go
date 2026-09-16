package cmd

import (
	"reflect"
	"testing"

	"whoo-cli/config"
)

func TestParseGlobalOptionsPrecedence(t *testing.T) {
	t.Setenv("WHOO_SECTION", "s-env")
	cfg := &config.Config{SectionID: "s-config"}

	args, options, err := ParseGlobalOptions(cfg, []string{
		"entries", "search", "--section", "s-flag", "--verbose", "--flat",
	})
	if err != nil {
		t.Fatalf("ParseGlobalOptions() error = %v", err)
	}
	wantArgs := []string{"entries", "search", "--flat"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args = %v, want %v", args, wantArgs)
	}
	if cfg.SectionID != "s-flag" {
		t.Errorf("SectionID = %q, want s-flag", cfg.SectionID)
	}
	if !cfg.Verbose {
		t.Error("Verbose = false, want true")
	}
	if !options.SectionOverridden {
		t.Error("SectionOverridden = false, want true")
	}
}

func TestParseGlobalOptionsUsesEnvironment(t *testing.T) {
	t.Setenv("WHOO_SECTION", "s-env")
	cfg := &config.Config{SectionID: "s-config"}

	_, _, err := ParseGlobalOptions(cfg, []string{"bs"})
	if err != nil {
		t.Fatalf("ParseGlobalOptions() error = %v", err)
	}
	if cfg.SectionID != "s-env" {
		t.Errorf("SectionID = %q, want s-env", cfg.SectionID)
	}
}

func TestParseGlobalOptionsRequiresSectionValue(t *testing.T) {
	t.Setenv("WHOO_SECTION", "")
	_, _, err := ParseGlobalOptions(&config.Config{}, []string{"bs", "--section"})
	if err == nil {
		t.Fatal("ParseGlobalOptions() error = nil, want error")
	}
}

func TestAddOutputSection(t *testing.T) {
	SetOutputSection("s-test")
	t.Cleanup(func() { SetOutputSection("") })

	object := map[string]interface{}{"code": 200}
	addOutputSection(object)
	if object["section_id"] != "s-test" {
		t.Errorf("object section_id = %v, want s-test", object["section_id"])
	}

	array := []interface{}{map[string]interface{}{"account_id": "x2"}}
	addOutputSection(array)
	row := array[0].(map[string]interface{})
	if row["section_id"] != "s-test" {
		t.Errorf("row section_id = %v, want s-test", row["section_id"])
	}
}

func TestIsSectionWriteCommand(t *testing.T) {
	if !isSectionWriteCommand([]string{"entries", "add"}) {
		t.Error("entries add should be a write command")
	}
	if isSectionWriteCommand([]string{"entries", "search"}) {
		t.Error("entries search should not be a write command")
	}
}
