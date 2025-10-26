# AI Integration for DikuMUD Parsing Failures - Design Document

## Overview

This document outlines the design for integrating AI assistance into the DikuMUD client to help with parsing failures and provide command suggestions. The system will use Small Language Models (SLMs) or other AI endpoints (OpenAI, Ollama) to provide assistance when the MUD doesn't understand user commands.

## Components

### 1. Trigger Enhancement: `<last_command>` Variable

**Purpose**: Capture the last command sent by the user so it can be referenced in trigger actions.

**Implementation**:
- Add `lastCommand` field to the `Model` struct in `internal/tui/app.go`
- Update the command sending logic to store commands in `lastCommand` before sending
- Extend the trigger action substitution to recognize `<last_command>` as a special placeholder
- When triggers fire, substitute `<last_command>` with the actual last command

**Location**: `internal/tui/app.go`, `internal/triggers/trigger.go`

### 2. AI Configuration Storage

**Purpose**: Store AI endpoint configuration (URL, type, API key) persistently.

**Implementation**:
- Add `AIConfig` struct to `internal/config/account.go`:
  ```go
  type AIConfig struct {
      Type   string `json:"type"`    // "openai", "ollama", or custom
      URL    string `json:"url"`     // API endpoint URL
      APIKey string `json:"-"`       // Not stored in JSON, stored in passwords file
  }
  ```
- Add `AIConfig` field to the `Config` struct
- Add methods to get/set AI configuration
- Store API key separately in the password store for security

**Location**: `internal/config/account.go`, `internal/config/passwords.go`

### 3. AI Prompt Configuration

**Purpose**: Store and manage AI prompts with presets for specific MUDs.

**Implementation**:
- Add `AIPrompt` field to the `Config` struct
- Provide default prompt for Barsoom MUD loaded from `data/presets/barsoom.prompt`
- Allow users to customize the prompt via `/ai-prompt` command
- The prompt should instruct the AI on how to interpret the failed command and suggest alternatives

**Location**: `internal/config/account.go`

### 4. `/configure-ai` Command

**Purpose**: Configure the AI endpoint settings.

**Syntax**: 
```
/configure-ai <type> <url> [api-key]
```

**Examples**:
```
/configure-ai openai https://api.openai.com/v1/chat/completions sk-...
/configure-ai ollama http://localhost:11434/api/generate
```

**Implementation**:
- Add `handleConfigureAICommand` function in `internal/tui/app.go`
- Parse the command arguments
- Store configuration in the `Config` struct
- Save API key to password store if provided
- Provide feedback to user

**Location**: `internal/tui/app.go`

### 5. `/ai-prompt` Command

**Purpose**: Configure the AI prompt template.

**Syntax**:
```
/ai-prompt "<prompt text>"
/ai-prompt preset barsoom
```

**Examples**:
```
/ai-prompt "You are a helpful DikuMUD assistant. The user tried this command but it failed: {command}. Suggest a correct alternative."
/ai-prompt preset barsoom
```

**Implementation**:
- Add `handleAIPromptCommand` function in `internal/tui/app.go`
- Support both custom prompts and presets
- Store in configuration
- Barsoom preset loaded from `data/presets/barsoom.prompt` containing comprehensive DikuMUD command translation rules

**Location**: `internal/tui/app.go`

### 6. `/ai` Command

**Purpose**: Send a prompt to the AI and execute the suggested command.

**Syntax**:
```
/ai <prompt>
```

**Example**:
```
/ai <last_command>
```

**Implementation**:

#### CLI Mode (local Go process):
- Add `handleAICommand` function in `internal/tui/app.go`
- Create AI client in `internal/ai/client.go`:
  - Support OpenAI API format
  - Support Ollama API format
  - Make HTTP request to configured endpoint
  - Parse response and extract suggested command
- Send the suggested command to the MUD
- Display the AI's response/suggestion to the user

