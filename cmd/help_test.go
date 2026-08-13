package cmd

import "testing"

func TestIsHelpArg(t *testing.T) {
	for _, s := range []string{"help", "--help", "-h"} {
		if !isHelpArg(s) {
			t.Errorf("isHelpArg(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "Help", "--h", "flow", "--from"} {
		if isHelpArg(s) {
			t.Errorf("isHelpArg(%q) = true, want false", s)
		}
	}
}

func TestWantsHelp(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{}, false},
		{[]string{"help"}, true},
		{[]string{"--help"}, true},
		{[]string{"-h"}, true},
		{[]string{"flow", "--help"}, true},
		{[]string{"--from", "20260801", "--help"}, true},
		{[]string{"flow", "--from", "20260801"}, false},
		{[]string{"assets"}, false},
	}
	for _, tc := range cases {
		if got := wantsHelp(tc.args); got != tc.want {
			t.Errorf("wantsHelp(%q) = %v, want %v", tc.args, got, tc.want)
		}
	}
}
