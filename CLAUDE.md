# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Secfarm is a CLI-based LLM agent for security alert analysis. It receives security alerts in JSON format (e.g., from Amazon GuardDuty), analyzes them using Gemini, and provides interactive analysis capabilities with external tool integration.

**Key Features:**
- Accept and parse security alerts from various sources
- Generate summaries and extract IOCs (Indicators of Compromise) using Gemini
- Create embeddings with Gemini API for similarity search
- Interactive chat-based analysis with Tool Call loop
- Policy-based filtering using OPA/Rego
- Store alerts and conversation history in Google Cloud (Firestore + Cloud Storage)

## Restriction & Rules

### Export Policy

**IMPORTANT**: Keep symbols (types, functions, variables) private (lowercase) by default throughout the entire project. Only export (capitalize) when absolutely necessary for external package usage.

**Guidelines:**
- Within a package (e.g., `pkg/cli/`), prefer private types and functions
- Only export symbols that are genuinely needed by other packages
- Config structs, helper functions, and internal utilities should be private
- Review each export: "Does another package really need this?"

**Examples:**
- ✅ `type config struct` in `pkg/cli/config.go` (only used within pkg/cli)
- ✅ `func globalFlags(cfg *config)` in `pkg/cli/config.go` (only used within pkg/cli)
- ❌ `type Config struct` when only used internally
- ❌ `func GlobalFlags()` when only called within the same package

### CLI Implementation Rules

**CRITICAL**: When implementing CLI commands in `pkg/cli/`, strictly follow these patterns:

1. **Environment Variable Support**: ALL CLI options MUST support environment variables using `cli/v3`'s `Sources` feature with `SECFARM_` prefix:
   ```go
   &cli.StringFlag{
       Name:        "alert-id",
       Sources:     cli.EnvVars("SECFARM_ALERT_ID"),  // Always use SECFARM_ prefix
       Destination: &alertID,
   }
   ```

   **Naming Convention**: All environment variables must use the `SECFARM_` prefix for consistency:
   - ✅ `SECFARM_PROJECT`, `SECFARM_CLAUDE_API_KEY`, `SECFARM_ALERT_ID`
   - ❌ `GOOGLE_CLOUD_PROJECT`, `CLAUDE_API_KEY`, `ALERT_ID`

2. **Destination Pattern**: ALWAYS use the `Destination` field to store flag values. NEVER use `c.String()`, `c.Bool()`, `c.Int()` etc. to retrieve values:
   ```go
   // ✅ CORRECT
   var alertID string
   &cli.StringFlag{
       Name:        "alert-id",
       Destination: &alertID,
   }
   // Then use alertID directly in Action

   // ❌ WRONG
   alertID := c.String("alert-id")  // NEVER do this
   ```

3. **Args Prohibition**: NEVER use `c.Args()` to get command arguments. All parameters MUST be passed via flags with environment variable support:
   ```go
   // ❌ WRONG
   func myCommand() *cli.Command {
       return &cli.Command{
           ArgsUsage: "<alert-id>",
           Action: func(ctx context.Context, c *cli.Command) error {
               alertID := c.Args().Get(0)  // NEVER do this
           },
       }
   }

   // ✅ CORRECT
   func myCommand() *cli.Command {
       var alertID string
       flags := []cli.Flag{
           &cli.StringFlag{
               Name:        "alert-id",
               Sources:     cli.EnvVars("SECFARM_ALERT_ID"),
               Destination: &alertID,
               Required:    true,
           },
       }
       return &cli.Command{
           Flags: flags,
           Action: func(ctx context.Context, c *cli.Command) error {
               // Use alertID directly
           },
       }
   }
   ```

### 3rd party packages
- **CLI Framework**: `github.com/urfave/cli/v3` - ALL CLI and environment variable handling
- **Logging**: `slog` with `github.com/gaebalai/clog` for console output
- **Error handling**: `github.com/gaebalai/goerr/v2`
- **Testing framework**: `github.com/gaebalai/gt`

## Architecture

The project follows a layered architecture with clear separation of concerns:

