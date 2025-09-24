#!/bin/bash

# Test script for MCP HTTP streaming transport

echo "Testing MCP HTTP Streaming Transport"
echo "====================================="
echo ""

PORT=1977
BASE_URL="http://localhost:$PORT"

echo "1. Testing health endpoint..."
curl -s "$BASE_URL/mcp/health" | python3 -m json.tool
echo ""

echo "2. Testing info endpoint..."
curl -s "$BASE_URL/mcp/info" | python3 -m json.tool
echo ""

echo "3. Testing OPTIONS preflight..."
curl -s -X OPTIONS "$BASE_URL/mcp/stream" \
  -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type, Mcp-Session-Id" \
  -v 2>&1 | grep -E "< HTTP|< Access-Control"
echo ""

echo "4. Testing GET for streaming (will create a new session)..."
echo "   (Press Ctrl+C to stop the stream)"
curl -N "$BASE_URL/mcp/stream" \
  -H "Accept: text/event-stream" \
  -v 2>&1 | head -20

echo ""
echo "Test complete!"
echo ""
echo "To start the server for testing, run:"
echo "  ./microM8 -mcp -mcp-mode streaming -mcp-port $PORT"