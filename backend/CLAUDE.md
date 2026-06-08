# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 📋 Project Overview

This is the Go-based backend for the Parv Creations headless e-commerce system. The backend handles core business logic including inventory management, order processing, payment validation, and administrative functions. It connects to a MongoDB database and provides API endpoints for the frontend storefront (Next.js) to consume.

## ⚙️ Development Commands

### Environment Setup
1. Ensure `../.env` exists with the required variables (copy from existing if needed):
   ```
   PORT=8080
   MONGO_PRODUCTION_URI=mongodb+srv://<username>:<password>@<cluster>.mongodb.net/parv_creation?retryWrites=true&w=majority
   ```

### Running the Application
```bash
# From the backend directory
go run ./cmd/api
```

### Building the Binary
```bash
# From the backend directory
go build -o parv-backend ./cmd/api
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests for a specific package
go test ./internal/database
```

### Code Formatting
```bash
# Format Go code
go fmt ./...

# Verify formatting
go vet ./...
```

### Dependency Management
```bash
# Add a new dependency
go get <package@version>

# Tidy up dependencies
go mod tidy

# Check for outdated dependencies
go list -m -u all
```

## 🏗️ Code Architecture

### Directory Structure
```
backend/
├── cmd/
│   └── api/                  # Application entrypoint
│       └── main.go           # HTTP server setup and health check
├── internal/
│   └── database/             # Database connection and operations
│       └── connection.go     # MongoDB client initialization
├── go.mod                    # Go module definition
└── go.sum                    # Dependency checksums
```

### Key Components

#### 1. Application Entrypoint (`cmd/api/main.go`)
- Loads environment variables from `.env` file (located in parent directory)
- Initializes MongoDB connection on startup
- Configures HTTP server with health check endpoint (`/api/v1/health`)
- Graceful shutdown handling

#### 2. Database Layer (`internal/database/connection.go`)
- Reads `MONGO_PRODUCTION_URI` from environment variables
- Configures MongoDB v2 client with connection pooling:
  - MaxPoolSize: 100 connections
  - MinPoolSize: 10 idle connections
  - MaxConnIdleTime: 5 minutes
- Performs connection verification via ping
- Returns initialized client and database handle

#### 3. Environment Configuration
- Environment variables loaded automatically via `init()` function in `main.go`
- Looks for `.env` file in parent directory when running from `backend/` folder
- Required variables:
  - `PORT`: HTTP server port (defaults to 8080)
  - `MONGO_PRODUCTION_URI`: MongoDB connection string

### Data Flow
1. Application starts → Loads environment → Initializes DB connection
2. HTTP server begins listening on configured port
3. Health check endpoint (`/api/v1/health`) returns system status
4. Future API endpoints will be added under `/api/v1/` for:
   - Product management
   - Order processing
   - Inventory control
   - Administrative functions

### Technology Stack
- **Language**: Go 1.26.3
- **Database**: MongoDB Driver v2
- **HTTP Server**: net/http (standard library)
- **Dependency Management**: Go Modules

## 📜 Parv Creations Developer Standards

- **Architecture**: Headless e-commerce platform. Go v1.26.3 backend + MongoDB Driver v2.
- **Directory Root**: Commands must be run relative to the 'backend/' folder workspace.
- **Verification Loop**: Always check compilation status by running `go run cmd/api/main.go` or `go test ./...`.
- **Code Quality**: Adhere strictly to the explicit BSON schemas and design paradigms outlined inside `KT_GUIDE.md`. Do not introduce deprecated mongo v1 structures.

## 🔑 Important Notes

1. **Environment Variables**: The `.env` file must be located in the project root (one level above `backend/` directory) for the application to load it correctly.

2. **Database Connection**: The MongoDB URI should point to a valid MongoDB Atlas cluster or local MongoDB instance. The connection string format is:
   ```
   mongodb+srv://<username>:<password>@<cluster>.mongodb.net/<database>?retryWrites=true&w=majority
   ```

3. **Health Check**: Verify the application is running correctly by accessing:
   ```
   http://localhost:8080/api/v1/health
   ```
   Should return:
   ```json
   {"status":"healthy","database":"connected","project":"Parv Creations Engine"}
   ```

4. **Error Handling**: The application follows explicit error handling patterns - database connection errors are returned rather than causing panics.

5. **Future Extension**: When adding new API endpoints:
   - Add handlers in `cmd/api/main.go` or create separate handler packages
   - Follow the existing pattern for health check endpoint
   - Ensure proper context handling and error propagation