```
cmd/secfarm/main.go          # Entry point
pkg/
├── cli/                     # CLI layer: command definitions, argument parsing, DI setup
├── usecase/                 # UseCase layer: business logic orchestration
├── repository/              # Repository layer: data persistence (Firestore)
├── adapter/                 # Adapter layer: external service integration
│   ├── gemini.go           # Gemini API client (LLM chat & embeddings)
│   └── storage.go          # Cloud Storage client (conversation history)
└── model/                   # Model layer: shared data structures
```

**Dependency Flow:** CLI → UseCase → Repository/Adapter (unidirectional, top-down)

### Layer Responsibilities

- **CLI Layer ([pkg/cli/](pkg/cli/))**: Parses command-line arguments, reads environment variables and config files, initializes repositories and adapters, performs dependency injection into use cases. ALL environment variable access must happen in this layer only.
- **UseCase Layer ([pkg/usecase/](pkg/usecase/))**: Implements core business logic, coordinates repositories and adapters, handles the Tool Call loop with Gemini. Receives all dependencies via constructor injection.
- **Repository Layer ([pkg/repository/](pkg/repository/))**: Abstracts data persistence to Firestore, provides interface for alert CRUD operations and vector search.
- **Adapter Layer ([pkg/adapter/](pkg/adapter/))**: Wraps external service APIs (Gemini, Cloud Storage), hides implementation details from upper layers.
- **Model Layer ([pkg/model/](pkg/model/))**: Defines shared data structures (Alert, Attribute, etc.).

## Core Data Models

### Alert Structure

```go
type Alert struct {
    ID          AlertID
    Title       string        // Generated by LLM
    Description string        // Generated by LLM
    Data        any           // Original JSON data
    Attributes  []*Attribute  // Extracted attributes (via policy or LLM)

    CreatedAt  time.Time
    ResolvedAt *time.Time
    Conclusion string
    Note       string
    MergedTo   AlertID
}

type Attribute struct {
    Key   string
    Value string
    Type  AttributeType  // string, number, ip_address, etc.
}
```

### Alert Lifecycle

```
new → Policy Check → LLM Summary → Unanalyzed
  ↓
  → chat (interactive analysis via Tool Call loop)
  ↓
  → resolve (mark as resolved)
  → merge (consolidate with another alert)
```

## LLM Integration Strategy

**Use LLM for:**
- Alert summarization and title generation
- IOC (Indicators of Compromise) extraction
- Interactive analysis and natural language queries
- Log query generation from natural language
- Dynamic tool selection based on context

**Do NOT use LLM for:**
- Deterministic filtering (use OPA/Rego policies instead)
- Regular expression-based pattern matching
- Simple threshold checks
- Bulk data processing (cost/rate limit concerns)
- Final impact/priority decisions (human judgment required)

## Commands

### `new` - Register new alert

```bash
secfarm new -i alert.json
```

1. Parse JSON alert data
2. Run policy evaluation (accept/reject)
3. Generate summary and extract IOCs via Gemini API
4. Generate embedding vector via Gemini API
5. Save to Firestore with generated alert ID

### `chat` - Interactive analysis

```bash
secfarm chat <alert-id>
```

1. Fetch alert from Firestore
2. Load conversation history from Cloud Storage
3. Start Tool Call loop with Gemini API
4. Execute external tools as requested by Gemini
5. Save updated conversation history

### `list` - List alerts

```bash
secfarm list       # Exclude merged alerts
secfarm list -a    # Include merged alerts
```

### `show` - Show alert details

```bash
secfarm show --alert-id <alert-id>
# Or using environment variable
SECFARM_ALERT_ID=abc123 secfarm show
```

Displays detailed information of a specific alert including title, description, attributes, timestamps, and metadata.

### `search` - Search for similar alerts

```bash
secfarm search -q "suspicious login from unknown IP"
secfarm search --query "AWS S3 bucket access denied" --limit 10
```

1. Generate embedding vector from query text via Gemini API
2. Perform vector search in Firestore
3. Return similar alerts ordered by similarity

