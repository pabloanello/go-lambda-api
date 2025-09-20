# Go Lambda API

![Build Status](https://github.com/pabloanello/go-lambda-api/actions/workflows/go.yml/badge.svg)


A serverless REST API built in Go, designed to run on AWS Lambda with DynamoDB as the backend. The project supports both local development and deployment to AWS using the Server
A boilerplate project for building serverless REST APIs using AWS Lambda, API Gateway, and Go.

## Features

- **AWS Lambda**: Deployable as a Lambda function.
- **API Gateway**: Easily connect to HTTP endpoints.
- **Go Modules**: Dependency management with Go modules.
- **Structured Logging**: Uses standard logging for observability.
- **Environment Variables**: Configuration via environment variables.
- **Unit Testing**: Includes basic unit tests.
- **CI/CD**: GitHub Actions workflow for build and test.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.18 or newer
- [AWS CLI](https://aws.amazon.com/cli/)
- [Docker](https://www.docker.com/) (for local testing)
- [Serverless Framework](https://www.serverless.com/) (optional)

### Installation

Clone the repository:

```sh
git clone https://github.com/pabloanello/go-lambda-api.git
cd go-lambda-api
```

Install dependencies:

```sh
go mod tidy
```

### Local Development

You can run the API locally:

```sh
go run main.go
```

Or using [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli.html):

```sh
sam local start-api
```

### Project Structure

```
go-lambda-api/
├── main.go
├── handler.go
├── go.mod
├── go.sum
├── README.md
├── tests/
│   └── handler_test.go
└── ...
```

- `main.go`: Lambda entry point.
- `handler.go`: API handler logic.
- `tests/`: Unit tests.

### Deployment

#### Using AWS CLI

1. Build the binary for Linux:

    ```sh
    GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
    zip function.zip bootstrap
    ```

2. Create or update your Lambda function:

    ```sh
    aws lambda create-function --function-name go-lambda-api \
      --zip-file fileb://function.zip --handler bootstrap --runtime provided.al2 \
      --role arn:aws:iam::<account-id>:role/<lambda-execution-role>
    ```

#### Using Serverless Framework

If you use Serverless Framework, add a `serverless.yml` and deploy:

```sh
serverless deploy
```

### Environment Variables

- `LOG_LEVEL`: Set the log level (default: `info`)
- `API_STAGE`: API Gateway stage (optional)

### Testing

Run unit tests:

```sh
go test ./...
```

### API Documentation

#### Health Check

- **GET** `/health`
  - Response: `{ "message": "Health Check OK" }`

#### Users

- **GET** `/users`
  - List all users.
  - Response: Array of user objects.

- **POST** `/users`
  - Create a new user.
  - Request body: `{ "name": "string", "email": "string" }`
  - Response: Created user object.

- **GET** `/users/{id}`
  - Get user by ID.
  - Response: User object or error.

- **PUT** `/users/{id}`
  - Update user by ID.
  - Request body: `{ "name": "string", "email": "string" }` (at least one field required)
  - Response: Updated user object.

- **DELETE** `/users/{id}`
  - Delete user by ID.
  - Response: No content.

#### Error Response Format

All errors return JSON:

```json
{ "error": "error message" }
```

### Example Requests

Get health:
```sh
curl -X GET https://<api-id>.execute-api.<region>.amazonaws.com/dev/health
```

Create user:
```sh
curl -X POST https://<api-id>.execute-api.<region>.amazonaws.com/dev/users \
  -H 'Content-Type: application/json' \
  -d '{"name": "Alice", "email": "alice@example.com"}'
```

### Local Development with Docker

You can run the API in a Docker container for local testing:

```sh
docker build -t go-lambda-api .
docker run -p 9000:8080 go-lambda-api
```

### Advanced Configuration

- `LOG_LEVEL`: Set the log level (default: `info`)
- `API_STAGE`: API Gateway stage (optional)

### Troubleshooting

- **CORS Issues**: Ensure your client allows the correct headers and methods. CORS is enabled by default in the API responses.
- **Lambda Timeout**: Increase the `timeout` in `serverless.yml` if requests take too long.
- **Dependency Issues**: Run `go mod tidy` to resolve Go module problems.

### Continuous Integration

GitHub Actions workflow runs on every push and PR to ensure build and tests pass.

### Contributing

Pull requests are welcome! For major changes, please open an issue first.

### Contact & Support

For questions, issues, or feature requests, please open an issue on [GitHub](https://github.com/pabloanello/go-lambda-api/issues).

### License

[MIT](LICENSE)

---

**Author:** Pablo Anello  
**Repository:** https://github.com/pabloanello/go-lambda-api