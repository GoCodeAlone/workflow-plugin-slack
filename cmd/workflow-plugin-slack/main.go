// Command workflow-plugin-slack is a workflow engine external plugin that
// provides Slack messaging, file uploads, and real-time event triggers via
// Socket Mode. It runs as a subprocess and communicates with the host workflow
// engine via the go-plugin protocol.
package main

import (
	"github.com/GoCodeAlone/workflow-plugin-slack/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func main() {
	sdk.Serve(internal.New())
}
