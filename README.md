# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

**Selfier** is a Go API service built with Huma v2 framework for handling AI-powered image processing jobs. The application follows clean architecture principles with domain-driven design patterns.

### Tech Stack
- **Language**: Go 1.25.1
- **Web Framework**: Huma v2 (OpenAPI-first REST framework)
- **Database**: PostgreSQL with GORM ORM
- **Configuration**: Viper with environment variable support
- **Logging**: Structured logging with slog
- **Testing**: testify with mocks
- **Linting**: golangci-lint with strict configuration

## Architecture

### High-Level Structure
```
cmd/api/           # Application entry point
internal/module/   # Business logic modules (job, aideselfie, event, notification)
pkg/              # Reusable packages (config, database, logger, middleware, router)
```

### Domain Module Architecture
The application uses ports and adapters pattern:
- **Ports**: Interfaces defined in `job.go`
- **Adapters**: HTTP handlers, repositories, external services
- **Domain Logic**: Service implementations with business rules
- **Models**: Separate domain entities from database models

### Key Modules
1. **Job Module**: Core business logic for AI image processing jobs
   - Handles job lifecycle (create, run, complete, fail)
   - Manages file uploads to object storage
   - Publishes events for async processing
   - Supports multiple job types (currently "deselfie")

2. **Configuration**: Environment-based configuration with validation
3. **Database**: PostgreSQL connection with connection pooling
4. **Middleware**: Request ID, authentication, and logging

## Development Commands

### Building and Running
```bash
# Run the application
go run cmd/api/main.go

# Build the application
go build -o bin/selfier cmd/api/main.go

# Run with specific environment
PRIMARY_ENV=development go run cmd/api/main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific module
go test ./internal/module/job/...

# Run specific test
go test -run TestJobHTTPHandler_CreateJob ./internal/module/job/
```

### Linting and Code Quality
```bash
# Run linter (uses .golangci.yaml config)
golangci-lint run

# Format code
go fmt ./...

# Run imports formatting
goimports -w .

# Check for security issues
gosec ./...
```

### Database Operations
```bash
# The application uses GORM for database operations
# Database migrations are handled through GORM auto-migration
# Connection settings are managed through configuration
```

## Configuration

The application uses a hierarchical configuration system:
1. **Default values** in `pkg/config/config.go`
2. **YAML config file** (optional)
3. **Environment variables** (highest priority)

### Environment Variables
Key environment variables (see `.env.example`):
- `PRIMARY_ENV`: Environment (development/staging/production)
- `SERVER_PORT`: API server port (default: 8080)
- `DATABASE_*`: Database connection settings
- `AWS_*`: AWS/S3 configuration for object storage
- `AUTH_SECRET_KEY`: JWT secret key

### Configuration Sections
- **Primary**: Basic service information
- **Server**: HTTP server configuration with timeouts
- **Database**: PostgreSQL connection and pool settings
- **Auth**: Authentication configuration
- **AWS**: Object storage configuration
- **Logger**: Logging level configuration

## Code Organization Patterns

### Error Handling
- Use wrapped errors with context: `fmt.Errorf("operation failed: %w", err)`
- Define domain-specific errors in `types.go`
- Return appropriate HTTP status codes through Huma

### Dependency Injection
- Constructor functions for all services
- Interface-based dependency injection
- Mock implementations for testing

### Database Models vs Domain Entities
- **Database Models**: `JobModel`, `JobImageModel` (GORM structs)
- **Domain Entities**: `Job`, `JobImage` (business logic structs)
- Conversion functions: `GetJobFromModel()`

### Testing Patterns
- Table-driven tests with test cases
- Mock services using testify/mock
- HTTP handler testing with httptest
- Multipart form testing utilities

### Middleware Chain
```go
api.UseMiddleware(middleware.AuthMiddleware())
api.UseMiddleware(middleware.RequestIDMiddleware())
api.UseMiddleware(middleware.LoggerMiddleware(baseLogger))
```

## API Design

### REST Endpoints
- `POST /jobs` - Create new image processing job
- `GET /jobs` - List all jobs
- `GET /jobs/{id}` - Get specific job
- `DELETE /jobs/{id}` - Delete job
- `GET /jobs/{id}/results` - Get job results
- `GET /health` - Health check

### Request/Response Patterns
- Multipart form uploads for file handling
- Structured JSON responses
- OpenAPI documentation via Huma
- Consistent error response format

### Job Processing Flow
1. Upload image to object storage
2. Create job record in database
3. Publish event for async processing
4. Update job status through lifecycle
5. Store results with presigned URLs

## Development Guidelines

### Code Style
- Follow Go idioms and conventions
- Use meaningful variable and function names
- Add doc comments for public interfaces
- Maintain consistent error handling patterns

### Commit Practices
- Use conventional commit messages
- Include tests with new features
- Update documentation when needed
- Ensure linter passes before commits

### File Organization
- Keep interfaces in module root (`job.go`)
- Separate concerns into focused files
- Use descriptive file names (`types.go`, `service.go`, `handler.go`)
- Group related functionality together

## External Dependencies

### Key Third-Party Libraries
- `github.com/danielgtaylor/huma/v2` - OpenAPI framework
- `gorm.io/gorm` - Database ORM
- `github.com/spf13/viper` - Configuration management
- `github.com/stretchr/testify` - Testing utilities
- `github.com/google/uuid` - UUID generation

### Integration Points
- **PostgreSQL**: Primary data storage
- **AWS S3**: Object storage for images
- **Inngest**: Event publishing (referenced but not fully implemented)
- **Redis**: Caching (configured but usage TBD)

## Troubleshooting

### Common Issues
- **Database connection failures**: Check DATABASE_* environment variables
- **Configuration errors**: Validate required fields in config struct
- **File upload issues**: Verify AWS credentials and bucket permissions
- **Test failures**: Ensure mock setup matches service interfaces

### Debugging Tips
- Enable debug logging: `LOGGER_LEVEL=debug`
- Check database connectivity with ping
- Use structured logging for tracing requests
- Leverage request ID middleware for correlation
