// cmd/entry_item_cmd_test.go

package cmd

import (
	"testing"
)

func TestParseItemCommand(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		base    string
		detail  string
		split   int
		repeat  int
		fee     float64
		strOut  string
	}{
		// 기본
		{
			input: "커피",
			base:  "커피",
		},
		// 괄호 메모
		{
			input:  "아이템(서울 강남)",
			base:   "아이템",
			detail: "서울 강남",
			strOut: "아이템(서울 강남)",
		},
		// 할부
		{
			input:  "노트북//3",
			base:   "노트북",
			split:  3,
			strOut: "노트북//3",
		},
		// 할부 + 수수료
		{
			input:  "노트북//3(28%)",
			base:   "노트북",
			split:  3,
			fee:    28,
			strOut: "노트북//3(28%)",
		},
		// 반복
		{
			input:  "월세**12",
			base:   "월세",
			repeat: 12,
			strOut: "월세**12",
		},
		// 할부 + 메모
		{
			input:  "노트북(업무용)//3",
			base:   "노트북",
			detail: "업무용",
			split:  3,
			strOut: "노트북(업무용)//3",
		},
		// 할부 + 메모 + 수수료
		{
			input:  "냉장고(삼성)//6(15.5%)",
			base:   "냉장고",
			detail: "삼성",
			split:  6,
			fee:    15.5,
		},
		// 공백 트림
		{
			input: "  커피  ",
			base:  "커피",
		},
		// 에러: 빈 문자열
		{
			input:   "",
			wantErr: true,
		},
		// 에러: 할부 + 반복 동시 사용
		{
			input:   "아이템//3**2",
			wantErr: true,
		},
		// 에러: 수수료만 사용 (할부 없음)
		{
			input:   "아이템(28%)",
			wantErr: true,
		},
		// 에러: 할부 n=1
		{
			input:   "아이템//1",
			wantErr: true,
		},
		// 에러: 반복 n=1
		{
			input:   "아이템**1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, err := ParseItemCommand(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseItemCommand(%q): 에러를 기대했지만 성공: %+v", tt.input, cmd)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseItemCommand(%q): 예상치 못한 에러: %v", tt.input, err)
				return
			}
			if cmd.Base != tt.base {
				t.Errorf("Base: got %q, want %q", cmd.Base, tt.base)
			}
			if cmd.Detail != tt.detail {
				t.Errorf("Detail: got %q, want %q", cmd.Detail, tt.detail)
			}
			if cmd.Split != tt.split {
				t.Errorf("Split: got %d, want %d", cmd.Split, tt.split)
			}
			if cmd.Repeat != tt.repeat {
				t.Errorf("Repeat: got %d, want %d", cmd.Repeat, tt.repeat)
			}
			if cmd.Fee != tt.fee {
				t.Errorf("Fee: got %f, want %f", cmd.Fee, tt.fee)
			}
			// String() 역직렬화 검증 (strOut이 명시된 경우)
			if tt.strOut != "" && cmd.String() != tt.strOut {
				t.Errorf("String(): got %q, want %q", cmd.String(), tt.strOut)
			}
		})
	}
}

func TestItemCommandPreview(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"커피", "커피"},
		{"노트북//3", "노트북 / 3개월 할부"},
		{"노트북//3(28%)", "노트북 / 3개월 할부 (수수료 28.0%)"},
		{"월세**12", "월세 / 12회 반복"},
		{"아이템(메모)//2", "아이템 / 메모: 메모 / 2개월 할부"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd, err := ParseItemCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseItemCommand(%q) 실패: %v", tt.input, err)
			}
			if got := cmd.Preview(); got != tt.want {
				t.Errorf("Preview(): got %q, want %q", got, tt.want)
			}
		})
	}
}
