package cmd

import (
	"testing"
	"time"
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

func TestCalcDateRangeLastMonthOnMonthEnd(t *testing.T) {
	// 고정: 2026-03-31 기준으로 지난 달은 2월 전체
	// calcDateRange는 time.Now()를 쓰므로 헬퍼 로직만 검증
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.Local)
	thisMonth := firstOfMonth(now)
	lastStart := thisMonth.AddDate(0, -1, 0)
	lastEnd := thisMonth.AddDate(0, 0, -1)
	if lastStart.Format("20060102") != "20260201" {
		t.Fatalf("last month start = %s", lastStart.Format("20060102"))
	}
	if lastEnd.Format("20060102") != "20260228" {
		t.Fatalf("last month end = %s", lastEnd.Format("20060102"))
	}

	// 깨진 방식: AddDate(0,-1,0) on Mar 31
	broken := now.AddDate(0, -1, 0)
	if broken.Month() == time.March {
		// Go normalizes Mar 31 - 1 month → Mar 2/3; must not use for start
		if broken.Format("200601") == lastStart.Format("200601") {
			t.Fatal("unexpected: broken calc matched fixed calc")
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
