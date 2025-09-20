package handlers

import (
	"context"
	"net/http"

	"go-lambda-api/utils"

	"github.com/aws/aws-lambda-go/events"
)

// HealthHandler struct for health check operations.
type HealthHandler struct{}

// NewHealthHandler creates and returns a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// GetHealthHandler returns a 200 OK response for health checks.
func (h *HealthHandler) GetHealthHandler(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return utils.APIResponse(http.StatusOK, map[string]string{"message": "Health Check OK"})
}
