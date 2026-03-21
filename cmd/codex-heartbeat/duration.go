package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationPattern = regexp.MustCompile(`(?i)^\s*(\d+)\s*([a-z]+)\s*$`)

type durationFlag struct {
	value time.Duration
	set   bool
}

func (d *durationFlag) Set(raw string) error {
	parsed, err := parseFlexibleDuration(raw)
	if err != nil {
		return err
	}

	d.value = parsed
	d.set = true
	return nil
}

func (d *durationFlag) String() string {
	if !d.set {
		return ""
	}
	return formatFlexibleDuration(d.value)
}

func (d *durationFlag) Duration() time.Duration {
	return d.value
}

func (d *durationFlag) IsSet() bool {
	return d.set
}

func parseFlexibleDuration(raw string) (time.Duration, error) {
	matches := durationPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if matches == nil {
		return 0, fmt.Errorf("invalid duration %q; use values like 30m, 2h, 1d, 15 minutes, 2 hours, or 1 day", raw)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", raw, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("duration %q must be greater than zero", raw)
	}

	unit := strings.ToLower(matches[2])
	switch unit {
	case "m", "min", "mins", "minute", "minutes":
		return time.Duration(value) * time.Minute, nil
	case "h", "hr", "hrs", "hour", "hours":
		return time.Duration(value) * time.Hour, nil
	case "d", "day", "days":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration unit %q; supported units are minutes, hours, and days", unit)
	}
}

func formatFlexibleDuration(value time.Duration) string {
	switch {
	case value%(24*time.Hour) == 0:
		return fmt.Sprintf("%dd", int(value/(24*time.Hour)))
	case value%time.Hour == 0:
		return fmt.Sprintf("%dh", int(value/time.Hour))
	case value%time.Minute == 0:
		return fmt.Sprintf("%dm", int(value/time.Minute))
	default:
		return value.String()
	}
}
