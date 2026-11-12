package main

import (
	"testing"
	"time"
)

func TestParseObserverFlags(t *testing.T) {
	cfg, err := parseArgs([]string{"orion", "--lat", "35.68", "--lon=139.77", "--time", "2026-01-15T21:00:00+09:00"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.observe || cfg.lat != 35.68 || cfg.lon != 139.77 {
		t.Errorf("observer: %+v", cfg)
	}
	want := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	if !cfg.at.Equal(want) {
		t.Errorf("time: %v, want %v", cfg.at, want)
	}

	cfg, err = parseArgs([]string{"--visible", "--lat", "-33.9", "--lon", "18.4", "--time", "2026-06-01 22:30"})
	if err != nil || !cfg.visible || cfg.at.Hour() != 22 || cfg.at.Location() != time.Local {
		t.Errorf("local time: %+v %v", cfg, err)
	}

	cfg, err = parseArgs([]string{"orion", "--lat", "10", "--lon", "20"})
	if err != nil || cfg.at.IsZero() {
		t.Errorf("default time is now: %+v %v", cfg, err)
	}

	for _, args := range [][]string{
		{"orion", "--lat", "35"},
		{"orion", "--lon", "135"},
		{"--visible"},
		{"orion", "--time", "2026-01-01T00:00"},
		{"orion", "--lat", "91", "--lon", "0"},
		{"orion", "--lat", "0", "--lon", "-181"},
		{"orion", "--lat", "x", "--lon", "0"},
		{"orion", "--lat", "0", "--lon", "0", "--time", "tomorrow"},
		{"orion", "--lat"},
		{"orion", "--lat", "0", "--lon", "0", "--time", "1600-01-01T00:00Z"},
		{"orion", "--lat", "0", "--lon", "0", "--time", "2101-01-01T00:00"},
	} {
		if _, err := parseArgs(args); err == nil {
			t.Errorf("parseArgs(%q) should fail", args)
		}
	}
}
