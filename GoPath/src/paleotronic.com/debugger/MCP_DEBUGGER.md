# MicroM8 Debugger MCP Integration Recommendations

## Overview
This document outlines the recommended debugger functions from `debugger.go` that would be useful to expose via MCP (Model Context Protocol) for external control of the MicroM8 Apple II emulator debugger.

## Important: Debugger Initialization
The debugger package maintains a global `DebuggerInstance` variable that must be initialized before use. All MCP debugger handlers MUST check if this instance exists and initialize it if needed:

```go
// At the start of EVERY MCP debugger handler:
if debugger.DebuggerInstance == nil {
    debugger.Start()  // This creates DebuggerInstance
    if debugger.DebuggerInstance != nil {
        debugger.DebuggerInstance.AttachSlot(0)  // Attach to default slot 0
    }
}
```

## Core Debugger Control Functions

### 1. CPU Execution Control
These functions control the CPU execution state and would be essential for MCP control:

- **`PauseCPU()`** - Pauses CPU execution
- **`ContinueCPU()`** - Resumes normal CPU execution
- **`StepCPU()`** - Single-step one instruction
- **`StepCPUOver()`** - Step over subroutine calls
- **`StepCPUOut()`** - Step out of current subroutine
- **`ContinueCPUOut()`** - Continue until out of current subroutine

### 2. Breakpoint Management
Complete breakpoint control for debugging:

- **`AddBreakpoint(b *CPUBreakpoint)`** - Add a new breakpoint
- **`RemoveBreakpoint(idx int)`** - Remove breakpoint by index
- **`UpdateBreakpoint(idx int, b *CPUBreakpoint)`** - Update existing breakpoint
- **`DisableBreakpoint(idx int)`** - Temporarily disable a breakpoint
- **`EnableBreakpoint(idx int)`** - Re-enable a disabled breakpoint
- **`GetBreakpoints()`** - Get list of all breakpoints
- **`ClearAllBreakpoints()`** - Remove all breakpoints
- **`DisableAllBreakpoints()`** - Disable all breakpoints
- **`EnableAllBreakpoints()`** - Enable all breakpoints
- **`SetBreakpointCounter(idx int, value int)`** - Set counter for conditional breakpoints
- **`ResetCounters()`** - Reset all breakpoint counters

### 3. Memory Operations
Direct memory access for inspection and modification:

- **`ReadBlob(addr int, count int)`** - Read bytes from memory
- **`WriteBlob(addr int, data []byte)`** - Write bytes to memory

### 4. CPU State Management
Functions to inspect and modify CPU registers:

- **`SetVal(name string, value int)`** - Set CPU register values
  - Supports: "6502.a", "6502.x", "6502.y", "6502.sp", "6502.pc", "6502.p"
- **`ToggleCPUFlag(name string)`** - Toggle CPU status flags
  - Supports: "N", "V", "I", "D", "C", "B", "Z"

### 5. Debugger Configuration
Runtime configuration of debugger behavior:

- **`SetVal()` for config options:**
  - "cpu-break-ill" - Break on illegal opcodes
  - "cpu-break-brk" - Break on BRK instruction
  - "cpu-full-record" - Enable full CPU recording
  - "cpu-update-ms" - CPU state update interval
  - "cpu-backlog-lines" - Instruction history size
  - "cpu-lookahead-lines" - Instruction lookahead size
  - "cpu-record-timing" - Timing points per second
  - "screen-refresh-ms" - Screen refresh interval

### 6. Tracing and Logging
CPU instruction tracing capabilities:

- **`Trace(verb string)`** - Start/stop CPU trace logging
  - "on" - Start tracing to file
  - "off" - Stop tracing

### 7. Soft Switch Control
Apple II specific hardware control:

- **`ToggleSoftSwitch(name string)`** - Toggle memory/video soft switches
- **`RequestSwitchStates()`** - Get current soft switch states

### 8. Session Management
Debugger session control:

- **`AttachSlot(slotid int)`** - Attach debugger to emulator slot
- **`Detach()`** - Detach debugger from current slot
- **`GetSlot()`** - Get current attached slot ID
- **`IsAttached()`** - Check if debugger is attached

