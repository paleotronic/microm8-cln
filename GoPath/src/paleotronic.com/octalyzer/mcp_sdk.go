package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"paleotronic.com/core/hardware/apple2helpers"
	"paleotronic.com/core/hardware/servicebus"
	"paleotronic.com/core/settings"
	"paleotronic.com/core/types"
	"paleotronic.com/debugger"
	"paleotronic.com/debugger/debugtypes"
	"paleotronic.com/files"
	"paleotronic.com/log"
	"paleotronic.com/octalyzer/backend"
	"paleotronic.com/octalyzer/tokenizer"
	"paleotronic.com/runestring"
	"paleotronic.com/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SSE connection tracking
var (
	activeConnections int64
	serverStartTime   = time.Now()
)

// SSE configuration
const (
	heartbeatInterval = 30 * time.Second  // Send heartbeat every 30 seconds
)

// Tool parameter structs
type RebootParams struct{}

type PauseParams struct{}

type BreakParams struct{}

type InsertDiskParams struct {
	Drive    int    `json:"drive" jsonschema:"description:Drive number (0-2),minimum:0,maximum:2"`
	Filename string `json:"filename" jsonschema:"description:Path to disk image file"`
}

type EjectDiskParams struct {
	Drive int `json:"drive" jsonschema:"description:Drive number (0-2),minimum:0,maximum:2"`
}

type KeyEventParams struct {
	Key       int `json:"key" jsonschema:"description:Key code"`
	Action    int `json:"action" jsonschema:"description:Action (0=release; 1=press; 2=repeat)"`
	Modifiers int `json:"modifiers,omitempty" jsonschema:"description:Modifier keys"`
}

type ScreenshotParams struct {
	Path string `json:"path" jsonschema:"description:Path to save screenshot"`
}

type GetTextScreenParams struct{}

type ReadMemoryParams struct {
	Address int `json:"address" jsonschema:"description:Memory address,minimum:0,maximum:65535"`
}

type WriteMemoryParams struct {
	Address int `json:"address" jsonschema:"description:Memory address,minimum:0,maximum:65535"`
	Value   int `json:"value" jsonschema:"description:Value to write,minimum:0,maximum:255"`
}

type WriteMemoryRangeParams struct {
	Address int   `json:"address" jsonschema:"description:Starting memory address,minimum:0,maximum:65535"`
	Values  []int `json:"values" jsonschema:"description:Array of values to write (each 0-255)"`
}

type SetCPUSpeedParams struct {
	Speed float64 `json:"speed" jsonschema:"description:Speed multiplier (0.25 to 4.0),minimum:0.25,maximum:4.0"`
}

type TypeTextParams struct {
	Text  string `json:"text" jsonschema:"description:Text to type"`
	Delay int    `json:"delay,omitempty" jsonschema:"description:Delay between keystrokes in milliseconds (default: 50),minimum:10,maximum:1000"`
}

type AssembleParams struct {
	Code string `json:"code" jsonschema:"description:6502 assembly code to assemble"`
}

type ApplesoftReadParams struct{}

type ApplesoftWriteParams struct {
	Code string `json:"code" jsonschema:"description:Applesoft BASIC source code to tokenize and write to memory"`
}

type ListFilesParams struct {
	Path string `json:"path,omitempty" jsonschema:"description:Path to list files from (default: /)"`
}

type ReadFileParams struct {
	Path string `json:"path" jsonschema:"description:Path to the file to read"`
}

type InsertDiskFileParams struct {
	Drive    int    `json:"drive" jsonschema:"description:Drive number (0-2),minimum:0,maximum:2"`
	Filepath string `json:"filepath" jsonschema:"description:Path to disk image file in the virtual file system"`
}

type LoadBasicProgramParams struct {
	Path    string `json:"path" jsonschema:"description:Path to BASIC program file in the virtual file system,required"`
	Dialect string `json:"dialect,omitempty" jsonschema:"description:BASIC dialect (fp for Applesoft, int for Integer, logo for Logo). Auto-detected if not specified,enum:[fp,int,logo]"`
	AutoRun bool   `json:"auto_run,omitempty" jsonschema:"description:Whether to run the program immediately after loading,default:true"`
}

type ReadInterpreterCodeParams struct {
	Dialect string `json:"dialect,omitempty" jsonschema:"description:Interpreter dialect to read from (fp for Applesoft/microBASIC, int for Integer BASIC, logo for Logo). Uses current if not specified,enum:[fp,int,logo]"`
	Procedure string `json:"procedure,omitempty" jsonschema:"description:Logo procedure name to read (Logo only)"`
}

type WriteInterpreterCodeParams struct {
	Code    string `json:"code" jsonschema:"description:The code to write to the interpreter,required"`
	Dialect string `json:"dialect,omitempty" jsonschema:"description:Interpreter dialect to write to (fp for Applesoft/microBASIC, int for Integer BASIC, logo for Logo). Uses current if not specified,enum:[fp,int,logo]"`
	Replace bool   `json:"replace,omitempty" jsonschema:"description:Whether to replace existing code (true) or append (false),default:true"`
}

type ListAppleIITreeParams struct{}

// Gaming-related parameter structs
type GetGraphicsScreenParams struct {
}

type JoystickEventParams struct {
	Controller int  `json:"controller,omitempty" jsonschema:"description:Controller number (0 or 1),minimum:0,maximum:1,default:0"`
	Type       string `json:"type,omitempty" jsonschema:"description:Controller type,enum:[joystick,paddle],default:joystick"`
	X          int    `json:"x,omitempty" jsonschema:"description:X position (-127 to 127 for joystick; 0-255 for paddle),minimum:-127,maximum:255"`
	Y          int    `json:"y,omitempty" jsonschema:"description:Y position (-127 to 127 for joystick only),minimum:-127,maximum:127"`
	Button0    bool   `json:"button0,omitempty" jsonschema:"description:Button 0 state"`
	Button1    bool   `json:"button1,omitempty" jsonschema:"description:Button 1 state"`
	Button2    bool   `json:"button2,omitempty" jsonschema:"description:Button 2 state (joystick only)"`
}

type GetGameStateParams struct {
	Game     string `json:"game" jsonschema:"description:Game identifier (e.g. mspacman),enum:[mspacman,pacman,donkeykong,custom]"`
	Detailed bool   `json:"detailed,omitempty" jsonschema:"description:Include detailed game-specific data"`
}

type GetCurrentGraphicsDataParams struct {
	X int `json:"x,omitempty" jsonschema:"description:X coordinate of rectangle (default: 0),minimum:0"`
	Y int `json:"y,omitempty" jsonschema:"description:Y coordinate of rectangle (default: 0),minimum:0"`
	W int `json:"w,omitempty" jsonschema:"description:Width of rectangle (default: full width),minimum:1"`
	H int `json:"h,omitempty" jsonschema:"description:Height of rectangle (default: full height),minimum:1"`
}

// Debugger parameter structs
type DebugCPUControlParams struct {
	Action string `json:"action" jsonschema:"description:CPU control action (pause/continue/step/step-over/step-out),enum:[pause,continue,step,step-over,step-out]"`
}

type DebugBreakpointAddParams struct {
	Address *int    `json:"address,omitempty" jsonschema:"description:Breakpoint address (PC)"`
	Type    string  `json:"type,omitempty" jsonschema:"description:Breakpoint type,enum:[address,register,memory-read,memory-write]"`
	Condition string `json:"condition,omitempty" jsonschema:"description:Breakpoint condition expression"`
}

type DebugBreakpointRemoveParams struct {
	Index int `json:"index" jsonschema:"description:Breakpoint index to remove"`
}

type DebugBreakpointListParams struct{}

type DebugMemoryReadParams struct {
	Address int `json:"address" jsonschema:"description:Memory address to read from,minimum:0,maximum:65535"`
	Count   int `json:"count" jsonschema:"description:Number of bytes to read,minimum:1,maximum:256"`
}

type DebugMemoryWriteParams struct {
	Address int   `json:"address" jsonschema:"description:Memory address to write to,minimum:0,maximum:65535"`
	Data    []int `json:"data" jsonschema:"description:Bytes to write (each 0-255)"`
}

type DebugRegisterGetParams struct{}

type DebugRegisterSetParams struct {
	Register string `json:"register" jsonschema:"description:Register name (a/x/y/sp/pc/p),enum:[a,x,y,sp,pc,p]"`
	Value    int    `json:"value" jsonschema:"description:Value to set,minimum:0,maximum:65535"`
}

type DebugInstructionTraceParams struct {
	Backlog   int `json:"backlog,omitempty" jsonschema:"description:Number of previous instructions to show (default: 10),minimum:0,maximum:100"`
	Lookahead int `json:"lookahead,omitempty" jsonschema:"description:Number of future instructions to show (default: 10),minimum:0,maximum:100"`
}

type DisassembleParams struct {
	Address int `json:"address" jsonschema:"description:Starting memory address to disassemble from,minimum:0,maximum:65535"`
	Count   int `json:"count,omitempty" jsonschema:"description:Number of instructions to disassemble (default: 20),minimum:1,maximum:100"`
}

type GetMountedDisksParams struct{}

type EmulatorStateParams struct{}

type EnableLiveRewindParams struct{}

type RewindBackParams struct {
	Milliseconds int `json:"milliseconds,omitempty" jsonschema:"description:Number of milliseconds to rewind (default: 5000),minimum:100,maximum:60000"`
}

type StartFileRecordingParams struct {
	FullCPU bool `json:"full_cpu,omitempty" jsonschema:"description:Enable full CPU recording (default: uses global setting)"`
}

type StopRecordingParams struct{}

// Assembler API types
type AssemblerFile struct {
	Name   string `json:"name"`
	Data   string `json:"data"`
	Binary bool   `json:"binary"`
}

type AssemblerRequest struct {
	Files []AssemblerFile `json:"files"`
}

type MerlinError struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
}

type StructuredASMResponse struct {
	Name    string         `json:"name"`
	Address int            `json:"addr"`
	Data    []byte         `json:"data"`
	Disk    bool           `json:"disk"`
	Err     []*MerlinError `json:"err"`
}

// Helper function to ensure debugger is initialized
func ensureDebuggerStarted() error {
	if debugger.DebuggerInstance == nil {
		debugger.Start()
		if debugger.DebuggerInstance == nil {
			return fmt.Errorf("failed to initialize debugger")
		}
	}

	// Check if debugger is already attached using the same logic as keymap.go
	if !debugger.DebuggerInstance.IsAttached() {
		// Attach to the selected slot
		settings.DebuggerAttachSlot = SelectedIndex + 1
		debugger.DebuggerInstance.AttachSlot(SelectedIndex)
		// Open the debugger web interface
		utils.OpenURL(fmt.Sprintf("http://localhost:%d/?attach=%d", settings.DebuggerPort, settings.DebuggerAttachSlot))
	}

	return nil
}

// Tool handler functions
func handleReboot(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[RebootParams]) (*mcp.CallToolResultFor[any], error) {
	backend.ProducerMain.RestartSlot(SelectedIndex)
	// settings.IntSetSlotRestart(SelectedIndex, true)
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: "Emulator rebooted"}},
	}, nil
}

func handlePause(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[PauseParams]) (*mcp.CallToolResultFor[any], error) {
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)

	if e.IsWaitingForWorld() {
		e.ResumeTheWorld()
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "Emulator resumed"}},
		}, nil
	} else {
		e.StopTheWorld()
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "Emulator paused"}},
		}, nil
	}
}

func handleBreak(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[BreakParams]) (*mcp.CallToolResultFor[any], error) {
	// Send Ctrl+C key press (keycode 3)
	syncWindow <- SyncWindowRequest{
		KeyEvent: &KeyRequest{
			Key:    3,  // Ctrl+C keycode
			Action: 1,  // press
		},
	}

	// Wait 50ms
	time.Sleep(50 * time.Millisecond)

	// Send key release
	syncWindow <- SyncWindowRequest{
		KeyEvent: &KeyRequest{
			Key:    3,  // Ctrl+C keycode
			Action: 0,  // release
		},
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: "Break signal (Ctrl+C) sent"}},
	}, nil
}

