package commands

import "time"

// RequestTimeout keeps the client and authentication transport within the
// command's existing request budget, instead of truncating image analysis.
func RequestTimeout(command string, args []string) time.Duration {
	if len(args) > 0 && ((command == "file" && args[0] == "analyze-image") ||
		(command == "catalog" && args[0] == "analyze-runtime-image")) {
		return runtimeImageAnalysisCLITimeout
	}
	if command == "workflow" && len(args) > 0 {
		switch args[0] {
		case "test-candidate", "test-custom-operator":
			return 10 * time.Minute
		}
	}
	return 30 * time.Second
}