#### Web Mode (browser makes request):
- Detect web mode using existing `webSessionID` field
- If in web mode, send a special message to the web client via WebSocket
- Extend `internal/web/websocket.go` to handle AI request messages
- Browser-side JavaScript (in `web/static/app.js`):
  - Receive AI request message
  - Make HTTP request to AI endpoint from browser
  - Send response back via WebSocket
  - TUI receives response and sends command to MUD

**Rationale for Web Mode**: In web mode, the browser makes the AI request because:
1. The API key should not be sent to the server
2. CORS policies may restrict server-side requests
3. The user's browser can handle authentication directly

**Location**: `internal/tui/app.go`, `internal/ai/client.go` (new), `internal/web/websocket.go`, `web/static/app.js`

### 7. `/howto` Command

**Purpose**: Ask the AI how to do something, but just display the answer (don't execute).

**Syntax**:
```
/howto <query>
```

**Example**:
```
/howto heal myself
```

**Implementation**:
- Similar to `/ai` command but:
  - Display AI response as informational output (like trigger matches)
  - Do not automatically send any commands to the MUD
  - Format output in the same style as trigger match notifications

**Location**: `internal/tui/app.go`

### 8. Integration Example

**Setup**:
```
/configure-ai ollama http://localhost:11434/api/generate
/ai-prompt preset barsoom
/trigger "Huh?!" "/ai <last_command>"
```

**Usage Flow**:
1. User types: `heall` (misspelled command)
2. MUD responds: `Huh?!`
3. Trigger fires, capturing "heall" as `<last_command>`
4. Trigger executes: `/ai heall`
5. AI receives prompt with the failed command
6. AI suggests: `heal`
7. Client sends `heal` to the MUD
8. User sees: `[Trigger: /ai heall]` and `[AI suggests: heal]`

## File Structure

### New Files:
- `internal/ai/client.go` - AI client implementation for CLI mode
- `internal/ai/client_test.go` - Tests for AI client

### Modified Files:
- `internal/config/account.go` - Add AI configuration storage
- `internal/config/passwords.go` - Store API keys securely
- `internal/triggers/trigger.go` - Support `<last_command>` substitution
- `internal/tui/app.go` - Add new commands and lastCommand tracking
- `internal/web/websocket.go` - Handle AI requests in web mode
- `web/static/app.js` - Handle AI requests from browser

## Implementation Plan

### Phase 1: Core Infrastructure
1. Add `lastCommand` tracking to Model
2. Extend trigger system to support `<last_command>`
3. Add AI configuration to Config struct
4. Implement `/configure-ai` command
5. Implement `/ai-prompt` command

### Phase 2: AI Client (CLI Mode)
1. Create `internal/ai/client.go` with support for:
   - OpenAI API format
   - Ollama API format
2. Implement `/ai` command for CLI mode
3. Implement `/howto` command for CLI mode

### Phase 3: Web Mode Support
1. Extend WebSocket protocol for AI requests
2. Update browser JavaScript to handle AI requests
3. Implement `/ai` and `/howto` for web mode

### Phase 4: Testing and Documentation
1. Add unit tests for AI client
2. Add integration tests for commands
3. Test trigger integration: `/trigger "Huh?!" "/ai <last_command>"`
4. Update README.md with AI features

## Security Considerations

1. **API Key Storage**: Store API keys in the password store, not in accounts.json
2. **Web Mode**: Browser makes AI requests to prevent exposing API keys to the server
3. **Input Validation**: Sanitize prompts before sending to AI
4. **Rate Limiting**: Consider adding rate limiting to prevent API abuse

## Error Handling

1. Network errors: Display error message, don't crash
2. Invalid AI response: Show error, don't send bad commands
3. Missing configuration: Prompt user to run `/configure-ai`
4. API key errors: Clear error message about authentication

## Future Enhancements

1. Support for additional AI providers (Anthropic, etc.)
2. Command history analysis for better suggestions
3. Context-aware prompts (current room, inventory, etc.)
4. Caching of common suggestions
5. Multi-turn conversations with AI