func handleInsertDisk(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[InsertDiskParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments

	// Read the disk file
	diskBytes, err := os.ReadFile(args.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read disk file: %w", err)
	}

	switch args.Drive {
	case 0:
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.DiskIIInsertBytes,
			servicebus.DiskTargetBytes{
				Filename: args.Filename,
				Drive:    args.Drive,
				Bytes:    diskBytes,
			},
		)
		settings.PureBootVolume[SelectedIndex] = "local:"+args.Filename
	case 1:
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.DiskIIInsertBytes,
			servicebus.DiskTargetBytes{
				Filename: args.Filename,
				Drive:    args.Drive,
				Bytes:    diskBytes,
			},
		)
		settings.PureBootVolume2[SelectedIndex] = "local:"+args.Filename
	case 2:
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.SmartPortInsertBytes,
			servicebus.DiskTargetBytes{
				Filename: args.Filename,
				Drive:    args.Drive,
				Bytes:    diskBytes,
			},
		)
	}

	// backend.ProducerMain.VMMediaChange(SelectedIndex, args.Drive, args.Filename)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Disk inserted in drive %d: %s", args.Drive, args.Filename),
		}},
	}, nil
}

func handleEjectDisk(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[EjectDiskParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	switch args.Drive {
	case 0, 1:
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.DiskIIEject,
			args.Drive,
		)
	case 2:
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.SmartPortEject,
			0,
		)
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Disk ejected from drive %d", args.Drive),
		}},
	}, nil
}

func handleKeyEvent(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[KeyEventParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	if args.Key == 13 {
		if args.Action == 1 {
			e := backend.ProducerMain.GetInterpreter(SelectedIndex)
			e.SetPasteBuffer(runestring.Cast(e.GetPasteBuffer().String() + string(rune(args.Key))))
		}
	} else {
		syncWindow <- SyncWindowRequest{
			KeyEvent: &KeyRequest{
				Key:       args.Key,
				Action:    args.Action,
				Modifiers: args.Modifiers,
			},
		}
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Key event sent: key=%d action=%d", args.Key, args.Action),
		}},
	}, nil
}

func handleScreenshot(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ScreenshotParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	settings.ScreenShotNeeded = true
	for settings.ScreenShotNeeded {
		time.Sleep(1 * time.Millisecond)
	}
	
	if err := os.WriteFile(args.Path, settings.ScreenShotJPEGData, 0644); err != nil {
		return nil, fmt.Errorf("failed to save screenshot: %w", err)
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Screenshot saved to %s", args.Path),
		}},
	}, nil
}

func handleGetTextScreen(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetTextScreenParams]) (*mcp.CallToolResultFor[any], error) {
	// Get the interpreter
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)

	// Get active text modes to determine if text layers are available
	activeTextModes := apple2helpers.GetActiveTextModes(e)

	// Check if TEXT or TXT2 is active
	var hasTextMode bool
	for _, mode := range activeTextModes {
		if mode == "TEXT" || mode == "TXT2" {
			hasTextMode = true
			break
		}
	}

	if !hasTextMode {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "No active TEXT or TXT2 mode"}},
		}, nil
	}

	// Try to get text screen content from TEXT or TXT2 layers
	var textScreen =  "TEXT or TXT2 layer active but no content available"
	for _, l := range HUDLayers[SelectedIndex] {
		if l == nil || !l.Spec.GetActive() {
			continue
		}

		layerID := l.Spec.GetID()

		// Only read from TEXT or TXT2 layers, ignore MONI and OOSD
		if ll, ok := e.GetHUDLayerByID(layerID); ok {
			if layerID == "TEXT" || layerID == "TXT2" {
				text := strings.Join(ll.Control.GetStrings(), "")
				if text != "" {
					textScreen = strings.ReplaceAll(text, "\n\n", "\n")
					break
				}
			}
		}
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: textScreen}},
	}, nil
}

func handleReadMemory(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ReadMemoryParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	value := int(e.GetMemory(args.Address)) & 0xff
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Memory at $%04X: $%02X (%d)", args.Address, value, value),
		}},
	}, nil
}

func handleWriteMemory(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[WriteMemoryParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	e.SetMemory(args.Address, uint64(args.Value))
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Memory at $%04X set to $%02X (%d)", args.Address, args.Value, args.Value),
		}},
	}, nil
}

func handleWriteMemoryRange(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[WriteMemoryRangeParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	// Validate address range
	if args.Address < 0 || args.Address > 65535 {
		return nil, fmt.Errorf("invalid address: must be 0-65535")
	}
	
	// Validate that address + length doesn't overflow
	if args.Address + len(args.Values) > 65536 {
		return nil, fmt.Errorf("memory range exceeds address space")
	}
	
	// Validate values are in valid byte range
	for i, value := range args.Values {
		if value < 0 || value > 255 {
			return nil, fmt.Errorf("invalid value at index %d: must be 0-255", i)
		}
	}
	
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Write each byte to memory
	for i, value := range args.Values {
		e.SetMemory(args.Address + i, uint64(value))
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Wrote %d bytes to memory starting at $%04X", len(args.Values), args.Address),
		}},
	}, nil
}

func handleSetCPUSpeed(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[SetCPUSpeedParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	cpu := apple2helpers.GetCPU(e)
	cpu.SetWarpUser(args.Speed)
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("CPU speed set to %.2fx", args.Speed),
		}},
	}, nil
}

func handleTypeText(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[TypeTextParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments

	// Default delay is 50ms
	delay := args.Delay
	if delay == 0 {
		delay = 50
	}

	e := backend.ProducerMain.GetInterpreter(SelectedIndex)

	s := time.Now()
	for e.GetPasteBuffer().String() != "" && time.Since(s) < 30 * time.Second {
		time.Sleep(250 * time.Millisecond)
	}

	e.SetPasteBuffer(runestring.Cast(args.Text))

// 	// Type the text synchronously (not in a goroutine)
// 	for _, char := range args.Text {
// 		// Convert rune to ASCII code
// 		keyCode := int(char)
//
// 		// Send key press
// 		syncWindow <- SyncWindowRequest{
// 			KeyEvent: &KeyRequest{
// 				Key:    keyCode,
// 				Action: 1, // press
// 			},
// 		}
//
// 		// Small delay for key press to register
// 		time.Sleep(10 * time.Millisecond)
//
// 		// Send key release
// 		syncWindow <- SyncWindowRequest{
// 			KeyEvent: &KeyRequest{
// 				Key:    keyCode,
// 				Action: 0, // release
// 			},
// 		}
//
// 		// Delay between characters
// 		time.Sleep(time.Duration(delay) * time.Millisecond)
// 	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Typing %d characters with %dms delay between keystrokes", len(args.Text), delay),
		}},
	}, nil
}

func handleAssemble(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[AssembleParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments

	// Base64 encode the assembly code
	encodedCode := base64.StdEncoding.EncodeToString([]byte(args.Code))

	makefile := `
main = "main.s"
	`
	encodedMake := base64.StdEncoding.EncodeToString([]byte(makefile))

	// Prepare the request
	req := AssemblerRequest{
		Files: []AssemblerFile{
			{
				Name:   "main.s",
				Data:   encodedCode,
				Binary: false,
			},
			{
				Name:   "makefile",
				Data:   encodedMake,
				Binary: false,
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP client that ignores certificate verification
	// WARNING: This is insecure and should only be used temporarily
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Make the HTTP request
	resp, err := client.Post("https://turtlespaces.org:6502/api/v1/asm/multifile",
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call assembler API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Now parse the actual response
	var asmResp StructuredASMResponse
	if err := json.Unmarshal(body, &asmResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assembler response: %w", err)
	}

	// Check for errors
	if len(asmResp.Err) > 0 {
		var errorMsg string
		for _, e := range asmResp.Err {
			errorMsg += fmt.Sprintf("Line %d in %s: %s\n", e.Line, e.Filename, e.Message)
		}
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Assembly error(s):\n%s", errorMsg),
			}},
		}, nil
	}

	// Success - the data is already in byte array format
	if len(asmResp.Data) == 0 || asmResp.Address == 0 {
		return nil, fmt.Errorf("unexpected response format: missing data or address")
	}

	// Write the assembled bytes to memory
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	for i, b := range asmResp.Data {
		e.SetMemory(asmResp.Address+i, uint64(b))
	}

	// Prepare success message
	successMsg := fmt.Sprintf("%s: %d bytes assembled to memory address $%04X",
		asmResp.Name, len(asmResp.Data), asmResp.Address)

	// If less than 32 bytes, show hex dump
	if len(asmResp.Data) < 32 {
		var hexBytes string
		for i, b := range asmResp.Data {
			if i > 0 {
				hexBytes += " "
			}
			hexBytes += fmt.Sprintf("%02X", b)
		}
		successMsg += fmt.Sprintf("\n\nHex bytes: %s", hexBytes)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: successMsg}},
	}, nil
}

func handleApplesoftRead(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ApplesoftReadParams]) (*mcp.CallToolResultFor[any], error) {
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)

	// Read Applesoft BASIC program from memory
	base := 0x801
	end := int(e.GetMemory(175)) + 256*int(e.GetMemory(176))
	length := end - base

	if length <= 0 || length > 0x8000 { // Sanity check - 32KB max
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: "No Applesoft BASIC program in memory",
			}},
		}, nil
	}

	// Read the program bytes
	var data = make([]byte, length)
	for i := range data {
		data[i] = byte(e.GetMemory(base + i))
	}

	// Detokenize the BASIC program
	code := tokenizer.ApplesoftDetoks(data)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: string(code),
		}},
	}, nil
}

func handleApplesoftWrite(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ApplesoftWriteParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments

	// Split the code into lines
	lines := strings.Split(args.Code, "\n")

	// Tokenize the BASIC program
	tokenized := tokenizer.ApplesoftTokenize(lines)

	// Validate tokenized data isn't too large
	if len(tokenized) > 0x7FFF { // ~32KB max
		return nil, fmt.Errorf("tokenized program too large (%d bytes)", len(tokenized))
	}

	// Write tokenized program to memory starting at 0x801
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	base := 0x801

	// Set the BASIC program start pointer
	e.SetMemory(103, uint64(base & 0xFF))      // 0x67 - TXTTAB low byte
	e.SetMemory(104, uint64((base >> 8) & 0xFF)) // 0x68 - TXTTAB high byte

	// Write the tokenized program
	for i, b := range tokenized {
		e.SetMemory(base + i, uint64(b))
	}

	// Calculate end address
	end := base + len(tokenized)
	arr := end  // Array storage starts at program end
	ft := 0x9600  // Free space pointer

	// Set various pointers that Applesoft BASIC uses
	e.SetMemory(74, uint64(end & 0xFF))       // 0x4A - VARTAB low byte (start of simple variables)
	e.SetMemory(75, uint64((end >> 8) & 0xFF)) // 0x4B - VARTAB high byte

	e.SetMemory(105, uint64(end & 0xFF))      // 0x69 - PRGEND low byte
	e.SetMemory(106, uint64((end >> 8) & 0xFF)) // 0x6A - PRGEND high byte

	e.SetMemory(175, uint64(end & 0xFF))      // 0xAF - VARTAB duplicate low byte
	e.SetMemory(176, uint64((end >> 8) & 0xFF)) // 0xB0 - VARTAB duplicate high byte

	e.SetMemory(107, uint64(arr & 0xFF))      // 0x6B - ARYTAB low byte (start of array variables)
	e.SetMemory(108, uint64((arr >> 8) & 0xFF)) // 0x6C - ARYTAB high byte

	e.SetMemory(109, uint64(arr & 0xFF))      // 0x6D - STREND low byte (end of arrays)
	e.SetMemory(110, uint64((arr >> 8) & 0xFF)) // 0x6E - STREND high byte

	e.SetMemory(111, uint64(ft & 0xFF))       // 0x6F - FRETOP low byte (start of string storage)
	e.SetMemory(112, uint64((ft >> 8) & 0xFF)) // 0x70 - FRETOP high byte

	e.SetMemory(115, uint64(ft & 0xFF))       // 0x73 - HIMEM low byte
	e.SetMemory(116, uint64((ft >> 8) & 0xFF)) // 0x74 - HIMEM high byte

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Wrote %d bytes of tokenized BASIC to memory at $%04X", len(tokenized), base),
		}},
	}, nil
}

