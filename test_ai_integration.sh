#!/bin/bash
# Integration test script for AI features

set -e

echo "=== AI Integration Test ==="
echo

# Build the client
echo "Building dikuclient..."
go build ./cmd/dikuclient
echo "✓ Build successful"
echo

# Run tests
echo "Running unit tests..."
go test ./internal/ai -v
go test ./internal/triggers -v -run TestLast
go test ./internal/config -v
echo "✓ All tests passed"
echo

echo "=== Manual Testing Instructions ==="
echo
echo "1. Test <last_command> substitution:"
echo "   ./dikuclient --host localhost --port 4000"
echo "   > /trigger \"Huh?!\" \"say The last command was: <last_command>\""
echo "   > invalidcommand"
echo "   (should output: The last command was: invalidcommand)"
echo
echo "2. Test /configure-ai command:"
echo "   > /configure-ai ollama http://localhost:11434/api/generate"
echo "   (should save AI config)"
echo
echo "3. Test /ai-prompt command:"
echo "   > /ai-prompt preset barsoom"
echo "   (should set prompt to PLACEHOLDER)"
echo
echo "4. Test /ai command (requires ollama running):"
echo "   > /ai what is 2+2"
echo "   (should get AI response and try to send it to MUD)"
echo
echo "5. Test /howto command (requires ollama running):"
echo "   > /howto heal myself"
echo "   (should display AI response without executing)"
echo
echo "6. Test integration with trigger:"
echo "   > /trigger \"Huh?!\" \"/ai <last_command>\""
echo "   > badcommand"
echo "   (should trigger AI suggestion when MUD responds with Huh?!)"
echo
echo "=== Test Complete ==="
