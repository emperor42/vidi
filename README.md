# VIDI

VIDI is a small, dependency-free browser library for rendering records from a JSON or XML endpoint as escaped cards, with pagination and CSV export.

The root of this component contains two separate things:

- **`vidi.js`** is the browser API. It does not need a Node.js runtime, a bundler, or a Go server.
- **`main.go`, `static/`, and `templates/`** are an optional Go template-engine

## Overview

`vidi` is a Go backend service that provides record rendering APIs, serving as the foundation for the VIDI browser library.

## Features

- **Record Rendering**: API for rendering records from JSON/XML endpoints
- **Pagination**: Built-in pagination support
- **CSV Export**: Export records as CSV
- **Health Monitoring**: Built-in health check endpoint
- **Security First**: Implements security best practices

## API Endpoints

### Record API
- **GET /api/records** - List all records with optional filtering
- **POST /api/records** - Create a new record

### System Status
- **GET /health** - Health check endpoint

## Technology Stack

- **Backend**: Go
- **API**: RESTful JSON
- **Security**: HTTP security headers, rate limiting, validation
- **Deployment**: Docker ready

## Running

### Prerequisites

```bash
go 1.21+
```

### Building and Running

```bash
# Build the application
go build

# Run the application
./vidi

# Or using Docker
docker build -t vidi .
docker run -p 8084:8084 vidi
```

### Testing

```bash
# Run tests
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Usage

### Local Development

```bash
# Start the server
./vidi

# List all records
curl http://localhost:8084/api/records

# Create a new record
curl -X POST -H "Content-Type: application/json" -d '{"title":"New Record","content":"Record content"}' http://localhost:8084/api/records

# Search records
curl "http://localhost:8084/api/records?title=search"

# Check system health
curl http://localhost:8084/health
```

### API Examples

#### Get All Records
```bash
curl -X GET http://localhost:8084/api/records
```

#### Get Filtered Records
```bash
curl -X GET "http://localhost:8084/api/records?title=search"
```

#### Create New Record
```bash
curl -X POST -H "Content-Type: application/json" -d '{"title":"New Record","content":"Record content","id":"custom-id"}' http://localhost:8084/api/records
```

#### Health Check
```bash
curl -X GET http://localhost:8084/health
```

## Deployment

### Docker Deployment

```Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
COPY . .

RUN go build -o vidi main.go

EXPOSE 8084

CMD ["./vidi"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vidi
  labels:
    app: vidi
spec:
  replicas: 2
  selector:
    matchLabels:
      app: vidi
  template:
    metadata:
      labels:
        app: vidi
    spec:
      containers:
      - name: vidi
        image: vidi:latest
        ports:
        - containerPort: 8084
        env:
        - name: NODE_ENV
          value: "production"
```

## Development

### Project Structure

```
.
├── main.go          # Main application
├── main_test.go     # Tests
├── go.mod          # Module configuration
└── README.md       # Documentation
```

### Features Implemented

- **Record Rendering**: API for managing records with JSON/XML endpoints
- **Pagination**: Built-in pagination support
- **Caching**: Performance optimization with caching
- **Security First**: Request validation and security
- **API-First**: RESTful API design
- **Developer Experience**: Comprehensive testing

## License

MIT License

## Future Enhancements

- [ ] Add record categorization
- [ ] Implement CSV export functionality
- [ ] Add real-time updates
- [ ] Implement record attachments
- [ ] Add user authentication
- [ ] Implement batch operations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add tests
5. Submit a pull request

## Contact

For questions or suggestions, please open an issue or contact the maintainer.