func handleListFiles(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ListFilesParams]) (*mcp.CallToolResultFor[any], error) {
	// Import needed for files package
	args := params.Arguments
	
	// Default to root if no path specified
	path := args.Path
	if path == "" {
		path = "/"
	}
	
	// Read directory via provider
	dirs, files, err := files.ReadDirViaProvider(path, "*.*")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}
	
	// Build response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Files in %s:\n\n", path))
	
	// Add parent directory if not at root
	if path != "/" && path != "" {
		result.WriteString("../  <parent directory>\n")
	}
	
	// List directories first
	if len(dirs) > 0 {
		result.WriteString("Directories:\n")
		for _, dir := range dirs {
			if dir.Name != ".." { // Skip the .. entry that ReadDirViaProvider adds
				desc := ""
				if dir.Description != "" {
					desc = fmt.Sprintf(" - %s", dir.Description)
				}
				result.WriteString(fmt.Sprintf("  %s/%s\n", dir.Name, desc))
			}
		}
		result.WriteString("\n")
	}
	
	// List files
	if len(files) > 0 {
		result.WriteString("Files:\n")
		for _, file := range files {
			// Format size
			sizeStr := fmt.Sprintf("%d", file.Size)
			if file.Size > 1024*1024 {
				sizeStr = fmt.Sprintf("%.1fM", float64(file.Size)/(1024*1024))
			} else if file.Size > 1024 {
				sizeStr = fmt.Sprintf("%.1fK", float64(file.Size)/1024)
			}
			
			// Build file entry
			entry := fmt.Sprintf("  %s", file.Name)
			if file.Extension != "" {
				entry += "." + file.Extension
			}
			entry += fmt.Sprintf(" (%s)", sizeStr)
			
			if file.Description != "" {
				entry += fmt.Sprintf(" - %s", file.Description)
			}
			
			result.WriteString(entry + "\n")
		}
	}
	
	// Summary
	result.WriteString(fmt.Sprintf("\nTotal: %d directories, %d files\n", len(dirs), len(files)))
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil
}

func handleReadFile(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ReadFileParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	// Parse the path to separate directory and filename
	path := args.Path
	var dir, filename string
	
	// Handle absolute paths starting with /
	if strings.HasPrefix(path, "/") {
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(parts) > 1 {
			dir = "/" + strings.Join(parts[:len(parts)-1], "/")
			filename = parts[len(parts)-1]
		} else {
			dir = "/"
			filename = parts[0]
		}
	} else {
		// Relative path
		lastSlash := strings.LastIndex(path, "/")
		if lastSlash >= 0 {
			dir = path[:lastSlash]
			filename = path[lastSlash+1:]
		} else {
			dir = ""
			filename = path
		}
	}
	
	// Read the file
	fileRecord, err := files.ReadBytesViaProvider(strings.Trim(dir, "/"), filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	
	// Prepare response based on file type
	var content string
	
	// Check if it's likely a text file based on extension
	ext := strings.ToLower(files.GetExt(filename))
	textExts := []string{"txt", "bas", "lst", "asm", "s", "cfg", "ini", "md", "log", "bat", "sh"}
	isText := false
	for _, textExt := range textExts {
		if ext == textExt {
			isText = true
			break
		}
	}
	
	// If it's a text file or small enough, try to display as text
	if isText || len(fileRecord.Content) < 8192 {
		// Try to interpret as text
		textContent := string(fileRecord.Content)
		// Check if it's printable
		isPrintable := true
		for _, r := range textContent {
			if r < 32 && r != '\n' && r != '\r' && r != '\t' {
				isPrintable = false
				break
			}
		}
		
		if isPrintable {
			content = fmt.Sprintf("File: %s\nSize: %d bytes\n\n%s", path, len(fileRecord.Content), textContent)
		} else {
			// Binary file - show hex dump
			content = fmt.Sprintf("File: %s\nSize: %d bytes\nType: Binary\n\n", path, len(fileRecord.Content))
			content += "Hex dump (first 256 bytes):\n"
			
			limit := len(fileRecord.Content)
			if limit > 256 {
				limit = 256
			}
			
			for i := 0; i < limit; i += 16 {
				content += fmt.Sprintf("%04X: ", i)
				
				// Hex bytes
				for j := 0; j < 16 && i+j < limit; j++ {
					content += fmt.Sprintf("%02X ", fileRecord.Content[i+j])
				}
				
				// Padding
				for j := limit - i; j < 16 && i+j >= limit; j++ {
					content += "   "
				}
				
				// ASCII representation
				content += " |"
				for j := 0; j < 16 && i+j < limit; j++ {
					b := fileRecord.Content[i+j]
					if b >= 32 && b < 127 {
						content += string(b)
					} else {
						content += "."
					}
				}
				content += "|\n"
			}
			
			if len(fileRecord.Content) > 256 {
				content += fmt.Sprintf("\n... (%d more bytes)", len(fileRecord.Content)-256)
			}
		}
	} else {
		// Large binary file
		content = fmt.Sprintf("File: %s\nSize: %d bytes\nType: Binary (too large to display)\n", path, len(fileRecord.Content))
		content += "\nFile read successfully but content is too large to display in MCP response."
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: content}},
	}, nil
}

func handleInsertDiskFile(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[InsertDiskFileParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	// Validate drive number
	if args.Drive < 0 || args.Drive > 2 {
		return nil, fmt.Errorf("invalid drive number %d (must be 0-2)", args.Drive)
	}
	
	// Ensure the filepath starts with /
	filepath := args.Filepath
	if !strings.HasPrefix(filepath, "/") {
		filepath = "/" + filepath
	}
	
	// Check if it's a high capacity disk (3.5" or hard disk image)
	// First, we need to check if the file exists and get its size
	dir := files.GetPath(filepath)
	filename := files.GetFilename(filepath)
	
	// Try to get file info by listing the directory
	dirs, fileList, err := files.ReadDirViaProvider(dir, "*.*")
	if err != nil {
		return nil, fmt.Errorf("failed to access directory %s: %w", dir, err)
	}
	
	// Find the file in the list
	var fileSize int64
	var fileExt string
	found := false
	
	for _, f := range fileList {
		fullName := f.Name
		if f.Extension != "" {
			fullName += "." + f.Extension
		}
		if strings.EqualFold(fullName, filename) {
			fileSize = f.Size
			fileExt = f.Extension
			found = true
			break
		}
	}
	
	if !found {
		// Check directories (in case it's a disk image inside a directory)
		for _, d := range dirs {
			if strings.EqualFold(d.Name, filename) {
				found = true
				fileExt = files.GetExt(filename)
				break
			}
		}
	}
	
	if !found {
		return nil, fmt.Errorf("file not found: %s", filepath)
	}
	
	// Determine if it's a high capacity disk
	isHighCapacity := files.Apple2IsHighCapacity(fileExt, int(fileSize))
	
	// Send the appropriate service bus message
	if isHighCapacity {
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.SmartPortInsertFilename,
			servicebus.DiskTargetString{
				Drive:    args.Drive,
				Filename: filepath,
			},
		)
	} else {
		servicebus.SendServiceBusMessage(
			SelectedIndex,
			servicebus.DiskIIInsertFilename,
			servicebus.DiskTargetString{
				Drive:    args.Drive,
				Filename: filepath,
			},
		)
	}
	
	// Mount the disk image for file system access
	files.MountDSKImage(files.GetPath(filepath), files.GetFilename(filepath), args.Drive)
	
	diskType := "5.25\" disk"
	if isHighCapacity {
		diskType = "3.5\" disk / hard disk"
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Inserted %s into drive %d: %s", diskType, args.Drive, filepath),
		}},
	}, nil
}

func handleLoadBasicProgram(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[LoadBasicProgramParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	// Ensure the path starts with /
	path := args.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	// Extract directory and filename
	dir := files.GetPath(path)
	filename := files.GetFilename(path)
	
	// Check if file exists
	_, fileList, err := files.ReadDirViaProvider(dir, "*.*")
	if err != nil {
		return nil, fmt.Errorf("failed to access directory %s: %w", dir, err)
	}
	
	// Find the file in the list
	var fileExt string
	found := false
	for _, f := range fileList {
		fullName := f.Name
		if f.Extension != "" {
			fullName += "." + f.Extension
		}
		if strings.EqualFold(fullName, filename) {
			fileExt = f.Extension
			found = true
			break
		}
	}
	
	if !found {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	
	// Auto-detect dialect if not specified
	dialect := args.Dialect
	if dialect == "" {
		// Check file extension to determine dialect
		info, ok := files.GetInfo(fileExt)
		if !ok || info.Dialect == "" {
			// Try to determine from common extensions
			switch strings.ToLower(fileExt) {
			case "bas", "a", "apl", "app":
				dialect = "fp"  // Applesoft BASIC
			case "i", "int":
				dialect = "int" // Integer BASIC
			case "l", "lgo":
				dialect = "logo" // Logo
			default:
				return nil, fmt.Errorf("unable to determine BASIC dialect for file extension: %s", fileExt)
			}
		} else {
			dialect = info.Dialect
		}
	}
	
	// Validate dialect
	if dialect != "fp" && dialect != "int" && dialect != "logo" {
		return nil, fmt.Errorf("invalid dialect: %s (must be fp, int, or logo)", dialect)
	}
	
	// Get the interpreter for the selected slot
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Build the run command
	var runCommand string
	if args.AutoRun {
		// Default to auto-run
		runCommand = fmt.Sprintf("run \"%s\"\r\n", path)
	} else {
		// Just load without running
		switch dialect {
		case "fp", "int":
			runCommand = fmt.Sprintf("load \"%s\"\r\n", path)
		case "logo":
			runCommand = fmt.Sprintf("load \"%s\"\r\n", path)
		}
	}
	
	// Set up the VMLauncherConfig
	cfg := &settings.VMLauncherConfig{
		WorkingDir: dir,
		Dialect:    dialect,
		RunCommand: runCommand,
	}
	
	// Apply the configuration
	settings.VMLaunch[e.GetMemIndex()] = cfg
	
	// Set the appropriate machine spec if needed
	if !strings.HasPrefix(settings.SpecFile[e.GetMemIndex()], "apple2") {
		settings.SpecFile[e.GetMemIndex()] = "apple2e-en.yaml"
	}
	
	// Trigger a restart to load and potentially run the program
	e.GetMemoryMap().IntSetSlotRestart(e.GetMemIndex(), true)
	
	// Prepare result message
	dialectName := map[string]string{
		"fp":   "Applesoft BASIC",
		"int":  "Integer BASIC",
		"logo": "Logo",
	}[dialect]
	
	action := "Loading and running"
	if !args.AutoRun {
		action = "Loading"
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("%s %s program: %s", action, dialectName, path),
		}},
	}, nil
}

