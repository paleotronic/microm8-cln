#!/usr/bin/env node

const readline = require('readline');
const http = require('http');

// Create readline interface for stdio communication
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

// Buffer for incomplete JSON messages
let buffer = '';

// Handle incoming messages from stdio
rl.on('line', (line) => {
  buffer += line;
  
  try {
    const message = JSON.parse(buffer);
    buffer = '';
    
    // Forward the message to the HTTP MCP server
    const options = {
      hostname: '127.0.0.1',
      port: 1979,
      path: '/',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(JSON.stringify(message))
      }
    };
    
    const req = http.request(options, (res) => {
      let responseData = '';
      
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      
      res.on('end', () => {
        // Send response back through stdio
        process.stdout.write(responseData + '\n');
      });
    });
    
    req.on('error', (error) => {
      const errorResponse = {
        jsonrpc: '2.0',
        id: message.id || null,
        error: {
          code: -32603,
          message: 'Internal error',
          data: error.message
        }
      };
      process.stdout.write(JSON.stringify(errorResponse) + '\n');
    });
    
    req.write(JSON.stringify(message));
    req.end();
    
  } catch (e) {
    // Not a complete JSON message yet, continue buffering
  }
});

// Handle process termination
process.on('SIGINT', () => {
  process.exit(0);
});

process.on('SIGTERM', () => {
  process.exit(0);
});
