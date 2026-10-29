package cmd

import (
	"testing"
)

func TestBackspaceRunes(t *testing.T) {
	if got := backspaceRunes("커피a"); got != "커피" {
		t.Fatalf("backspaceRunes = %q, want %q", got, "커피")
	}
	if got := backspaceRunes(""); got != "" {
		t.Fatalf("empty backspace = %q", got)
	}
	if got := backspaceRunes("한"); got != "" {
		t.Fatalf("single hangul backspace = %q", got)
	}
}

func TestAddMonthsYYYYMM(t *testing.T) {
	// 202607 + 11 = 202706 (12개월 구간)
	if got := addMonthsYYYYMM(202607, 11); got != 202706 {
		t.Fatalf("addMonthsYYYYMM(202607,11)=%d, want 202706", got)
	}
	// 연말 경계
	if got := addMonthsYYYYMM(202512, 1); got != 202601 {
		t.Fatalf("addMonthsYYYYMM(202512,1)=%d, want 202601", got)
	}
	// 잘못된 산술 회귀: from+11 은 202618 같은 비정상 값 생성
	if bad := 202607 + 11; bad == 202618 {
		if addMonthsYYYYMM(202607, 11) == bad {
			t.Fatal("helper must not equal naive ym+11 arithmetic")
		}
	}
}

func TestValidateDateInputMaxOneYear(t *testing.T) {
	if err := validateDateInput("20260101", "20270101"); err != nil {
		t.Fatalf("exactly 1 year should be allowed (not after start+1y): %v", err)
	}
	if err := validateDateInput("20260101", "20270102"); err == nil {
		t.Fatal("more than 1 year should fail")
	}
}
