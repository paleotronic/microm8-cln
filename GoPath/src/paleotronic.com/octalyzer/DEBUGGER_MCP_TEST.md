# MicroM8 Debugger MCP Implementation Test Guide

## High Priority Features Implemented

### 1. CPU Control Functions
- `debug_cpu_control` - Control CPU execution
  - Actions: `pause`, `continue`, `step`, `step-over`, `step-out`
  
Example usage:
```
mcp__microm8__debug_cpu_control action="pause"
mcp__microm8__debug_cpu_control action="step"
mcp__microm8__debug_cpu_control action="continue"
```

### 2. Breakpoint Operations
- `debug_breakpoint_add` - Add a breakpoint at an address
- `debug_breakpoint_remove` - Remove a breakpoint by index
- `debug_breakpoint_list` - List all breakpoints

Example usage:
```
mcp__microm8__debug_breakpoint_add address=0xFCA8
mcp__microm8__debug_breakpoint_list
mcp__microm8__debug_breakpoint_remove index=0
```

### 3. Memory Operations
- `debug_memory_read` - Read bytes from memory
- `debug_memory_write` - Write bytes to memory

Example usage:
```
mcp__microm8__debug_memory_read address=0x300 count=16
mcp__microm8__debug_memory_write address=0x300 data=[0xA9, 0x00, 0x60]
```

### 4. Register Operations  
- `debug_register_get` - Get all CPU register values
- `debug_register_set` - Set a specific register value

Example usage:
```
mcp__microm8__debug_register_get
mcp__microm8__debug_register_set register="a" value=0x42
mcp__microm8__debug_register_set register="pc" value=0xFCA8
```

## Implementation Details

1. **Automatic Initialization**: All debugger functions check if `debugger.DebuggerInstance` is nil and automatically initialize it if needed, attaching to slot 0.

2. **Error Handling**: Functions validate parameters and return appropriate error messages.

3. **Formatting**: 
   - Memory reads show hex dump with ASCII representation
   - Register displays show both hex and decimal values
   - Breakpoint lists show index, condition, and status

## Testing Procedure

After restarting the MCP server, test each function:

1. Start with a simple program loaded
2. Pause the CPU
3. Read registers
4. Set a breakpoint
5. Step through code
6. Read/write memory
7. Continue execution

## Notes

- The debugger automatically starts and attaches to slot 0 on first use
- Breakpoint implementation is simplified - full implementation would support all breakpoint types (register values, memory access, etc.)
- Register names are lowercase: a, x, y, sp, pc, p
- Memory addresses are 0-65535 (16-bit)
- Step operations work with the CPU's built-in stepping modes