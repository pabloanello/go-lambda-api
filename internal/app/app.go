package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	localLambda "go-lambda-api/cmd/lambda"
	"go-lambda-api/handlers"
	"go-lambda-api/models"

	"github.com/aws/aws-lambda-go/events"
	aws_lambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/joho/godotenv"
)

// NewDB returns a new DynamoDB client
func NewDB() dynamodbiface.DynamoDBAPI {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not load .env file, assuming production environment: %v", err)
	}

	// Use the AWS_REGION environment variable, if available
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1" // Default to us-east-1 if not set
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})

	if err != nil {
		log.Fatalf("Error creating AWS session: %v", err)
	}

	return dynamodb.New(sess)
}

// Main is the entry point for the application.
// It determines whether to run as a local server or a Lambda function.
func Main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Could not load .env file, assuming production environment: %v", err)
	}

	if os.Getenv("LOCAL_SERVER") == "true" {
		startLocalServer()
	} else {
		startLambda()
	}
}

func startLocalServer() {
	log.Println("Starting local server...")

	dbClient := NewDB()
	userRepo := models.NewDynamoDBUserRepository(dbClient, os.Getenv("DYNAMODB_TABLE_NAME"))

	healthHandler := handlers.NewHealthHandler()

	r := http.NewServeMux()

	r.HandleFunc("GET /health", adapt(healthHandler.GetHealthHandler))
	r.HandleFunc("POST /users", adapt(handlers.NewUserHandler(userRepo).CreateUserHandler))
	r.HandleFunc("GET /users/{id}", adapt(handlers.NewUserHandler(userRepo).GetUserHandler))
	r.HandleFunc("GET /users", adapt(handlers.NewUserHandler(userRepo).GetAllUsersHandler))
	r.HandleFunc("PUT /users/{id}", adapt(handlers.NewUserHandler(userRepo).UpdateUserHandler))
	r.HandleFunc("DELETE /users/{id}", adapt(handlers.NewUserHandler(userRepo).DeleteUserHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		fmt.Printf("Local server listening on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", port, err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped.")
}

func startLambda() {
	log.Println("Starting Lambda function...")

	dbClient := NewDB()
	userRepo := models.NewDynamoDBUserRepository(dbClient, os.Getenv("DYNAMODB_TABLE_NAME"))
	healthHandler := handlers.NewHealthHandler()

	aws_lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return localLambda.Router(ctx, request, userRepo, healthHandler)
	})
}

type apiGatewayHandler func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// adapt converts a standard http.HandlerFunc to an apiGatewayHandler signature
// This allows reusing handler logic designed for Lambda with a local HTTP server.
func adapt(handler apiGatewayHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Convert http.Request to APIGatewayProxyRequest
		apiReq := events.APIGatewayProxyRequest{
			Path:                  r.URL.Path,
			HTTPMethod:            r.Method,
			Headers:               make(map[string]string),
			QueryStringParameters: make(map[string]string),
			PathParameters:        make(map[string]string),
		}

		for name, values := range r.Header {
			if len(values) > 0 {
				apiReq.Headers[name] = values[0]
			}
		}

		for name, values := range r.URL.Query() {
			if len(values) > 0 {
				apiReq.QueryStringParameters[name] = values[0]
			}
		}

		// Extract path parameters (simple example, might need more robust parsing)
		if r.PathValue("id") != "" {
			apiReq.PathParameters["id"] = r.PathValue("id")
		}

		// Execute the Lambda handler
		apiResp, err := handler(r.Context(), apiReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write response back to http.ResponseWriter
		for key, value := range apiResp.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(apiResp.StatusCode)
		_, err = w.Write([]byte(apiResp.Body))
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}
}
