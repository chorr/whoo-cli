package api

import (
	"strings"
	"testing"
)

func TestAPIErrorHidesDetailsByDefault(t *testing.T) {
	err := &APIError{
		Code:     403,
		Endpoint: "/report/assets,liabilities.json",
		Reason:   "forbidden",
		Message:  "API 요청이 거부되었습니다",
		Details:  "<html>blocked</html>",
	}

	message := err.Error()
	if strings.Contains(message, "<html>") {
		t.Fatalf("Error() exposed HTML: %q", message)
	}
	for _, want := range []string{"code=403", "endpoint=/report/assets,liabilities.json", "reason=forbidden"} {
		if !strings.Contains(message, want) {
			t.Errorf("Error() = %q, missing %q", message, want)
		}
	}
}

func TestAPIErrorShowsDetailsInVerboseMode(t *testing.T) {
	err := &APIError{
		Code: 403, Message: "거부됨", Details: "<html>blocked</html>", Verbose: true,
	}
	if message := err.Error(); !strings.Contains(message, "<html>blocked</html>") {
		t.Errorf("Error() = %q, want verbose details", message)
	}
}

func TestCompactDetailsLimitsBody(t *testing.T) {
	body := []byte(strings.Repeat("x", 5000))
	if got := compactDetails(body); len(got) > 4100 {
		t.Errorf("compactDetails length = %d, want <= 4100", len(got))
	}
}
