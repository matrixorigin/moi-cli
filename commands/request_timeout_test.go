package commands

import (
	"strings"
	"testing"
	"time"
)

func TestRequestTimeout(t *testing.T) {
	for _, test := range []struct {
		command string
		args    []string
		want    time.Duration
	}{
		{"file", []string{"analyze-image", "--file", "image.jpg"}, runtimeImageAnalysisCLITimeout},
		{"catalog", []string{"analyze-runtime-image"}, runtimeImageAnalysisCLITimeout},
		{"file", []string{"download"}, 30 * time.Second},
		{"file", []string{"upload"}, 30 * time.Second},
		{"catalog", []string{"list"}, 30 * time.Second},
		{"workflow", []string{"run"}, 30 * time.Second},
		{"file", nil, 30 * time.Second},
	} {
		t.Run(test.command+"/"+strings.Join(test.args, " "), func(t *testing.T) {
			if got := RequestTimeout(test.command, test.args); got != test.want {
				t.Fatalf("request timeout = %s, want %s", got, test.want)
			}
		})
	}
}