func handleReadInterpreterCode(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ReadInterpreterCodeParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	// Get the interpreter for the selected slot
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	if e == nil {
		return nil, fmt.Errorf("no interpreter available")
	}
	
	// Get the current dialect if not specified
	dialect := args.Dialect
	if dialect == "" {
		currentDialect := e.GetDialect()
		if currentDialect != nil {
			// Get the short name from the dialect
			switch currentDialect.GetTitle() {
			case "Applesoft", "microBASIC":
				dialect = "fp"
			case "Integer BASIC":
				dialect = "int"
			case "Logo":
				dialect = "logo"
			default:
				return nil, fmt.Errorf("unknown current dialect: %s", currentDialect.GetTitle())
			}
		} else {
			return nil, fmt.Errorf("no dialect active, please specify dialect parameter")
		}
	}
	
	var codeText string
	
	// Read the code based on dialect
	switch dialect {
	case "logo":
		// For Logo, we can get specific procedures or all workspace
		lines := e.GetDialect().GetWorkspaceBody(false, args.Procedure)
		codeText = strings.Join(lines, "\r\n")
		
	case "fp", "int":
		// For BASIC variants, get the program listing
		algorithm := e.GetCode()
		if algorithm == nil {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: "No program in memory"}},
			}, nil
		}
		
		// Build the listing similar to how EDIT does it
		var lines []string
		lineNum := algorithm.GetLowIndex()
		highIndex := algorithm.GetHighIndex()
		nlen := len(utils.IntToStr(highIndex))
		
		for lineNum != -1 {
			line, _ := algorithm.Get(lineNum)
			lineStr := fmt.Sprintf("%*d  ", nlen, lineNum)
			
			stmtCount := 0
			for _, stmt := range line {
				if stmtCount > 0 {
					lineStr += ":"
				}
				
				// Convert statement tokens to string
				stmtTokens := *stmt.SubList(0, stmt.Size())
				stmtStr := e.TokenListAsString(stmtTokens)
				lineStr += stmtStr
				stmtCount++
			}
			
			lines = append(lines, lineStr)
			lineNum = algorithm.NextAfter(lineNum)
		}
		
		codeText = strings.Join(lines, "\r\n")
		
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
	
	// Return the code
	if codeText == "" {
		codeText = "No code in memory"
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: codeText}},
	}, nil
}

func handleWriteInterpreterCode(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[WriteInterpreterCodeParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	// Get the interpreter for the selected slot
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	if e == nil {
		return nil, fmt.Errorf("no interpreter available")
	}
	
	// Get the current dialect if not specified
	dialect := args.Dialect
	if dialect == "" {
		currentDialect := e.GetDialect()
		if currentDialect != nil {
			// Get the short name from the dialect
			switch currentDialect.GetTitle() {
			case "Applesoft", "microBASIC":
				dialect = "fp"
			case "Integer BASIC":
				dialect = "int"
			case "Logo":
				dialect = "logo"
			default:
				return nil, fmt.Errorf("unknown current dialect: %s", currentDialect.GetTitle())
			}
		} else {
			return nil, fmt.Errorf("no dialect active, please specify dialect parameter")
		}
	}
	
	// Validate that the current interpreter matches the requested dialect
	currentDialect := e.GetDialect()
	if currentDialect != nil {
		currentTitle := currentDialect.GetTitle()
		switch dialect {
		case "fp":
			if currentTitle != "Applesoft" && currentTitle != "microBASIC" {
				return nil, fmt.Errorf("current interpreter is %s, not Applesoft/microBASIC", currentTitle)
			}
		case "int":
			if currentTitle != "Integer BASIC" {
				return nil, fmt.Errorf("current interpreter is %s, not Integer BASIC", currentTitle)
			}
		case "logo":
			if currentTitle != "Logo" {
				return nil, fmt.Errorf("current interpreter is %s, not Logo", currentTitle)
			}
		}
	}
	
	// Handle code replacement or appending
	if args.Replace {
		// Clear existing code
		e.SetCode(types.NewAlgorithm())
	}
	
	// Parse and load the new code
	lines := strings.Split(args.Code, "\n")
	
	// Set skip mem parse flag to avoid triggering memory parse during load
	if currentDialect != nil {
		currentDialect.SetSkipMemParse(true)
	}
	
	// Parse each line
	for _, line := range lines {
		// Clean up the line
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// For Logo, we might need special handling
		if dialect == "logo" {
			// Logo procedures are typically defined with TO/END blocks
			e.Parse(line)
		} else {
			// For BASIC, parse as normal
			// Remove line numbers if they're included (we'll let the parser handle them)
			e.Parse(line)
		}
	}
	
	// Reset skip mem parse flag
	if currentDialect != nil {
		currentDialect.SetSkipMemParse(false)
	}
	
	// Return success message
	action := "Replaced"
	if !args.Replace {
		action = "Appended to"
	}
	
	dialectName := map[string]string{
		"fp":   "Applesoft/microBASIC",
		"int":  "Integer BASIC",
		"logo": "Logo",
	}[dialect]
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("%s %s code (%d lines)", action, dialectName, len(lines)),
		}},
	}, nil
}

func handleListAppleIITree(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ListAppleIITreeParams]) (*mcp.CallToolResultFor[any], error) {
	// Recursive function to collect all folders
	var collectFolders func(path string, depth int) ([]string, error)
	collectFolders = func(path string, depth int) ([]string, error) {
		// Prevent infinite recursion
		if depth > 10 {
			return nil, nil
		}
		
		// Get directories in current path
		dirs, _, err := files.ReadDirViaProvider(path, "*.*")
		if err != nil {
			// Silently skip directories that can't be read
			return nil, nil
		}
		
		var folders []string
		
		// Process each directory
		for _, dir := range dirs {
			// Skip parent directory entry
			if dir.Name == ".." {
				continue
			}
			
			// Skip hidden directories (starting with .)
			if strings.HasPrefix(dir.Name, ".") {
				continue
			}
			
			// Build full path
			fullPath := strings.TrimRight(path, "/") + "/" + dir.Name
			
			// Add this folder to the list
			folders = append(folders, fullPath)
			
			// Recursively collect subdirectories
			subFolders, _ := collectFolders(fullPath, depth+1)
			if subFolders != nil {
				folders = append(folders, subFolders...)
			}
		}
		
		return folders, nil
	}
	
	// Start collection from /appleii/
	startPath := "/appleii"
	allFolders, err := collectFolders(startPath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to traverse directory tree: %w", err)
	}
	
	// Sort the folders for consistent output
	sort.Strings(allFolders)
	
	// Build response
	var result strings.Builder
	result.WriteString("Apple II Directory Tree:\n")
	result.WriteString("========================\n\n")
	
	if len(allFolders) == 0 {
		result.WriteString("No subdirectories found under /appleii/\n")
	} else {
		// Group by depth for better visualization
		currentDepth := 0
		for _, folder := range allFolders {
			// Calculate depth by counting slashes after /appleii
			relPath := strings.TrimPrefix(folder, "/appleii")
			depth := strings.Count(relPath, "/")
			
			// Add spacing between depth levels
			if depth != currentDepth {
				result.WriteString("\n")
				currentDepth = depth
			}
			
			// Add indentation based on depth
			indent := strings.Repeat("  ", depth)
			
			// Extract just the folder name
			parts := strings.Split(strings.TrimRight(folder, "/"), "/")
			folderName := parts[len(parts)-1]
			
			result.WriteString(fmt.Sprintf("%s%s/\n", indent, folderName))
		}
		
		result.WriteString(fmt.Sprintf("\nTotal folders: %d\n", len(allFolders)))
		
		// Also provide a flat list for easy parsing
		result.WriteString("\nFlat list of all paths:\n")
		result.WriteString("-----------------------\n")
		for _, folder := range allFolders {
			result.WriteString(folder + "\n")
		}
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil
}

// Gaming handler functions
func handleGetGraphicsScreen(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetGraphicsScreenParams]) (*mcp.CallToolResultFor[any], error) {
	
	// Enable unified render mode
	settings.UnifiedRender[SelectedIndex] = true
	
	// Wait a bit for the frame to be available
	time.Sleep(50 * time.Millisecond)
	
	// Get the unified render frame
	frame := settings.UnifiedRenderFrame[SelectedIndex]
	if frame == nil {
		return nil, fmt.Errorf("unified render frame not available")
	}
	

	// Encode the image to PNG or JPEG
	var buf bytes.Buffer
	var err error

	// Scale the image if requested
	outputImage := frame
	// if args.Scale > 1 {
	// Simple nearest-neighbor scaling
	bounds := frame.Bounds()
	newWidth := bounds.Dx()
	newHeight := bounds.Dy() * 2
	scaled := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x
			srcY := y / 2
			scaled.Set(x, y, frame.At(srcX, srcY))
		}
	}
	outputImage = scaled
	// }

	// Encode as PNG for better quality
	err = png.Encode(&buf, outputImage)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Convert to base64
	// encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.ImageContent{
			Meta: mcp.Meta{
				"subject": "Apple II Video Screen",
			},
			Data: buf.Bytes(),
			MIMEType: "image/png",
		}},
	}, nil
	

}

// Helper function to get basic color statistics
func getColorStats(img *image.RGBA) map[string]interface{} {
	bounds := img.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()
	colorCounts := make(map[uint32]int)
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Reduce to 8-bit color for counting
			color := ((r >> 11) << 10) | ((g >> 11) << 5) | (b >> 11)
			colorCounts[color]++
		}
	}
	
	// Find dominant colors
	var dominantColors []map[string]interface{}
	for color, count := range colorCounts {
		if count > totalPixels/100 { // More than 1% of pixels
			r := (color >> 10) & 0x1F
			g := (color >> 5) & 0x1F
			b := color & 0x1F
			dominantColors = append(dominantColors, map[string]interface{}{
				"r":         r * 8,  // Scale back to 0-255
				"g":         g * 8,
				"b":         b * 8,
				"pixels":    count,
				"percentage": float64(count) * 100.0 / float64(totalPixels),
			})
		}
	}
	
	return map[string]interface{}{
		"total_pixels":    totalPixels,
		"unique_colors":   len(colorCounts),
		"dominant_colors": dominantColors,
	}
}

func handleJoystickEvent(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[JoystickEventParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	// Validate controller number
	if args.Controller < 0 || args.Controller > 1 {
		return nil, fmt.Errorf("invalid controller number: %d", args.Controller)
	}
	
	// Get the interpreter
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Joystick/paddle values are typically read from these memory locations:
	// $C064-$C067: Paddle/joystick analog inputs
	// $C061-$C063: Button inputs
	
	if args.Type == "joystick" {
		// Map joystick X/Y to paddle values
		// Joystick center is 0,0 ranging to -127,127
		// Paddle values are 0-255 with 128 as center
		
		paddleX := args.X + 128
		paddleY := args.Y + 128
		
		// Clamp values
		if paddleX < 0 {
			paddleX = 0
		} else if paddleX > 255 {
			paddleX = 255
		}
		
		if paddleY < 0 {
			paddleY = 0
		} else if paddleY > 255 {
			paddleY = 255
		}
		
		// Set paddle values (these are read via $C064/C065)
		// Note: This is a simplified approach - real implementation would
		// need to interface with the paddle emulation system
		
		// For now, we'll set some known memory locations that games might check
		// This is game-specific and would need proper paddle emulation
		
		// Button states
		buttonState := 0
		if args.Button0 {
			buttonState |= 0x80 // Button 0 pressed
		}
		if args.Button1 {
			buttonState |= 0x40 // Button 1 pressed
		}
		if args.Button2 {
			buttonState |= 0x20 // Button 2 pressed
		}
		
		// Set button state at $C061-$C063
		if args.Controller == 0 {
			e.SetMemory(0xC061, uint64(buttonState))
		} else {
			e.SetMemory(0xC062, uint64(buttonState))
		}
		
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Joystick %d: X=%d Y=%d Buttons=%02X", 
					args.Controller, args.X, args.Y, buttonState),
			}},
		}, nil
	}
	
	if args.Type == "paddle" {
		// Paddle mode - X is the paddle position (0-255)
		// Set paddle value
		
		// Button states
		buttonState := 0
		if args.Button0 {
			buttonState |= 0x80
		}
		if args.Button1 {
			buttonState |= 0x40
		}
		
		// Set button state
		if args.Controller == 0 {
			e.SetMemory(0xC061, uint64(buttonState))
		} else {
			e.SetMemory(0xC062, uint64(buttonState))
		}
		
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Paddle %d: Position=%d Buttons=%02X", 
					args.Controller, args.X, buttonState),
			}},
		}, nil
	}
	
	return nil, fmt.Errorf("unsupported controller type: %s", args.Type)
}

