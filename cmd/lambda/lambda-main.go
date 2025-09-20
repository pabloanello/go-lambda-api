package lambda

import (
	"context"
	"errors"
	"log"
	"net/http"

	"go-lambda-api/handlers"
	"go-lambda-api/models"
	"go-lambda-api/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	UsersIDPath = "/users/{id}"
	RootPath    = "/"
	HealthPath  = "/health"
	UsersPath   = "/users"
)

// Router handles routing of API Gateway requests to appropriate handlers.
func Router(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
	userRepo models.UserRepository,
	healthHandler *handlers.HealthHandler,
) (events.APIGatewayProxyResponse, error) {
	var response events.APIGatewayProxyResponse
	var err error

	// Initialize common headers, including CORS
	commonHeaders := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*", // Allow all origins for simplicity
		"Access-Control-Allow-Methods":     "GET,POST,PUT,DELETE,OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type,Authorization,X-Amz-Date,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Credentials": "true",
	}

	// Handle OPTIONS pre-flight requests
	if request.HTTPMethod == http.MethodOptions {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    commonHeaders,
		}, nil
	}

	// Set initial response headers (will be merged later if APIResponse is used)
	response.Headers = commonHeaders

	switch {
	case request.Path == RootPath && request.HTTPMethod == http.MethodGet:
		response, err = handleRootGet(request)
	case request.Path == HealthPath && request.HTTPMethod == http.MethodGet:
		response, err = healthHandler.GetHealthHandler(ctx, request)
	case request.Path == UsersPath && request.HTTPMethod == http.MethodPost:
		response, err = handleCreateUser(ctx, request, userRepo)
	case request.Path == UsersIDPath && request.HTTPMethod == http.MethodGet:
		response, err = handleGetUser(ctx, request, userRepo)
	case request.Path == UsersPath && request.HTTPMethod == http.MethodGet:
		response, err = handleGetAllUsers(ctx, request, userRepo)
	case request.Path == UsersIDPath && request.HTTPMethod == http.MethodPut:
		response, err = handleUpdateUser(ctx, request, userRepo)
	case request.Path == UsersIDPath && request.HTTPMethod == http.MethodDelete:
		response, err = handleDeleteUser(ctx, request, userRepo)
	default:
		response, err = utils.ErrorResponse(http.StatusNotFound, errors.New("not found"))
	}

	if err != nil {
		log.Printf("Error processing request: %v", err)

		// If an error occurs, preserve the original status code if it's an error from utils.ErrorResponse,
		// otherwise, default to InternalServerError.
		// We also ensure CORS headers are present in the error response.
		statusCode := http.StatusInternalServerError
		if response.StatusCode != 0 && response.StatusCode != http.StatusOK {
			statusCode = response.StatusCode
		}

		errorResponse, _ := utils.ErrorResponse(statusCode, err)
		for k, v := range commonHeaders {
			errorResponse.Headers[k] = v
		}
		return errorResponse, nil
	}

	// Merge common headers with the handler's response headers
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	for k, v := range commonHeaders {
		// Only set if not already set by the handler to allow overrides
		if _, ok := response.Headers[k]; !ok {
			response.Headers[k] = v
		}
	}

	return response, nil
}

func main() {
	log.Println("Lambda cold start")

	healthHandler := handlers.NewHealthHandler()
	// For now, using a mock user repository. Replace with actual implementation.
	userRepo := models.NewInMemoryUserRepository()

	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return Router(ctx, request, userRepo, healthHandler)
	})
}

func handleRootGet(_ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return utils.APIResponse(http.StatusOK, map[string]string{"message": "Welcome to the Go Lambda API"})
}

func handleCreateUser(
	ctx context.Context, request events.APIGatewayProxyRequest, userRepo models.UserRepository,
) (events.APIGatewayProxyResponse, error) {
	userHandler := handlers.NewUserHandler(userRepo)

	return userHandler.CreateUserHandler(ctx, request)
}

func handleGetUser(
	ctx context.Context, request events.APIGatewayProxyRequest, userRepo models.UserRepository,
) (events.APIGatewayProxyResponse, error) {
	userHandler := handlers.NewUserHandler(userRepo)

	return userHandler.GetUserHandler(ctx, request)
}

func handleGetAllUsers(
	ctx context.Context, request events.APIGatewayProxyRequest, userRepo models.UserRepository,
) (events.APIGatewayProxyResponse, error) {
	userHandler := handlers.NewUserHandler(userRepo)

	return userHandler.GetAllUsersHandler(ctx, request)
}

func handleUpdateUser(
	ctx context.Context, request events.APIGatewayProxyRequest, userRepo models.UserRepository,
) (events.APIGatewayProxyResponse, error) {
	userHandler := handlers.NewUserHandler(userRepo)

	return userHandler.UpdateUserHandler(ctx, request)
}

func handleDeleteUser(
	ctx context.Context, request events.APIGatewayProxyRequest, userRepo models.UserRepository,
) (events.APIGatewayProxyResponse, error) {
	userHandler := handlers.NewUserHandler(userRepo)

	return userHandler.DeleteUserHandler(ctx, request)
}
