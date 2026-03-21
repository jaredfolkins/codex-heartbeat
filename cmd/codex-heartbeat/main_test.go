package main

import (
	"testing"
	"time"
)

func TestParseFlexibleDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  time.Duration
	}{
		{input: "30m", want: 30 * time.Minute},
		{input: "2h", want: 2 * time.Hour},
		{input: "1d", want: 24 * time.Hour},
		{input: "15 minute", want: 15 * time.Minute},
		{input: "15 minutes", want: 15 * time.Minute},
		{input: "2 hour", want: 2 * time.Hour},
		{input: "2 hours", want: 2 * time.Hour},
		{input: "1 day", want: 24 * time.Hour},
		{input: "3 days", want: 72 * time.Hour},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := parseFlexibleDuration(tc.input)
			if err != nil {
				t.Fatalf("parseFlexibleDuration(%q) returned error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("parseFlexibleDuration(%q) = %s, want %s", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseFlexibleDurationRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	inputs := []string{"", "0m", "10w", "minutes", "1.5h"}
	for _, input := range inputs {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			if _, err := parseFlexibleDuration(input); err == nil {
				t.Fatalf("parseFlexibleDuration(%q) unexpectedly succeeded", input)
			}
		})
	}
}
