module go-lambda-api

go 1.22

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go v1.55.8
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)


replace go-lambda-api/lambdarouter => ./cmd/lambda
