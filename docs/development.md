# Voyage Development Guidelines

This document provides guidelines and best practices for developing the Voyage travel plan aggregator.

## Development Environment Setup

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Notmuch email indexer
- isync/mbsync for email synchronization

### Local Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/voyage.git
   cd voyage
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up Notmuch (if not already installed):
   ```bash
   # For Debian/Ubuntu
   sudo apt-get install notmuch
   
   # For macOS
   brew install notmuch
   ```

4. Set up isync/mbsync:
   ```bash
   # For Debian/Ubuntu
   sudo apt-get install isync
   
   # For macOS
   brew install isync
   ```

5. Configure your development environment:
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

6. Run the development server:
   ```bash
   go run cmd/api/main.go
   ```

## Project Structure

The project follows a standard Go project layout:

```
voyage/
├── docs/                       # Documentation
│   ├── architecture.md         # System architecture overview
│   ├── api.md                  # API documentation
│   └── development.md          # Development guidelines
├── cmd/                        # Application entry points
│   └── api/                    # API server
│       └── main.go             # Main application
├── internal/                   # Private application code
│   ├── config/                 # Configuration handling
│   ├── email/                  # Email processing
│   │   ├── fetcher/            # Email fetching (isync integration)
│   │   ├── parser/             # Email parsing logic
│   │   └── notmuch/            # Notmuch integration
│   ├── models/                 # Data models
│   ├── api/                    # API handlers
│   │   ├── middleware/         # API middleware
│   │   └── routes/             # API routes
│   └── storage/                # Storage interfaces
├── pkg/                        # Public libraries
│   └── traveldata/             # Travel data extraction utilities
├── scripts/                    # Utility scripts
├── docker/                     # Docker configuration
│   ├── api/                    # API service
│   └── mail/                   # Mail fetching service
├── docker-compose.yml          # Docker Compose configuration
├── go.mod                      # Go module definition
└── README.md                   # Project overview
```

## Coding Standards

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` or `goimports` to format your code
- Aim for clear, idiomatic Go code
- Document all exported functions, types, and constants

### Error Handling

- Use descriptive error messages
- Wrap errors with context using `fmt.Errorf("context: %w", err)` or a similar approach
- Return errors rather than panicking in library code
- Log errors appropriately at the application boundaries

### Logging

- Use structured logging (e.g., with [zerolog](https://github.com/rs/zerolog) or [zap](https://github.com/uber-go/zap))
- Include relevant context in log entries
- Use appropriate log levels (debug, info, warn, error)
- Avoid logging sensitive information

## Testing

### Unit Tests

- Write unit tests for all packages
- Aim for high test coverage, especially for critical components
- Use table-driven tests where appropriate
- Use mocks or stubs for external dependencies

Example:

```go
func TestEmailParser(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected *TravelData
        wantErr  bool
    }{
        {
            name:  "valid flight email",
            input: testdata.FlightConfirmationEmail,
            expected: &TravelData{
                Type:     "flight",
                Provider: "Test Airlines",
                // ...
            },
            wantErr: false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewEmailParser()
            result, err := parser.Parse(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Integration Tests

- Write integration tests for API endpoints
- Test the interaction with Notmuch and the database
- Use a test database or mock the database layer

## Working with Notmuch

Notmuch is used for email indexing and searching. Here are some guidelines for working with it:

### Querying Emails

Use the notmuch Go bindings to query emails:

```go
// Example: Query emails with a specific tag
db, err := notmuch.Open("/path/to/mail", notmuch.DBReadOnly)
if err != nil {
    return err
}
defer db.Close()

query := db.NewQuery("tag:voyage-trip-123")
messages, err := query.Messages()
if err != nil {
    return err
}

for messages.Valid() {
    message := messages.Get()
    // Process message
    messages.MoveToNext()
}
```

### Tagging Emails

To associate an email with a trip, add a tag to it:

```go
// Example: Tag an email with a trip ID
db, err := notmuch.Open("/path/to/mail", notmuch.DBReadWrite)
if err != nil {
    return err
}
defer db.Close()

message, err := db.FindMessage("message-id")
if err != nil {
    return err
}

err = message.AddTag("voyage-trip-123")
if err != nil {
    return err
}
```

## Email Processing Pipeline

The email processing pipeline is a critical component of Voyage. Here are some guidelines for working with it:

### Email Classification

When classifying emails as travel-related:

- Look for common travel-related keywords in the subject and body
- Check for sender domains associated with travel providers
- Consider using machine learning for more accurate classification

### Data Extraction

When extracting travel data from emails:

- Create parsers for common email formats from major travel providers
- Extract key information like dates, locations, confirmation numbers, etc.
- Handle different email formats (HTML, plain text, etc.)
- Consider using regular expressions or HTML parsing for extraction

### Trip Aggregation

When aggregating travel items into trips:

- Group items by date proximity
- Consider location proximity for items close in time
- Allow manual adjustments to trip groupings

## Docker Deployment

The application is designed to be deployed using Docker Compose:

```bash
# Build and start the containers
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the containers
docker-compose down
```

## Continuous Integration

- Write CI workflows for automated testing
- Include linting and static analysis
- Automate Docker image building
- Consider using GitHub Actions or a similar CI service

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Versioning

Use [Semantic Versioning](https://semver.org/) for the project:

- MAJOR version for incompatible API changes
- MINOR version for backward-compatible functionality additions
- PATCH version for backward-compatible bug fixes