func handleGetGameState(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetGameStateParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments
	
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Game-specific memory maps
	if args.Game == "mspacman" {
		// Ms. Pac-Man memory locations (these are approximations)
		score := 0
		lives := 0
		level := 0
		
		// Read score (BCD format, typically 3 bytes)
		score = int(e.GetMemory(0x4E00))<<16 | int(e.GetMemory(0x4E01))<<8 | int(e.GetMemory(0x4E02))
		
		// Read lives
		lives = int(e.GetMemory(0x4E14))
		
		// Read level
		level = int(e.GetMemory(0x4E13))
		
		// Check if game is active (simplified check)
		gameActive := lives > 0 && lives < 10
		
		result := map[string]interface{}{
			"game_active": gameActive,
			"score":       score,
			"lives":       lives,
			"level":       level,
		}
		
		if args.Detailed {
			// Add detailed information
			result["pac_position"] = map[string]int{
				"x": int(e.GetMemory(0x4D30)),
				"y": int(e.GetMemory(0x4D31)),
			}
			
			// Ghost states (simplified)
			ghosts := []map[string]int{}
			for i := 0; i < 4; i++ {
				ghost := map[string]int{
					"x":     int(e.GetMemory(0x4D80 + i*8)),
					"y":     int(e.GetMemory(0x4D81 + i*8)),
					"state": int(e.GetMemory(0x4D82 + i*8)),
				}
				ghosts = append(ghosts, ghost)
			}
			result["ghosts"] = ghosts
		}
		
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
		}, nil
	}
	
	// Default response for unknown games
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Game state for '%s' not implemented", args.Game),
		}},
	}, nil
}

func handleGetCurrentGraphicsData(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetCurrentGraphicsDataParams]) (*mcp.CallToolResultFor[any], error) {
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	args := params.Arguments
	
	// Get the current video mode
	videoMode := apple2helpers.GetVideoMode(e)
	
	// Default response for TEXT mode
	if videoMode == "TEXT" {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: "No graphics mode active - currently in TEXT mode",
			}},
		}, nil
	}
	
	// Define mode-specific properties
	var fullWidth, fullHeight int
	var page string
	var palette []map[string]interface{}
	
	switch videoMode {
	case "LOGR":
		// 40x48 Low Resolution Graphics
		fullWidth = 40
		fullHeight = 48
		page = "1"
		
		// Lo-res color palette (16 colors)
		palette = []map[string]interface{}{
			{"index": 0, "name": "Black", "rgb": []int{0, 0, 0}},
			{"index": 1, "name": "Magenta", "rgb": []int{227, 30, 96}},
			{"index": 2, "name": "Dark Blue", "rgb": []int{96, 78, 189}},
			{"index": 3, "name": "Purple", "rgb": []int{255, 68, 253}},
			{"index": 4, "name": "Dark Green", "rgb": []int{0, 163, 96}},
			{"index": 5, "name": "Grey 1", "rgb": []int{156, 156, 156}},
			{"index": 6, "name": "Medium Blue", "rgb": []int{20, 207, 253}},
			{"index": 7, "name": "Light Blue", "rgb": []int{208, 195, 255}},
			{"index": 8, "name": "Brown", "rgb": []int{96, 114, 3}},
			{"index": 9, "name": "Orange", "rgb": []int{255, 106, 60}},
			{"index": 10, "name": "Grey 2", "rgb": []int{156, 156, 156}},
			{"index": 11, "name": "Pink", "rgb": []int{255, 160, 208}},
			{"index": 12, "name": "Light Green", "rgb": []int{20, 245, 60}},
			{"index": 13, "name": "Yellow", "rgb": []int{208, 221, 141}},
			{"index": 14, "name": "Aqua", "rgb": []int{114, 255, 208}},
			{"index": 15, "name": "White", "rgb": []int{255, 255, 255}},
		}
		
	case "DLGR":
		// 80x48 Double Low Resolution Graphics
		fullWidth = 80
		fullHeight = 48
		page = "1"
		
		// Same palette as LOGR
		palette = []map[string]interface{}{
			{"index": 0, "name": "Black", "rgb": []int{0, 0, 0}},
			{"index": 1, "name": "Magenta", "rgb": []int{227, 30, 96}},
			{"index": 2, "name": "Dark Blue", "rgb": []int{96, 78, 189}},
			{"index": 3, "name": "Purple", "rgb": []int{255, 68, 253}},
			{"index": 4, "name": "Dark Green", "rgb": []int{0, 163, 96}},
			{"index": 5, "name": "Grey 1", "rgb": []int{156, 156, 156}},
			{"index": 6, "name": "Medium Blue", "rgb": []int{20, 207, 253}},
			{"index": 7, "name": "Light Blue", "rgb": []int{208, 195, 255}},
			{"index": 8, "name": "Brown", "rgb": []int{96, 114, 3}},
			{"index": 9, "name": "Orange", "rgb": []int{255, 106, 60}},
			{"index": 10, "name": "Grey 2", "rgb": []int{156, 156, 156}},
			{"index": 11, "name": "Pink", "rgb": []int{255, 160, 208}},
			{"index": 12, "name": "Light Green", "rgb": []int{20, 245, 60}},
			{"index": 13, "name": "Yellow", "rgb": []int{208, 221, 141}},
			{"index": 14, "name": "Aqua", "rgb": []int{114, 255, 208}},
			{"index": 15, "name": "White", "rgb": []int{255, 255, 255}},
		}
		
	case "HGR1":
		// 280x192 High Resolution Graphics Page 1
		fullWidth = 280
		fullHeight = 192
		page = "1"
		
		// HGR color palette (6 colors + black/white)
		palette = []map[string]interface{}{
			{"index": 0, "name": "Black", "rgb": []int{0, 0, 0}},
			{"index": 1, "name": "Green", "rgb": []int{20, 245, 60}},
			{"index": 2, "name": "Purple", "rgb": []int{255, 68, 253}},
			{"index": 3, "name": "White", "rgb": []int{255, 255, 255}},
			{"index": 4, "name": "Black", "rgb": []int{0, 0, 0}},
			{"index": 5, "name": "Orange", "rgb": []int{255, 106, 60}},
			{"index": 6, "name": "Blue", "rgb": []int{20, 207, 253}},
			{"index": 7, "name": "White", "rgb": []int{255, 255, 255}},
		}
		
	case "HGR2":
		// 280x192 High Resolution Graphics Page 2
		fullWidth = 280
		fullHeight = 192
		page = "2"
		
		// Same palette as HGR1
		palette = []map[string]interface{}{
			{"index": 0, "name": "Black", "rgb": []int{0, 0, 0}},
			{"index": 1, "name": "Green", "rgb": []int{20, 245, 60}},
			{"index": 2, "name": "Purple", "rgb": []int{255, 68, 253}},
			{"index": 3, "name": "White", "rgb": []int{255, 255, 255}},
			{"index": 4, "name": "Black", "rgb": []int{0, 0, 0}},
			{"index": 5, "name": "Orange", "rgb": []int{255, 106, 60}},
			{"index": 6, "name": "Blue", "rgb": []int{20, 207, 253}},
			{"index": 7, "name": "White", "rgb": []int{255, 255, 255}},
		}
		
	case "DHR1":
		// 560x192 Double High Resolution Graphics Page 1 (or 140x192 in color mode)
		// For simplicity, we'll read it as 140x192 color mode
		fullWidth = 140
		fullHeight = 192
		page = "1"
		
		// DHR uses the same 16-color palette as lo-res
		palette = []map[string]interface{}{
			{"index": 0, "name": "Black", "rgb": []int{0, 0, 0}},
			{"index": 1, "name": "Magenta", "rgb": []int{227, 30, 96}},
			{"index": 2, "name": "Dark Blue", "rgb": []int{96, 78, 189}},
			{"index": 3, "name": "Purple", "rgb": []int{255, 68, 253}},
			{"index": 4, "name": "Dark Green", "rgb": []int{0, 163, 96}},
			{"index": 5, "name": "Grey 1", "rgb": []int{156, 156, 156}},
			{"index": 6, "name": "Medium Blue", "rgb": []int{20, 207, 253}},
			{"index": 7, "name": "Light Blue", "rgb": []int{208, 195, 255}},
			{"index": 8, "name": "Brown", "rgb": []int{96, 114, 3}},
			{"index": 9, "name": "Orange", "rgb": []int{255, 106, 60}},
			{"index": 10, "name": "Grey 2", "rgb": []int{156, 156, 156}},
			{"index": 11, "name": "Pink", "rgb": []int{255, 160, 208}},
			{"index": 12, "name": "Light Green", "rgb": []int{20, 245, 60}},
			{"index": 13, "name": "Yellow", "rgb": []int{208, 221, 141}},
			{"index": 14, "name": "Aqua", "rgb": []int{114, 255, 208}},
			{"index": 15, "name": "White", "rgb": []int{255, 255, 255}},
		}
		
	case "DHR2":
		// 560x192 Double High Resolution Graphics Page 2 (or 140x192 in color mode)
		// For simplicity, we'll read it as 140x192 color mode
		fullWidth = 140
		fullHeight = 192
		page = "2"
		
		// DHR uses the same 16-color palette as lo-res
		palette = []map[string]interface{}{
			{"index": 0, "name": "Black", "rgb": []int{0, 0, 0}},
			{"index": 1, "name": "Magenta", "rgb": []int{227, 30, 96}},
			{"index": 2, "name": "Dark Blue", "rgb": []int{96, 78, 189}},
			{"index": 3, "name": "Purple", "rgb": []int{255, 68, 253}},
			{"index": 4, "name": "Dark Green", "rgb": []int{0, 163, 96}},
			{"index": 5, "name": "Grey 1", "rgb": []int{156, 156, 156}},
			{"index": 6, "name": "Medium Blue", "rgb": []int{20, 207, 253}},
			{"index": 7, "name": "Light Blue", "rgb": []int{208, 195, 255}},
			{"index": 8, "name": "Brown", "rgb": []int{96, 114, 3}},
			{"index": 9, "name": "Orange", "rgb": []int{255, 106, 60}},
			{"index": 10, "name": "Grey 2", "rgb": []int{156, 156, 156}},
			{"index": 11, "name": "Pink", "rgb": []int{255, 160, 208}},
			{"index": 12, "name": "Light Green", "rgb": []int{20, 245, 60}},
			{"index": 13, "name": "Yellow", "rgb": []int{208, 221, 141}},
			{"index": 14, "name": "Aqua", "rgb": []int{114, 255, 208}},
			{"index": 15, "name": "White", "rgb": []int{255, 255, 255}},
		}
		
	default:
		// Other modes not yet implemented
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Graphics mode '%s' not yet supported", videoMode),
			}},
		}, nil
	}
	
	// Determine the actual rectangle to read
	startX := args.X
	startY := args.Y
	width := args.W
	height := args.H
	
	// Default to full screen if parameters not provided
	if width == 0 {
		width = fullWidth - startX
	}
	if height == 0 {
		height = fullHeight - startY
	}
	
	// Validate and clamp the rectangle to screen bounds
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if startX >= fullWidth {
		return nil, fmt.Errorf("x coordinate %d is outside screen bounds (0-%d)", startX, fullWidth-1)
	}
	if startY >= fullHeight {
		return nil, fmt.Errorf("y coordinate %d is outside screen bounds (0-%d)", startY, fullHeight-1)
	}
	if startX + width > fullWidth {
		width = fullWidth - startX
	}
	if startY + height > fullHeight {
		height = fullHeight - startY
	}
	
	// Create pixel array for the requested rectangle
	pixels := make([]uint64, width*height)
	
	// Read pixels based on the mode
	switch videoMode {
	case "LOGR":
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				pixels[y*width+x] = apple2helpers.GR40At(e, uint64(startX+x), uint64(startY+y))
			}
		}
	case "DLGR":
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				pixels[y*width+x] = apple2helpers.GR80At(e, uint64(startX+x), uint64(startY+y))
			}
		}
	case "HGR1", "HGR2":
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				pixels[y*width+x] = apple2helpers.HGRAt(e, uint64(startX+x), uint64(startY+y))
			}
		}
	case "DHR1", "DHR2":
		// For DHR color mode, we need to scale the x coordinate
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				pixels[y*width+x] = apple2helpers.HGRAt(e, uint64((startX+x)*4), uint64(startY+y))
			}
		}
	}
	
	// Build the response
	result := map[string]interface{}{
		"mode":       videoMode,
		"page":       page,
		"x":          startX,
		"y":          startY,
		"width":      width,
		"height":     height,
		"fullWidth":  fullWidth,
		"fullHeight": fullHeight,
		"pixels":     pixels,
		"palette":    palette,
	}
	
	// Convert to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: string(jsonData),
		}},
	}, nil
}