### 9. Keyboard Input
Send keystrokes to the emulated system:

- **`SendKey(keycode int)`** - Send a key to the Apple II keyboard buffer

## Recommended MCP Tool Structure

### Tool Categories

1. **debug_cpu_control** - CPU execution control (pause, continue, step)
2. **debug_breakpoint** - Breakpoint management operations
3. **debug_memory** - Memory read/write operations
4. **debug_registers** - CPU register inspection/modification
5. **debug_config** - Debugger configuration
6. **debug_trace** - Tracing control
7. **debug_session** - Session management

### Example MCP Tool Definitions

```go
// Helper function to ensure debugger is initialized
func ensureDebuggerStarted() error {
    if debugger.DebuggerInstance == nil {
        debugger.Start()
        if debugger.DebuggerInstance == nil {
            return fmt.Errorf("failed to initialize debugger")
        }
        // Attach to default emulator slot
        debugger.DebuggerInstance.AttachSlot(0)
    }
    return nil
}

// CPU Control Tool Handler
func handleDebugCPUControl(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DebugCPUControlParams]) (*mcp.CallToolResultFor[any], error) {
    if err := ensureDebuggerStarted(); err != nil {
        return nil, err
    }
    
    // Now safe to use debugger.DebuggerInstance
    switch params.Arguments.Action {
    case "pause":
        debugger.DebuggerInstance.PauseCPU()
    case "continue":
        debugger.DebuggerInstance.ContinueCPU()
    case "step":
        debugger.DebuggerInstance.StepCPU()
    // ... etc
    }
}

// Register tools
mcp.AddTool(server, &mcp.Tool{
    Name:        "debug_cpu_control",
    Description: "Control CPU execution (pause, continue, step)",
}, handleDebugCPUControl)

mcp.AddTool(server, &mcp.Tool{
    Name:        "debug_breakpoint",
    Description: "Manage debugger breakpoints",
}, handleDebugBreakpoint)

mcp.AddTool(server, &mcp.Tool{
    Name:        "debug_memory",
    Description: "Read/write emulator memory",
}, handleDebugMemory)
```

## Implementation Notes

1. **Debugger Initialization**: CRITICAL - All MCP debugger calls must check if `DebuggerInstance` is initialized:
   ```go
   func ensureDebuggerStarted() {
       if debugger.DebuggerInstance == nil {
           debugger.Start()  // Initialize the debugger
           if debugger.DebuggerInstance != nil {
               debugger.DebuggerInstance.AttachSlot(0)  // Attach to default slot
           }
       }
   }
   ```
   This check should be at the beginning of every MCP debugger handler function.

2. **Thread Safety**: The debugger uses mutexes for thread safety. MCP handlers should respect this.

3. **Slot Management**: The debugger operates on a specific emulator slot. After initialization, it's attached to slot 0 by default. MCP tools should verify slot attachment before operations.

4. **State Synchronization**: The debugger maintains WebSocket connections for real-time updates. MCP operations should trigger appropriate state updates.

4. **Breakpoint Structures**: Breakpoints support complex conditions including:
   - Address matching (PC)
   - Register value matching (A, X, Y, SP, P)
   - Memory read/write matching
   - Actions (break, log, chime, speed change, etc.)

5. **Error Handling**: MCP tools should validate parameters and handle cases where:
   - Debugger is not attached
   - Invalid addresses or indices
   - CPU is in incompatible state

## Priority Implementation Order

1. **High Priority** (Core debugging):
   - CPU control (pause, continue, step)
   - Basic breakpoint operations (add, remove, list)
   - Memory read/write
   - Register inspection

2. **Medium Priority** (Enhanced debugging):
   - Advanced breakpoint features
   - Tracing control
   - Soft switch management
   - Configuration options

3. **Low Priority** (Nice to have):
   - Screenshot/screen state
   - Full state serialization
   - Complex breakpoint actions

## Security Considerations

- Memory operations should validate address ranges
- Configuration changes should be reversible
- Trace files should use secure paths
- Consider rate limiting for expensive operations
