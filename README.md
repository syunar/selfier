# Selfier

An AI-powered image processing application that transforms selfies into natural-looking photos taken from a distance. Perfect for solo travelers who want professional-looking travel photos without a selfie stick or asking strangers for help.

## Overview

Selfier converts close-up selfie photos into images that appear as if they were taken by someone else from a distance. Using advanced AI diffusion models, the service transforms the perspective, composition, and framing to create natural travel photography. The application follows hexagonal architecture principles with a domain-driven design approach.

### Architecture

- **Backend API**: Go service with Huma v2 framework for OpenAPI-first REST endpoints
- **Frontend**: Next.js 15 with React 19, Tailwind CSS, and shadcn/ui components
- **AI Processing**: Modal Labs for serverless GPU inference using diffusion models
- **Event Processing**: Inngest for async job orchestration
- **Database**: PostgreSQL with GORM ORM
- **Storage**: AWS S3 for image uploads and results

## Tech Stack

### Backend

- **Go 1.25.1**
- **Huma v2** - OpenAPI framework
- **GORM** - Database ORM
- **PostgreSQL** - Primary database
- **AWS SDK v2** - S3 integration
- **Inngest Go SDK** - Event orchestration
- **Viper** - Configuration management
- **slog** - Structured logging

### Frontend

- **Next.js 15** with App Router
- **React 19**
- **TypeScript**
- **Tailwind CSS 4**
- **shadcn/ui** - Component library
- **Prisma** - Database client
- **openapi-fetch** - Type-safe API client

### AI/ML

- **Modal Labs** - Serverless GPU platform
- **Qwen Image Edit 2509** - Diffusion model (4-bit quantized)
- **Diffusers** - Hugging Face library
- **PyTorch** - Deep learning framework

## Getting Started

### Prerequisites

#### Option 1: Docker (Recommended)
- Docker 20.10+
- Docker Compose v2.0+
- AWS account (for S3)
- Modal account (for AI processing)

#### Option 2: Local Development
- Go 1.25+
- Node.js 18+
- PostgreSQL 14+
- AWS account (for S3)
- Modal account (for AI processing)

### Environment Setup

1. Copy the example environment file:

```bash
cp .env.example .env
```

2. Configure the following required variables:

```env
# Server
PRIMARY_ENV=development
SERVER_PORT=8080

# Database (for Docker, use defaults; for local, customize)
DATABASE_HOST=localhost  # Use 'db' if running in Docker
DATABASE_PORT=5432
DATABASE_USER=selfier_user
DATABASE_PASSWORD=selfier_password
DATABASE_NAME=selfier_db

# AWS S3 (Required)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_UPLOAD_BUCKET=your-bucket-name

# Authentication (Required)
AUTH_SECRET_KEY=your-jwt-secret-key
```

### Installation

#### Option 1: Docker Installation (Recommended)

The easiest way to run the full stack with all dependencies:

```bash
# Build and start all services (PostgreSQL, Backend, Frontend)
docker-compose up --build

# Or run in detached mode
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v
```

Services will be available at:
- **Backend API**: http://localhost:8080
- **Frontend**: http://localhost:3000
- **PostgreSQL**: localhost:5432

#### Option 2: Local Development Installation

**Backend:**

```bash
# Install Go dependencies
go mod download

# Install Air for hot reload (optional)
go install github.com/air-verse/air@latest
```

**Frontend:**

```bash
cd web
npm install
npm run db:push  # Setup Prisma schema
```

**PostgreSQL:**

```bash
# macOS (using Homebrew)
brew install postgresql@16
brew services start postgresql@16

# Create database
createdb selfier_db

# Linux (Ubuntu/Debian)
sudo apt-get install postgresql-16
sudo systemctl start postgresql
```

#### Modal (AI Processing)

```bash
# Install Modal CLI
pip install modal

# Authenticate
modal token new

# Deploy the deselfie function
modal deploy modal/deselfie.py
```

> **Note**: The TypeScript API types are auto-generated from the OpenAPI spec when running `make dev`. You don't need to generate them manually.

### Running the Application

#### Docker Mode

```bash
# Start all services
docker-compose up

# Access the application
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
# API Docs: http://localhost:8080/docs
```

#### Local Development Mode (All Services)

```bash
make dev
```