// Debugger handler functions
func handleDebugCPUControl(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugCPUControlParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	args := params.Arguments
	var result string

	switch args.Action {
	case "pause":
		debugger.DebuggerInstance.PauseCPU()
		result = "CPU paused"
	case "continue":
		debugger.DebuggerInstance.ContinueCPU()
		result = "CPU resumed"
	case "step":
		debugger.DebuggerInstance.StepCPU()
		result = "CPU stepped one instruction"
	case "step-over":
		debugger.DebuggerInstance.StepCPUOver()
		result = "CPU stepped over"
	case "step-out":
		debugger.DebuggerInstance.StepCPUOut()
		result = "CPU stepped out"
	default:
		return nil, fmt.Errorf("unknown action: %s", args.Action)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

func handleDebugBreakpointAdd(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugBreakpointAddParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	args := params.Arguments
	bp := &debugtypes.CPUBreakpoint{
		Disabled: false,
	}

	// Set breakpoint based on type
	var responseMsg string
	if args.Address != nil {
		switch args.Type {
		case "memory-write":
			bp.WriteAddress = args.Address
			responseMsg = fmt.Sprintf("Write breakpoint added at $%04X", *args.Address)
		case "memory-read":
			bp.ReadAddress = args.Address
			responseMsg = fmt.Sprintf("Read breakpoint added at $%04X", *args.Address)
		case "address", "":
			bp.ValuePC = args.Address
			responseMsg = fmt.Sprintf("Breakpoint added at $%04X", *args.Address)
		default:
			return nil, fmt.Errorf("unsupported breakpoint type: %s", args.Type)
		}
	} else {
		return nil, fmt.Errorf("address is required for breakpoint")
	}

	// Parse condition if provided
	if args.Condition != "" {
		if !bp.ParseArg(args.Condition) {
			return nil, fmt.Errorf("invalid breakpoint condition: %s", args.Condition)
		}
	}

	debugger.DebuggerInstance.AddBreakpoint(bp)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: responseMsg,
		}},
	}, nil
}

func handleDebugBreakpointRemove(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugBreakpointRemoveParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	debugger.DebuggerInstance.RemoveBreakpoint(params.Arguments.Index)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Breakpoint %d removed", params.Arguments.Index),
		}},
	}, nil
}

func handleDebugBreakpointList(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugBreakpointListParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	bps := debugger.DebuggerInstance.GetBreakpoints()
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Total breakpoints: %d\n", len(bps)))
	for i, bp := range bps {
		status := "enabled"
		if bp.Disabled {
			status = "disabled"
		}
		result.WriteString(fmt.Sprintf("[%d] %s - %s\n", i, bp.String(), status))
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil
}

func handleDebugMemoryRead(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugMemoryReadParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	args := params.Arguments
	data := debugger.DebuggerInstance.ReadBlob(args.Address, args.Count)

	// Format as hex dump
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Memory at $%04X:\n", args.Address))

	for i := 0; i < len(data); i += 16 {
		result.WriteString(fmt.Sprintf("%04X: ", args.Address + i))

		// Hex bytes
		for j := 0; j < 16 && i+j < len(data); j++ {
			result.WriteString(fmt.Sprintf("%02X ", data[i+j]))
		}

		// ASCII representation
		result.WriteString(" |")
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b < 127 {
				result.WriteByte(b)
			} else {
				result.WriteByte('.')
			}
		}
		result.WriteString("|\n")
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil
}

func handleDebugMemoryWrite(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugMemoryWriteParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	args := params.Arguments

	// Convert int slice to byte slice
	data := make([]byte, len(args.Data))
	for i, v := range args.Data {
		if v < 0 || v > 255 {
			return nil, fmt.Errorf("invalid byte value at index %d: %d", i, v)
		}
		data[i] = byte(v)
	}

	debugger.DebuggerInstance.WriteBlob(args.Address, data)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Wrote %d bytes to memory at $%04X", len(data), args.Address),
		}},
	}, nil
}

func handleDebugRegisterGet(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugRegisterGetParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	// Get current CPU state
	cpu := apple2helpers.GetCPU(backend.ProducerMain.GetInterpreter(0))

	result := fmt.Sprintf(
		"CPU Registers:\nA: $%02X (%d)\nX: $%02X (%d)\nY: $%02X (%d)\nSP: $%02X\nPC: $%04X\nP: $%02X (N:%d V:%d -:%d B:%d D:%d I:%d Z:%d C:%d)",
		cpu.A, cpu.A,
		cpu.X, cpu.X,
		cpu.Y, cpu.Y,
		cpu.SP,
		cpu.PC,
		cpu.P,
		(cpu.P>>7)&1, (cpu.P>>6)&1, (cpu.P>>5)&1, (cpu.P>>4)&1,
		(cpu.P>>3)&1, (cpu.P>>2)&1, (cpu.P>>1)&1, cpu.P&1,
	)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

func handleDebugRegisterSet(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugRegisterSetParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	args := params.Arguments

	// Map register name to debugger SetVal format
	regName := fmt.Sprintf("6502.%s", args.Register)
	debugger.DebuggerInstance.SetVal(regName, args.Value)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Register %s set to $%02X (%d)", strings.ToUpper(args.Register), args.Value, args.Value),
		}},
	}, nil
}

func handleDebugInstructionTrace(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugInstructionTraceParams]) (*mcp.CallToolResultFor[any], error) {
	if err := ensureDebuggerStarted(); err != nil {
		return nil, err
	}

	args := params.Arguments

	// Default values
	backlog := args.Backlog
	if backlog == 0 {
		backlog = 10
	}
	lookahead := args.Lookahead
	if lookahead == 0 {
		lookahead = 10
	}

	// Get the instruction trace
	trace := debugger.DebuggerInstance.GetInstructionTrace(backlog, lookahead)
	if trace == nil {
		return nil, fmt.Errorf("failed to get instruction trace")
	}

	// Format the output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("CPU Instruction Trace (-%d/+%d):\n", backlog, lookahead))
	result.WriteString("Address  Bytes       Instruction\n")
	result.WriteString("-------  ----------  -----------\n")

	for i, instr := range trace.Instructions {
		// Mark current instruction
		marker := "  "
		if i == backlog {
			marker = "> "
		}

		// Format bytes
		var byteStr strings.Builder
		for _, b := range instr.Bytes {
			byteStr.WriteString(fmt.Sprintf("%02X ", b))
		}

		// Add historic marker
		historic := ""
		if instr.Historic {
			historic = " [H]"
		}

		result.WriteString(fmt.Sprintf("%s$%04X   %-11s %s%s\n",
			marker, instr.Address, byteStr.String(), instr.Instruction, historic))
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil
}

func handleDisassemble(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DisassembleParams]) (*mcp.CallToolResultFor[any], error) {
	args := params.Arguments

	// Default to 20 instructions if not specified
	count := args.Count
	if count == 0 {
		count = 20
	}

	// Get the CPU interface
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	cpu := apple2helpers.GetCPU(e)

	// Format the output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Disassembly starting at $%04X:\n", args.Address))
	result.WriteString("Address  Bytes       Instruction\n")
	result.WriteString("-------  ----------  -----------\n")

	addr := args.Address
	for i := 0; i < count && addr <= 0xFFFF; i++ {
		// Decode the instruction at current address
		bytes, desc, cycles := cpu.DecodeInstruction(addr)

		// Format bytes
		var byteStr strings.Builder
		for _, b := range bytes {
			byteStr.WriteString(fmt.Sprintf("%02X ", b))
		}

		// Write the disassembled line
		result.WriteString(fmt.Sprintf("$%04X   %-11s %s",
			addr, byteStr.String(), desc))

		// Add cycle count if available
		if cycles > 0 {
			result.WriteString(fmt.Sprintf(" ; %d cycles", cycles))
		}
		result.WriteString("\n")

		// Move to next instruction
		addr += len(bytes)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil
}

func handleGetMountedDisks(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetMountedDisksParams]) (*mcp.CallToolResultFor[any], error) {
	// Create a structured response for mounted disks
	type DiskInfo struct {
		Drive        int    `json:"drive"`
		Path         string `json:"path"`
		WriteProtect bool   `json:"write_protect"`
		Type         string `json:"type"`
	}
	
	var disks []DiskInfo
	
	// Drive 0 (5.25" floppy)
	if settings.PureBootVolume[SelectedIndex] != "" {
		disks = append(disks, DiskInfo{
			Drive:        0,
			Path:         settings.PureBootVolume[SelectedIndex],
			WriteProtect: settings.PureBootVolumeWP[SelectedIndex],
			Type:         "5.25\" floppy",
		})
	}
	
	// Drive 1 (5.25" floppy)
	if settings.PureBootVolume2[SelectedIndex] != "" {
		disks = append(disks, DiskInfo{
			Drive:        1,
			Path:         settings.PureBootVolume2[SelectedIndex],
			WriteProtect: settings.PureBootVolumeWP2[SelectedIndex],
			Type:         "5.25\" floppy",
		})
	}
	
	// Drive 2 (SmartPort - 3.5" or hard disk)
	if settings.PureBootSmartVolume[SelectedIndex] != "" {
		disks = append(disks, DiskInfo{
			Drive:        2,
			Path:         settings.PureBootSmartVolume[SelectedIndex],
			WriteProtect: false, // SmartPort devices don't have a WP flag in settings
			Type:         "SmartPort (3.5\" or HD)",
		})
	}
	
	// Format the response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Mounted disks for slot %d:\n\n", SelectedIndex))
	
	if len(disks) == 0 {
		result.WriteString("No disks mounted\n")
	} else {
		for _, disk := range disks {
			result.WriteString(fmt.Sprintf("Drive %d (%s):\n", disk.Drive, disk.Type))
			result.WriteString(fmt.Sprintf("  Path: %s\n", disk.Path))
			result.WriteString(fmt.Sprintf("  Write Protected: %v\n", disk.WriteProtect))
			result.WriteString("\n")
		}
	}
	
	// Also return as JSON for programmatic access
	jsonData, err := json.MarshalIndent(disks, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal disk info: %w", err)
	}
	
	result.WriteString("\nJSON representation:\n")
	result.WriteString(string(jsonData))
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil
}

