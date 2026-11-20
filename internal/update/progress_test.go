package update

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestProgressTracksActualWritesAndPreservesErrors(t *testing.T) {
	var display, data bytes.Buffer
	p := &progressDisplay{out: &display}
	w := p.writer(&data, "astra", 100)
	if _, err := w.Write(bytes.Repeat([]byte("x"), 25)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(display.String(), "25%") {
		t.Fatal(display.String())
	}
	if _, err := w.Write(bytes.Repeat([]byte("x"), 75)); err != nil {
		t.Fatal(err)
	}
	p.clear()
	if data.Len() != 100 || !strings.Contains(display.String(), "100%") || !strings.HasSuffix(display.String(), "\r\x1b[2K") {
		t.Fatal(display.String())
	}
	w = p.writer(errorWriter{}, "astra", 100)
	if n, err := w.Write([]byte("data")); n != 0 || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("write failure lost: %d %v", n, err)
	}
	var silent *progressDisplay
	if silent.writer(&data, "astra", 100) != &data {
		t.Fatal("silent path wrapped writer")
	}
	for _, key := range []string{"MOI_AGENT_CALL", "CI"} {
		t.Setenv(key, "1")
		if progressEnabled() {
			t.Fatal("machine progress enabled")
		}
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestInstallationProgressIncludesVerificationAndActivation(t *testing.T) {
	f := newFixture(t)
	f.publish("progress-test", "1.0.0", 1)
	var display bytes.Buffer
	f.m.progress = &progressDisplay{out: &display}
	result, err := f.m.Apply(t.Context(), false)
	if err != nil || result.Status != "updated" {
		t.Fatalf("%+v %v", result, err)
	}
	for _, stage := range []string{"Downloading astra", "Downloading astra-edge", "Downloading moi-cli", "100%", "Verifying moi-cli", "Extracting moi-cli", "Checking binaries and Skill", "Activating release"} {
		if !strings.Contains(display.String(), stage) {
			t.Errorf("missing %s", stage)
		}
	}
}