This single command will:
1. Start the Go backend with hot reload (Air) on port 8080
2. Wait for the backend to be ready
3. Start the Inngest dev server at http://localhost:8080/api/inngest
4. Auto-generate TypeScript API types from OpenAPI spec to `web/src/lib/api-types.ts`
5. Start the Next.js frontend with Turbo

#### Individual Services

**Backend Only:**

```bash
# With hot reload
air

# Or standard go run
go run cmd/api/main.go
```

**Frontend Only:**

```bash
cd web
npm run dev
```

**Inngest Dev Server:**

```bash
npx inngest-cli@latest dev --no-discovery -u http://localhost:8080/api/inngest
```

## API Documentation

Once the backend is running, visit:

- **OpenAPI Docs**: http://localhost:8080/docs
- **OpenAPI Spec**: http://localhost:8080/openapi.json

### Key Endpoints

- `POST /api/jobs` - Create new image processing job
- `GET /api/jobs` - List all jobs
- `GET /api/jobs/{id}` - Get job details
- `DELETE /api/jobs/{id}` - Delete job
- `GET /api/jobs/{id}/results` - Get processed images
- `GET /health` - Health check

## Project Structure

```
selfier/
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   └── modules/          # Business logic modules
│       ├── job/          # Job processing domain
│       └── ...
├── pkg/                  # Reusable packages
│   ├── config/           # Configuration management
│   ├── database/         # Database connection
│   ├── logger/           # Structured logging
│   ├── router/           # HTTP router setup
│   └── aws/              # AWS S3 client
├── modal/                # Modal AI processing functions
│   ├── deselfie.py       # Main deselfie pipeline
│   └── mock_deselfie.py  # Local testing mock
├── web/                  # Next.js frontend
│   ├── src/
│   │   ├── app/          # App router pages
│   │   ├── components/   # React components
│   │   └── lib/          # Utilities
│   └── prisma/           # Database schema
├── bin/                  # Compiled binaries
├── tmp/                  # Temporary files (Air)
├── .env                  # Environment variables
├── .air.toml             # Hot reload config
├── Makefile              # Development commands
└── go.mod                # Go dependencies
```

## Development

### Testing

**Backend:**

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific module tests
go test ./internal/modules/job/...

# Using Make
make test
```

**Frontend:**

```bash
cd web
npm run lint
npm run typecheck
npm run format:check
```

### Code Quality

**Backend:**

```bash
# Lint
golangci-lint run

# Format
go fmt ./...
goimports -w .
```

**Frontend:**

```bash
cd web
npm run lint:fix
npm run format:write
```

## Deployment

### Docker Deployment (Recommended)

**Production deployment with Docker:**

```bash
# Build images
docker-compose build

# Run in production mode
PRIMARY_ENV=production docker-compose up -d

# View logs
docker-compose logs -f

# Check service health
docker-compose ps
```

Ensure your production `.env` file has:
- Valid AWS credentials
- Secure `AUTH_SECRET_KEY`
- Production database credentials
- Appropriate CORS origins

### Manual Deployment

**Backend:**

Build the Go binary:

```bash
go build -o bin/selfier cmd/api/main.go
```

Run in production:

```bash
PRIMARY_ENV=production ./bin/selfier
```

**Frontend:**

```bash
cd web
npm run build
npm run start
```

**Modal Functions:**

Deploy to Modal:

```bash
modal deploy modal/deselfie.py
```

## Configuration

The application uses a hierarchical configuration system:

1. Default values in code
2. Environment variables (highest priority)
3. Optional YAML config file

See `.env.example` for all available configuration options.

## How It Works

1. **Upload**: Solo traveler uploads a selfie photo
2. **Storage**: Image is uploaded to S3 and a job record is created
3. **Event**: Job creation triggers an Inngest event
4. **Processing**: Modal function downloads the selfie, runs AI model to transform it into a non-selfie perspective, and uploads result
5. **Completion**: Job status is updated with presigned URLs to results
6. **Display**: Frontend displays original selfie and transformed travel photo side-by-side

## Contributing

1. Follow Go idioms and conventions
2. Add tests for new features
3. Use conventional commit messages
4. Ensure linter passes before commits
5. Update documentation as needed

## License

[Add your license here]

## Support

For issues and questions:

- Create an issue in the repository
- Check the [API documentation](http://localhost:8080/docs)
- Review existing issues and pull requests
