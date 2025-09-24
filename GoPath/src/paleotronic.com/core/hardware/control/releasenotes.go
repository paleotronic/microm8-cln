package control

import (
	"encoding/json"
	"io/ioutil"
	"paleotronic.com/core/interfaces"
	"paleotronic.com/files"
	"paleotronic.com/update"
)

var ReleaseNotes = `
==New in this version

* Fixes
  - MacOS intel - fix crash due to buggy purego cgo bridge bindings.
  - Moved to the oto library for audio (deprecating portaudio)
  - Ensure requested sample rate for audio is met.

* MCP (Model Context Protocol) Server Integration added HTTP streaming for local clients
  - Enables AI assistants and LLMs to interact with microM8
  - Compatible with LM Studio
  - Two transport modes: stdio (default) and SSE (Server-Sent Events)

* How to Enable MCP Server via http-streaming
  
  HTTP Streaming Mode (for web-based clients):
    ./microM8 -mcp -mcp-mode http-streaming
    ./microM8 -mcp -mcp-mode http-streaming -mcp-port 8080
    - HTTP-based streaming transport
    - Default port: 1983
    - Includes CORS support
    - Health check endpoint at /mcp/health

* New MCP Tools
  - emulator_state: provides cpu, disk and screen state for non-vision enabled models.
`

type LastReleaseNotes struct {
	Version string `json:"version"`
}

func CheckNewReleaseNotes(ent interfaces.Interpretable, force bool) {
	var lastDate = "000000000000"
	var lastVerFile = files.GetUserDirectory(files.BASEDIR) + "/.lastRelease"
	if data, err := ioutil.ReadFile(lastVerFile); err == nil {
		var lrn LastReleaseNotes
		if err = json.Unmarshal(data, &lrn); err == nil {
			lastDate = lrn.Version
		}
	}

	if lastDate < update.GetBuildNumber() || force {
		var lrn LastReleaseNotes
		lrn.Version = update.GetBuildNumber()
		if data, err := json.Marshal(&lrn); err == nil {
			ioutil.WriteFile(lastVerFile, data, 0755)
		}
		if len(ReleaseNotes) > 0 {
			// Display notes
			hc := NewHelpControllerString(ent, "Release Notes", ReleaseNotes+"\n\nPress Ctrl+Shift+Q or ESC to continue...")
			hc.Do(ent)
		}
	}
}