### `resolve` - Mark alert as resolved

```bash
secfarm resolve --alert-id <alert-id> --conclusion false_positive --note "Verified safe"
# Or using environment variables
SECFARM_ALERT_ID=abc123 SECFARM_RESOLVE_CONCLUSION=false_positive secfarm resolve
```

Available conclusions: `unaffected`, `false_positive`, `true_positive`, `inconclusive`

### `merge`/`unmerge` - Consolidate similar alerts

```bash
secfarm merge --source-id <source-id> --target-id <target-id>
# Or using environment variables
SECFARM_MERGE_SOURCE_ID=abc123 SECFARM_MERGE_TARGET_ID=def456 secfarm merge

secfarm unmerge --alert-id <alert-id>
# Or using environment variable
SECFARM_ALERT_ID=abc123 secfarm unmerge
```

## Development Commands

### Build and Run

```bash
# Build binary
go build -o secfarm ./cmd/secfarm

# Run directly without build
go run ./cmd/secfarm --help
go run ./cmd/secfarm new -i testdata/alert.json

# Run with environment variables
SECFARM_FIRESTORE_PROJECT=your-project SECFARM_GEMINI_PROJECT=your-project go run ./cmd/secfarm new -i alert.json
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./pkg/usecase

# Run specific test
go test -run TestCreateAlert ./pkg/usecase

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (if configured)
golangci-lint run
```

## External Dependencies

- **Gemini API (Google GenAI)**: Main LLM for analysis and tool orchestration, and embedding generation
- **Firestore**: Alert storage and vector search
- **Cloud Storage**: Conversation history persistence
- **OPA/Rego**: Policy-based alert filtering (implementation in progress)

## Environment Setup

All environment variables use the `SECFARM_` prefix for consistency.

### Global Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SECFARM_LOG_LEVEL` | Log level (debug, info, warn, error) | "info" | No |
| `SECFARM_VERBOSE` | Enable verbose mode (show stack traces) | false | No |
| `SECFARM_FIRESTORE_PROJECT` | Google Cloud project ID for Firestore | - | Yes |
| `SECFARM_FIRESTORE_DATABASE_ID` | Firestore database ID | "(default)" | No |
| `SECFARM_STORAGE_BUCKET` | Cloud Storage bucket name | - | Yes (for chat) |
| `SECFARM_STORAGE_PREFIX` | Cloud Storage object key prefix | - | No |

### LLM Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SECFARM_GEMINI_PROJECT` | Google Cloud project ID for Gemini | - | Yes |
| `SECFARM_GEMINI_LOCATION` | Google Cloud location for Gemini | "us-central1" | No |

### Command-Specific Variables

| Variable | Command | Description |
|----------|---------|-------------|
| `SECFARM_INPUT` | new | Input file path |
| `SECFARM_ALERT_ID` | show, chat, resolve, unmerge | Alert ID |
| `SECFARM_LIST_ALL` | list | Include merged alerts |
| `SECFARM_LIST_OFFSET` | list | Pagination offset |
| `SECFARM_LIST_LIMIT` | list | Maximum results |
| `SECFARM_SEARCH_QUERY` | search | Natural language query |
| `SECFARM_SEARCH_LIMIT` | search | Maximum results |
| `SECFARM_RESOLVE_CONCLUSION` | resolve | Conclusion type |
| `SECFARM_RESOLVE_NOTE` | resolve | Additional note |
| `SECFARM_MERGE_SOURCE_ID` | merge | Source alert ID |
| `SECFARM_MERGE_TARGET_ID` | merge | Target alert ID |

### Additional Setup

- ADC (Application Default Credentials) for GCP services: `gcloud auth application-default login`

## Tool Call Loop Pattern

The chat command uses Gemini's Tool Use (Function Calling) pattern:

1. Send prompt + tool definitions to Gemini
2. Receive function call instruction from Gemini
3. Execute the requested tool
4. Send function result back to Gemini
5. Repeat until Gemini returns final answer (no more function calls)

This enables Gemini to dynamically call external APIs like threat intelligence services, log databases, etc.