func handleEmulatorState(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[EmulatorStateParams]) (*mcp.CallToolResultFor[any], error) {
	// Get the interpreter for the selected slot
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Determine if CPU is running (check various states)
	isRunning := false
	cpuState := "Unknown"
	isZ80Mode := false
	if e != nil {
		state := e.GetState()
		switch state {
		case types.EXEC6502:
			isRunning = true
			cpuState = "Running (6502)"
		case types.DIRECTEXEC6502:
			isRunning = true
			cpuState = "Running Interactive (6502)"
		case types.EXECZ80:
			isRunning = true
			cpuState = "Running (Z80)"
			isZ80Mode = true
		case types.DIRECTEXECZ80:
			isRunning = true
			cpuState = "Running Interactive (Z80)"
			isZ80Mode = true
		case types.STOPPED:
			isRunning = false
			cpuState = "Stopped"
		case types.PAUSED:
			isRunning = false
			cpuState = "Paused"
		case types.BREAK:
			isRunning = false
			cpuState = "Break"
		case types.INPUT:
			isRunning = false
			cpuState = "Input"
		case types.RUNNING:
			isRunning = false
			cpuState = "BASIC Running"
		case types.DIRECTRUNNING:
			isRunning = false
			cpuState = "BASIC Direct Mode"
		default:
			isRunning = false
			cpuState = fmt.Sprintf("State: %v", state)
		}
	}

	// Get appropriate CPU state based on mode
	var cpuInfo map[string]interface{}
	if isZ80Mode {
		// Get Z80 CPU state
		z80core := apple2helpers.GetZ80CPU(e)
		if z80core != nil && z80core.Z80() != nil {
			z80cpu := z80core.Z80()
			cpuInfo = map[string]interface{}{
				"running":         isRunning,
				"state":           cpuState,
				"architecture":    "Z80",
				"program_counter": fmt.Sprintf("$%04X", z80cpu.PC()),
				"pc_decimal":      z80cpu.PC(),
				"registers": map[string]interface{}{
					"A":  fmt.Sprintf("$%02X", z80cpu.A),
					"F":  fmt.Sprintf("$%02X", z80cpu.F),
					"BC": fmt.Sprintf("$%04X", z80cpu.BC()),
					"DE": fmt.Sprintf("$%04X", z80cpu.DE()),
					"HL": fmt.Sprintf("$%04X", z80cpu.HL()),
					"IX": fmt.Sprintf("$%04X", z80cpu.IX()),
					"IY": fmt.Sprintf("$%04X", z80cpu.IY()),
					"SP": fmt.Sprintf("$%04X", z80cpu.SP()),
					"PC": fmt.Sprintf("$%04X", z80cpu.PC()),
					// Alternate registers
					"A'":  fmt.Sprintf("$%02X", z80cpu.A_),
					"F'":  fmt.Sprintf("$%02X", z80cpu.F_),
					"BC'": fmt.Sprintf("$%04X", uint16(z80cpu.B_)<<8|uint16(z80cpu.C_)),
					"DE'": fmt.Sprintf("$%04X", uint16(z80cpu.D_)<<8|uint16(z80cpu.E_)),
					"HL'": fmt.Sprintf("$%04X", uint16(z80cpu.H_)<<8|uint16(z80cpu.L_)),
				},
				"flags": map[string]interface{}{
					"S": (z80cpu.F >> 7) & 1,  // Sign
					"Z": (z80cpu.F >> 6) & 1,  // Zero
					"H": (z80cpu.F >> 4) & 1,  // Half Carry
					"P": (z80cpu.F >> 2) & 1,  // Parity/Overflow
					"N": (z80cpu.F >> 1) & 1,  // Add/Subtract
					"C": z80cpu.F & 1,         // Carry
				},
			}
		} else {
			// Fallback if Z80 CPU not available
			cpuInfo = map[string]interface{}{
				"running":      isRunning,
				"state":        cpuState,
				"architecture": "Z80",
				"error":        "Z80 CPU state not available",
			}
		}
	} else {
		// Get 6502 CPU state
		cpu := apple2helpers.GetCPU(e)
		cpuInfo = map[string]interface{}{
			"running":         isRunning,
			"state":           cpuState,
			"architecture":    "6502",
			"program_counter": fmt.Sprintf("$%04X", cpu.PC),
			"pc_decimal":      cpu.PC,
			"registers": map[string]interface{}{
				"A":  fmt.Sprintf("$%02X", cpu.A),
				"X":  fmt.Sprintf("$%02X", cpu.X),
				"Y":  fmt.Sprintf("$%02X", cpu.Y),
				"SP": fmt.Sprintf("$%02X", cpu.SP),
				"P":  fmt.Sprintf("$%02X", cpu.P),
			},
			"flags": map[string]interface{}{
				"N": (cpu.P >> 7) & 1,  // Negative
				"V": (cpu.P >> 6) & 1,  // Overflow
				"B": (cpu.P >> 4) & 1,  // Break
				"D": (cpu.P >> 3) & 1,  // Decimal
				"I": (cpu.P >> 2) & 1,  // Interrupt disable
				"Z": (cpu.P >> 1) & 1,  // Zero
				"C": cpu.P & 1,         // Carry
			},
		}
	}
	
	// Get machine profile
	machineProfile := settings.SpecFile[SelectedIndex]
	// Remove file extension if present
	if idx := strings.LastIndex(machineProfile, "."); idx != -1 {
		machineProfile = machineProfile[:idx]
	}
	
	// Get video mode
	videoMode := apple2helpers.GetVideoMode(e)

	// Get active text modes to determine if text screen should be read
	activeTextModes := apple2helpers.GetActiveTextModes(e)

	// Prepare text screen content if TEXT or TXT2 is active
	var textScreen string
	var hasTextMode bool
	for _, mode := range activeTextModes {
		if mode == "TEXT" || mode == "TXT2" {
			hasTextMode = true
			break
		}
	}

	if hasTextMode {
		// Try to get text screen content from TEXT or TXT2 layers only
		for _, l := range HUDLayers[SelectedIndex] {
			if l == nil || !l.Spec.GetActive() {
				continue
			}

			layerID := l.Spec.GetID()

			// Only read from TEXT or TXT2 layers, ignore MONI and OOSD
			if ll, ok := e.GetHUDLayerByID(layerID); ok {
				if layerID == "TEXT" || layerID == "TXT2" {
					text := strings.Join(ll.Control.GetStrings(), "")
					if text != "" {
						textScreen = strings.ReplaceAll(text, "\n\n", "\n")
						break
					}
				}
			}
		}

		if textScreen == "" {
			textScreen = "(Text screen not available)"
		}
	}
	
	// Get disk information
	var disks []map[string]interface{}
	
	// Drive 0 (5.25" floppy)
	if settings.PureBootVolume[SelectedIndex] != "" {
		disks = append(disks, map[string]interface{}{
			"drive":         0,
			"path":          settings.PureBootVolume[SelectedIndex],
			"write_protect": settings.PureBootVolumeWP[SelectedIndex],
			"type":          "5.25\" floppy",
		})
	}
	
	// Drive 1 (5.25" floppy)
	if settings.PureBootVolume2[SelectedIndex] != "" {
		disks = append(disks, map[string]interface{}{
			"drive":         1,
			"path":          settings.PureBootVolume2[SelectedIndex],
			"write_protect": settings.PureBootVolumeWP2[SelectedIndex],
			"type":          "5.25\" floppy",
		})
	}
	
	// Drive 2 (SmartPort - 3.5" or hard disk)
	if settings.PureBootSmartVolume[SelectedIndex] != "" {
		disks = append(disks, map[string]interface{}{
			"drive":         2,
			"path":          settings.PureBootSmartVolume[SelectedIndex],
			"write_protect": false,
			"type":          "SmartPort",
		})
	}
	
	// Create structured state information
	stateInfo := map[string]interface{}{
		"machine": machineProfile,
		"cpu": cpuInfo,
		"video": map[string]interface{}{
			"mode": videoMode,
			"description": getVideoModeDescription(videoMode),
		},
		"disks": disks,
		"slot": SelectedIndex,
	}
	
	// Add text screen content if available
	if textScreen != "" && hasTextMode {
		stateInfo["text_screen"] = textScreen
	}
	
	// Format the response
	var result strings.Builder
	result.WriteString("=== Emulator State ===\n\n")
	
	// Machine Information
	result.WriteString(fmt.Sprintf("Machine: %s\n", machineProfile))
	
	// CPU Information
	result.WriteString(fmt.Sprintf("CPU Status: %s\n", cpuState))

	if isZ80Mode {
		// Display Z80 registers
		z80core := apple2helpers.GetZ80CPU(e)
		if z80core != nil && z80core.Z80() != nil {
			z80cpu := z80core.Z80()
			result.WriteString(fmt.Sprintf("Architecture: Z80\n"))
			result.WriteString(fmt.Sprintf("Program Counter: $%04X (%d)\n", z80cpu.PC(), z80cpu.PC()))
			result.WriteString(fmt.Sprintf("Main Registers: A=$%02X F=$%02X BC=$%04X DE=$%04X HL=$%04X\n",
				z80cpu.A, z80cpu.F, z80cpu.BC(), z80cpu.DE(), z80cpu.HL()))
			result.WriteString(fmt.Sprintf("Index Registers: IX=$%04X IY=$%04X SP=$%04X\n",
				z80cpu.IX(), z80cpu.IY(), z80cpu.SP()))
			result.WriteString(fmt.Sprintf("Alternate Set: A'=$%02X F'=$%02X BC'=$%04X DE'=$%04X HL'=$%04X\n",
				z80cpu.A_, z80cpu.F_, uint16(z80cpu.B_)<<8|uint16(z80cpu.C_),
				uint16(z80cpu.D_)<<8|uint16(z80cpu.E_), uint16(z80cpu.H_)<<8|uint16(z80cpu.L_)))
			result.WriteString(fmt.Sprintf("Flags: S=%d Z=%d H=%d P=%d N=%d C=%d\n\n",
				(z80cpu.F>>7)&1, (z80cpu.F>>6)&1, (z80cpu.F>>4)&1, (z80cpu.F>>2)&1,
				(z80cpu.F>>1)&1, z80cpu.F&1))
		} else {
			result.WriteString("Z80 CPU state not available\n\n")
		}
	} else {
		// Display 6502 registers
		cpu := apple2helpers.GetCPU(e)
		result.WriteString(fmt.Sprintf("Architecture: 6502\n"))
		result.WriteString(fmt.Sprintf("Program Counter: $%04X (%d)\n", cpu.PC, cpu.PC))
		result.WriteString(fmt.Sprintf("Registers: A=$%02X X=$%02X Y=$%02X SP=$%02X P=$%02X\n",
			cpu.A, cpu.X, cpu.Y, cpu.SP, cpu.P))
		result.WriteString(fmt.Sprintf("Flags: N=%d V=%d B=%d D=%d I=%d Z=%d C=%d\n\n",
			(cpu.P>>7)&1, (cpu.P>>6)&1, (cpu.P>>4)&1, (cpu.P>>3)&1,
			(cpu.P>>2)&1, (cpu.P>>1)&1, cpu.P&1))
	}
	
	// Video Information
	result.WriteString(fmt.Sprintf("Video Mode: %s (%s)\n\n", videoMode, getVideoModeDescription(videoMode)))
	
	// Disk Information
	if len(disks) > 0 {
		result.WriteString("Mounted Disks:\n")
		for _, disk := range disks {
			result.WriteString(fmt.Sprintf("  Drive %d: %s (%s)\n", 
				disk["drive"], disk["path"], disk["type"]))
		}
		result.WriteString("\n")
	} else {
		result.WriteString("No disks mounted\n\n")
	}
	
	// Text Screen Content (if in TEXT mode)
	if textScreen != "" && hasTextMode {
		result.WriteString("Text Screen Content:\n")
		result.WriteString("-------------------\n")
		result.WriteString(textScreen)
		result.WriteString("\n-------------------\n\n")
	}
	
	// JSON representation for programmatic access
	jsonData, err := json.MarshalIndent(stateInfo, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state info: %w", err)
	}
	
	result.WriteString("JSON representation:\n")
	result.WriteString(string(jsonData))
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil
}

// Helper function to get video mode description
func getVideoModeDescription(mode string) string {
	switch mode {
	case "TEXT":
		return "40x24 Text Mode"
	case "TXT2":
		return "80x24 Text Mode"
	case "LOGR":
		return "40x48 Low Resolution Graphics"
	case "DLGR":
		return "80x48 Double Low Resolution Graphics"
	case "HGR1":
		return "280x192 High Resolution Graphics Page 1"
	case "HGR2":
		return "280x192 High Resolution Graphics Page 2"
	case "DHGR":
		return "560x192 Double High Resolution Graphics"
	default:
		return "Unknown Mode"
	}
}

func handleEnableLiveRewind(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[EnableLiveRewindParams]) (*mcp.CallToolResultFor[any], error) {
	// Get the interpreter for the selected slot
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Check if already recording
	if e.IsRecordingVideo() {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: "Live rewind is already enabled (recording in progress)",
			}},
		}, nil
	}
	
	// Start recording with empty filename (for live rewind) and false for not saving to file
	e.StartRecording("", false)
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Live rewind enabled for slot %d", SelectedIndex),
		}},
	}, nil
}

