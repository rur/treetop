package internal

import (
	"strconv"
	"testing"
	"time"
)

func TestScript(t *testing.T) {
	if len(ScriptContent) == 0 {
		t.Error("Script Content is empty")
	}
}

func TestModifiedTime(t *testing.T) {
	ts, err := strconv.ParseInt(Modified, 10, 64)
	if err != nil {
		t.Errorf("Expecting unix epoch timestamp, error parsing integer: %s", err)
	}
	modTime := time.Unix(ts, 0)
	if err != nil {
		t.Errorf("Error parsing modified unix TS into a date: %s", err)
	}
	base, err := time.Parse(time.RFC3339Nano, "2021-01-14T04:24:14.554197000+00:00")
	if err != nil {
		t.Fatal(err)
	}
	if base.After(modTime) {
		t.Errorf("Expected time to be after the base time %s, got %s", base, modTime)
	}
}
