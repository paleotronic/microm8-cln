# MCP Development Guide for Go

This guide explains how to build Model Context Protocol (MCP) servers in Go based on the microM8 emulator's implementation pattern.

## Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [Basic Setup](#basic-setup)
4. [Tool Parameter Definitions](#tool-parameter-definitions)
5. [Tool Handler Functions](#tool-handler-functions)
6. [Server Initialization](#server-initialization)
7. [Advanced Patterns](#advanced-patterns)
8. [Best Practices](#best-practices)
9. [Testing and Debugging](#testing-and-debugging)

## Overview

The Model Context Protocol (MCP) allows applications to expose functionality through a standardized interface. This guide demonstrates how to create an MCP server in Go using the official SDK pattern.

### Key Components

1. **Parameter Structs**: Define the input parameters for each tool
2. **Handler Functions**: Implement the logic for each tool
3. **Server Setup**: Initialize and register tools with the MCP server
4. **Transport Layer**: Handle communication (stdio, HTTP, etc.)

## Project Structure

```
your-project/
├── mcp_server.go        # Main MCP server implementation
├── handlers.go          # Tool handler functions (optional separation)
├── types.go            # Parameter struct definitions (optional separation)
└── main.go             # Application entry point
```

## Basic Setup

### 1. Install Dependencies

```bash
go get github.com/modelcontextprotocol/go-sdk/mcp
```

### 2. Import Required Packages

```go
import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
)
```

## Tool Parameter Definitions

### Basic Parameter Struct

```go
// Simple parameter struct with no inputs
type RebootParams struct{}

// Parameter struct with basic types
type ReadFileParams struct {
    Path string `json:"path" jsonschema:"description:Path to the file to read"`
}
```

### Advanced Parameter Struct

```go
type ComplexParams struct {
    // Required field with validation
    Address int `json:"address" jsonschema:"description:Memory address,minimum:0,maximum:65535"`
    
    // Optional field with default
    Count int `json:"count,omitempty" jsonschema:"description:Number of items (default: 10),minimum:1,maximum:100"`
    
    // Enum field
    Mode string `json:"mode" jsonschema:"description:Operation mode,enum:[read,write,execute]"`
    
    // Array field
    Values []int `json:"values" jsonschema:"description:Array of values (each 0-255)"`
    
    // Optional pointer field
    Condition *string `json:"condition,omitempty" jsonschema:"description:Optional condition"`
}
```

### JSON Schema Tags

The `jsonschema` tags provide validation and documentation:

- `description`: Human-readable description
- `minimum`/`maximum`: Numeric bounds
- `enum`: Allowed string values
- `default`: Default value (documentation only)

## Tool Handler Functions

### Basic Handler Pattern

```go
func handleToolName(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[YourParams]) (*mcp.CallToolResultFor[any], error) {
    args := params.Arguments
    
    // Perform tool logic here
    result := doSomething(args)
    
    // Return success response
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{
            Text: fmt.Sprintf("Operation completed: %s", result),
        }},
    }, nil
}
```

### Error Handling

```go
func handleWithErrors(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[YourParams]) (*mcp.CallToolResultFor[any], error) {
    args := params.Arguments
    
    // Validate inputs
    if args.Value < 0 {
        return nil, fmt.Errorf("value must be non-negative")
    }
    
    // Handle operation errors
    result, err := riskyOperation(args)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }
    
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{Text: result}},
    }, nil
}
```

### Multi-Content Responses

```go
func handleMultiContent(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[YourParams]) (*mcp.CallToolResultFor[any], error) {
    // Return multiple content items
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{
            &mcp.TextContent{Text: "First result"},
            &mcp.TextContent{Text: "Second result"},
        },
    }, nil
}
```

## Server Initialization

### Basic Server Setup

```go
func StartMCPServer() error {
    // Create a new server with metadata
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "your-service",
        Version: "1.0.0",
    }, nil)
    
    // Register tools
    mcp.AddTool(server, &mcp.Tool{
        Name:        "tool_name",
        Description: "What this tool does",
    }, handleToolName)
    
    // Run the server on stdio
    return server.Run(context.Background(), mcp.NewStdioTransport())
}
```

### Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Parameter definitions
type CalculateParams struct {
    A        float64 `json:"a" jsonschema:"description:First number"`
    B        float64 `json:"b" jsonschema:"description:Second number"`
    Operation string `json:"operation" jsonschema:"description:Math operation,enum:[add,subtract,multiply,divide]"`
}

// Handler function
func handleCalculate(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[CalculateParams]) (*mcp.CallToolResultFor[any], error) {
    args := params.Arguments
    var result float64
    
    switch args.Operation {
    case "add":
        result = args.A + args.B
    case "subtract":
        result = args.A - args.B
    case "multiply":
        result = args.A * args.B
    case "divide":
        if args.B == 0 {
            return nil, fmt.Errorf("division by zero")
        }
        result = args.A / args.B
    default:
        return nil, fmt.Errorf("unknown operation: %s", args.Operation)
    }
    
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{
            Text: fmt.Sprintf("%.2f %s %.2f = %.2f", args.A, args.Operation, args.B, result),
        }},
    }, nil
}

// Server initialization
func StartCalculatorMCP() error {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "calculator",
        Version: "1.0.0",
    }, nil)
    
    mcp.AddTool(server, &mcp.Tool{
        Name:        "calculate",
        Description: "Perform basic math operations",
    }, handleCalculate)
    
    log.Println("Starting Calculator MCP server...")
    return server.Run(context.Background(), mcp.NewStdioTransport())
}

func main() {
    if err := StartCalculatorMCP(); err != nil {
        log.Fatal(err)
    }
}
```

## Advanced Patterns

### 1. State Management

```go
type ServerState struct {
    // Your state fields
    connections map[string]*Connection
    mu          sync.RWMutex
}

var state = &ServerState{
    connections: make(map[string]*Connection),
}

func handleStatefulOperation(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[YourParams]) (*mcp.CallToolResultFor[any], error) {
    state.mu.Lock()
    defer state.mu.Unlock()
    
    // Access and modify state safely
    state.connections[params.Arguments.ID] = newConnection
    
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{Text: "State updated"}},
    }, nil
}
```

### 2. External API Integration

```go
func handleAPICall(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[APIParams]) (*mcp.CallToolResultFor[any], error) {
    // Create HTTP client
    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    
    // Make API request
    resp, err := client.Get(params.Arguments.URL)
    if err != nil {
        return nil, fmt.Errorf("API request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Process response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{Text: string(body)}},
    }, nil
}
```

### 3. Async Operations

```go
func handleAsyncOperation(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[AsyncParams]) (*mcp.CallToolResultFor[any], error) {
    // Start async operation
    done := make(chan string)
    
    go func() {
        // Simulate long-running operation
        time.Sleep(2 * time.Second)
        done <- "Operation completed"
    }()
    
    // Wait with timeout
    select {
    case result := <-done:
        return &mcp.CallToolResultFor[any]{
            Content: []mcp.Content{&mcp.TextContent{Text: result}},
        }, nil
    case <-time.After(5 * time.Second):
        return nil, fmt.Errorf("operation timed out")
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### 4. Binary Data Handling

```go
type FileUploadParams struct {
    Filename string `json:"filename" jsonschema:"description:Name of the file"`
    Data     string `json:"data" jsonschema:"description:Base64-encoded file data"`
}

func handleFileUpload(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[FileUploadParams]) (*mcp.CallToolResultFor[any], error) {
    // Decode base64 data
    decoded, err := base64.StdEncoding.DecodeString(params.Arguments.Data)
    if err != nil {
        return nil, fmt.Errorf("invalid base64 data: %w", err)
    }
    
    // Process binary data
    if err := os.WriteFile(params.Arguments.Filename, decoded, 0644); err != nil {
        return nil, fmt.Errorf("failed to write file: %w", err)
    }
    
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{
            Text: fmt.Sprintf("File saved: %s (%d bytes)", params.Arguments.Filename, len(decoded)),
        }},
    }, nil
}
```

## Best Practices

### 1. Input Validation

Always validate inputs before processing:

```go
func validateAndProcess(args *YourParams) error {
    // Range validation
    if args.Port < 1 || args.Port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    
    // String validation
    if args.Name == "" {
        return fmt.Errorf("name cannot be empty")
    }
    
    // Array validation
    if len(args.Items) == 0 {
        return fmt.Errorf("at least one item required")
    }
    
    return nil
}
```

### 2. Error Messages

Provide clear, actionable error messages:

```go
// Good
return nil, fmt.Errorf("file not found at path: %s", path)

// Better
return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
```

### 3. Logging

Use appropriate logging for debugging:

```go
func handleWithLogging(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[YourParams]) (*mcp.CallToolResultFor[any], error) {
    log.Printf("Handling request: %+v", params.Arguments)
    
    result, err := processRequest(params.Arguments)
    if err != nil {
        log.Printf("Error processing request: %v", err)
        return nil, err
    }
    
    log.Printf("Request successful: %s", result)
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{Text: result}},
    }, nil
}
```

### 4. Resource Cleanup

Always clean up resources:

```go
func handleWithCleanup(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[YourParams]) (*mcp.CallToolResultFor[any], error) {
    // Open resource
    resource, err := openResource(params.Arguments.Path)
    if err != nil {
        return nil, err
    }
    defer resource.Close() // Ensure cleanup
    
    // Use resource
    result, err := resource.Process()
    if err != nil {
        return nil, fmt.Errorf("processing failed: %w", err)
    }
    
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{Text: result}},
    }, nil
}
```

## Testing and Debugging

### 1. Unit Testing Handlers

```go
func TestHandleCalculate(t *testing.T) {
    tests := []struct {
        name    string
        params  CalculateParams
        want    string
        wantErr bool
    }{
        {
            name: "addition",
            params: CalculateParams{A: 5, B: 3, Operation: "add"},
            want: "5.00 add 3.00 = 8.00",
        },
        {
            name: "division by zero",
            params: CalculateParams{A: 5, B: 0, Operation: "divide"},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := handleCalculate(context.Background(), nil, &mcp.CallToolParamsFor[CalculateParams]{
                Arguments: &tt.params,
            })
            
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr && result.Content[0].(*mcp.TextContent).Text != tt.want {
                t.Errorf("got %v, want %v", result.Content[0].(*mcp.TextContent).Text, tt.want)
            }
        })
    }
}
```

### 2. Manual Testing

Create a test client to interact with your MCP server:

```bash
# Start your MCP server
go run .

# In another terminal, use the MCP test client
mcp-client stdio ./your-mcp-server

# Or test with curl if using HTTP transport
curl -X POST http://localhost:8080/tools/calculate \
  -H "Content-Type: application/json" \
  -d '{"a": 5, "b": 3, "operation": "add"}'
```

### 3. Debug Logging

Enable verbose logging during development:

```go
func init() {
    if os.Getenv("MCP_DEBUG") == "true" {
        log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    }
}
```

## Conclusion

This guide covers the essential patterns for building MCP servers in Go. The key steps are:

1. Define parameter structs with proper JSON schema tags
2. Implement handler functions with proper error handling
3. Initialize the server and register tools
4. Follow best practices for validation, logging, and resource management

For more examples, refer to the microM8 emulator's MCP implementation or the official MCP Go SDK documentation.