func handleRewindBack(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[RewindBackParams]) (*mcp.CallToolResultFor[any], error) {
	// Get the interpreter for the selected slot
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Get milliseconds parameter with default of 5000
	milliseconds := params.Arguments.Milliseconds
	if milliseconds == 0 {
		milliseconds = 5000
	}
	
	// Check if recording is active (required for rewind)
	if !e.IsRecordingVideo() {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: "Cannot rewind: Live rewind is not enabled. Use enable_live_rewind first.",
			}},
		}, nil
	}
	
	// Trigger the backstep in a goroutine as required
	go e.BackstepVideo(milliseconds)
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Rewinding %d milliseconds for slot %d", milliseconds, SelectedIndex),
		}},
	}, nil
}

func handleStartFileRecording(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[StartFileRecordingParams]) (*mcp.CallToolResultFor[any], error) {
	// Get the interpreter for the selected slot
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Stop any existing recording first
	e.StopRecording()
	
	// Determine full CPU recording setting
	fullCPU := settings.FileFullCPURecord
	if params.Arguments.FullCPU {
		fullCPU = true
	}
	
	// Start file recording
	e.RecordToggle(fullCPU)
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("File recording started for slot %d (full CPU: %v)", SelectedIndex, fullCPU),
		}},
	}, nil
}

func handleStopRecording(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[StopRecordingParams]) (*mcp.CallToolResultFor[any], error) {
	// Get the interpreter for the selected slot
	e := backend.ProducerMain.GetInterpreter(SelectedIndex)
	
	// Stop recording
	e.StopRecording()
	
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Recording stopped for slot %d", SelectedIndex),
		}},
	}, nil
}

// createMCPServer creates and configures the MCP server with all tools
func createMCPServer() *mcp.Server {
	// Create a new server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "microM8",
		Version: "1.0.0",
	}, nil)
	
	// Register all tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "reboot",
		Description: "Reboot the emulator",
	}, handleReboot)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pause",
		Description: "Pause or unpause the emulator",
	}, handlePause)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "break",
		Description: "Send a break signal (Ctrl+C) to the emulator",
	}, handleBreak)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "insert_disk",
		Description: "Insert a disk into a drive",
	}, handleInsertDisk)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "eject_disk",
		Description: "Eject a disk from a drive",
	}, handleEjectDisk)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_mounted_disks",
		Description: "Get information about currently mounted disks",
	}, handleGetMountedDisks)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "emulator_state",
		Description: "Get comprehensive emulator state including CPU status, video mode, text screen, and disk information",
	}, handleEmulatorState)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "enable_live_rewind",
		Description: "Enable live rewind functionality by starting video recording",
	}, handleEnableLiveRewind)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rewind_back",
		Description: "Rewind the emulator by specified milliseconds (default: 5000ms)",
	}, handleRewindBack)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "start_file_recording",
		Description: "Start recording emulator output to a file",
	}, handleStartFileRecording)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "stop_recording",
		Description: "Stop any active recording (file or live rewind)",
	}, handleStopRecording)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "key_event",
		Description: "Send a keyboard event to the emulator",
	}, handleKeyEvent)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "screenshot",
		Description: "Take a screenshot of the emulator",
	}, handleScreenshot)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_text_screen",
		Description: "Get the current text screen contents",
	}, handleGetTextScreen)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_memory",
		Description: "Read a byte from memory",
	}, handleReadMemory)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_memory",
		Description: "Write a byte to memory",
	}, handleWriteMemory)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_memory_range",
		Description: "Write multiple bytes to memory starting at a given address",
	}, handleWriteMemoryRange)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_cpu_speed",
		Description: "Set CPU speed multiplier",
	}, handleSetCPUSpeed)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "type_text",
		Description: "Type text into the emulator with configurable delay between keystrokes",
	}, handleTypeText)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "assemble",
		Description: "Assemble 6502 assembly code and write it to emulator memory",
	}, handleAssemble)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "applesoft_read",
		Description: "Read and detokenize the Applesoft BASIC program currently in memory",
	}, handleApplesoftRead)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "applesoft_write",
		Description: "Write an Applesoft BASIC program to memory (tokenizes the source code)",
	}, handleApplesoftWrite)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_files",
		Description: "List files and directories from the microM8 virtual file system",
	}, handleListFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_file",
		Description: "Read a file from the microM8 virtual file system",
	}, handleReadFile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "insert_disk_file",
		Description: "Insert a disk image file from the virtual file system into a drive",
	}, handleInsertDiskFile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "load_basic_program",
		Description: "Load and optionally run a BASIC program (Applesoft, Integer BASIC, or Logo)",
	}, handleLoadBasicProgram)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_interpreter_code",
		Description: "Read the current code from an interpreter (Applesoft/microBASIC, Integer BASIC, or Logo)",
	}, handleReadInterpreterCode)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_interpreter_code",
		Description: "Write code to an interpreter (Applesoft/microBASIC, Integer BASIC, or Logo)",
	}, handleWriteInterpreterCode)

	// mcp.AddTool(server, &mcp.Tool{
	// 	Name:        "list_appleii_tree",
	// 	Description: "Recursively list all folders under /appleii/ directory",
	// }, handleListAppleIITree)

	// Gaming tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_graphics_screen",
		Description: "Capture the current graphics screen as image/png",
	}, handleGetGraphicsScreen)
 //
	mcp.AddTool(server, &mcp.Tool{
		Name:        "joystick_event",
		Description: "Send joystick or paddle controller events to the emulator",
	}, handleJoystickEvent)
 //
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_game_state",
		Description: "Get structured game state for known games (score, lives, level, etc.)",
	}, handleGetGameState)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_current_graphics_data",
		Description: "Get the current graphics mode data including pixel color indices and palette",
	}, handleGetCurrentGraphicsData)

	// Debugger tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_cpu_control",
		Description: "Control CPU execution (pause, continue, step, step-over, step-out)",
	}, handleDebugCPUControl)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_breakpoint_add",
		Description: "Add a breakpoint at a specific address",
	}, handleDebugBreakpointAdd)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_breakpoint_remove",
		Description: "Remove a breakpoint by index",
	}, handleDebugBreakpointRemove)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_breakpoint_list",
		Description: "List all breakpoints",
	}, handleDebugBreakpointList)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_memory_read",
		Description: "Read bytes from memory",
	}, handleDebugMemoryRead)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_memory_write",
		Description: "Write bytes to memory",
	}, handleDebugMemoryWrite)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_register_get",
		Description: "Get CPU register values",
	}, handleDebugRegisterGet)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_register_set",
		Description: "Set CPU register value",
	}, handleDebugRegisterSet)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_instruction_trace",
		Description: "Get CPU instruction trace showing previous and upcoming instructions",
	}, handleDebugInstructionTrace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "disassemble",
		Description: "Disassemble 6502 machine code from memory using CPU instruction decode logic",
	}, handleDisassemble)

	return server
}

// StartMCPServerSDK starts the MCP server using the official Go SDK
func StartMCPServerSDK() error {
	// Create the server
	server := createMCPServer()
	
	// Check transport mode
	switch *mcpTransport {
	case "sse":
		// Start SSE server
		return startMCPServerSSE(server, *mcpPort)
	case "streaming", "http-streaming":
		// Start HTTP streaming server
		return startMCPServerHTTPStreaming(server, *mcpPort)
	default:
		// Default to stdio transport
		log.Println("Starting MCP server on stdio...")
		if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
			return fmt.Errorf("MCP server error: %w", err)
		}
	}
	
	return nil
}

// startMCPServerSSE starts the MCP server with SSE transport
func startMCPServerSSE(server *mcp.Server, port int) error {
	// Create SSE handler using the official SDK pattern
	sseHandler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		// Return the server for all requests
		log.Printf("Handling MCP request for URL %s", r.URL.Path)
		return server
	})

	// Set up HTTP handler with CORS and heartbeat support
	http.HandleFunc("/mcp/sse", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received: %s %s", r.Method, r.URL.Path)
		log.Printf("Headers: %v", r.Header)
		log.Printf("Remote addr: %s", r.RemoteAddr)
		
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, mcp-session-id")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			log.Printf("Handling OPTIONS preflight request")
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// For SSE connections, set proper headers and handle keepalive
		if r.Method == "GET" && r.Header.Get("Accept") == "text/event-stream" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering
			
			// Flush headers immediately
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			
			log.Printf("SSE connection established, starting keepalive")
			atomic.AddInt64(&activeConnections, 1)
			
			defer func() {
				atomic.AddInt64(&activeConnections, -1)
				log.Printf("SSE connection closed (remaining: %d)", atomic.LoadInt64(&activeConnections))
			}()
			
			// Start a goroutine to send keepalive messages
			ctx := r.Context()
			go func() {
				ticker := time.NewTicker(heartbeatInterval)
				defer ticker.Stop()
				
				for {
					select {
					case <-ctx.Done():
						log.Printf("SSE connection closed, stopping keepalive")
						return
					case <-ticker.C:
						// Send a comment as keepalive (SSE comment starts with :)
						fmt.Fprintf(w, ":keepalive %s\n\n", time.Now().Format(time.RFC3339))
						if flusher, ok := w.(http.Flusher); ok {
							flusher.Flush()
						}
						log.Printf("Sent SSE keepalive")
					}
				}
			}()
		}
		
		log.Printf("Passing request to SSE handler")
		// Serve the SSE handler
		sseHandler.ServeHTTP(w, r)
	})

	// Add health check endpoint
	http.HandleFunc("/mcp/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","connections":%d,"uptime":"%s"}`,
			atomic.LoadInt64(&activeConnections),
			time.Since(serverStartTime).String())
	})

	// Start HTTP server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting MCP server on %s with SSE transport (heartbeat enabled)", addr)

	// Start HTTP server
	return http.ListenAndServe(addr, nil)
}

// startMCPServerHTTPStreaming starts the MCP server with HTTP streaming transport
func startMCPServerHTTPStreaming(server *mcp.Server, port int) error {
	// Create StreamableHTTPHandler using the official SDK
	streamHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			// Return the server for all requests
			log.Printf("Handling MCP streaming request for URL %s", r.URL.Path)
			return server
		},
		nil, // Use default options for now
	)

	// Create a new ServeMux to handle multiple endpoints
	mux := http.NewServeMux()
	
	// Main streaming endpoint
	mux.HandleFunc("/mcp/stream", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Streaming request received: %s %s", r.Method, r.URL.Path)
		log.Printf("Headers: %v", r.Header)
		log.Printf("Remote addr: %s", r.RemoteAddr)
		
		// Set CORS headers for all requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Mcp-Session-Id, Mcp-Protocol-Version, Last-Event-ID")
		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id, Mcp-Protocol-Version")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			log.Printf("Handling OPTIONS preflight request for streaming")
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Track connections for GET requests
		if r.Method == "GET" {
			atomic.AddInt64(&activeConnections, 1)
			defer func() {
				atomic.AddInt64(&activeConnections, -1)
				log.Printf("Streaming connection closed (remaining: %d)", atomic.LoadInt64(&activeConnections))
			}()
		}
		
		// Log session ID if present
		if sessionID := r.Header.Get("Mcp-Session-Id"); sessionID != "" {
			log.Printf("Session ID: %s", sessionID)
		}
		
		// Pass to the streaming handler
		streamHandler.ServeHTTP(w, r)
	})
	
	// Health check endpoint (same as SSE)
	mux.HandleFunc("/mcp/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, `{"status":"healthy","transport":"streaming","connections":%d,"uptime":"%s"}`,
			atomic.LoadInt64(&activeConnections),
			time.Since(serverStartTime).String())
	})
	
	// Info endpoint for streaming transport
	mux.HandleFunc("/mcp/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, `{"transport":"http-streaming","endpoint":"/mcp/stream","methods":["GET","POST","DELETE"],"version":"1.0"}`)
	})

	// Start HTTP server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting MCP server on %s with HTTP streaming transport", addr)
	
	return http.ListenAndServe(addr, mux)
}
