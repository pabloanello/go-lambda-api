package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"

	"go-lambda-api/models"
	"go-lambda-api/utils"
)

// UserHandler struct holds the UserRepository interface.
type UserHandler struct {
	Repo models.UserRepository
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userRepo models.UserRepository) *UserHandler {
	return &UserHandler{Repo: userRepo}
}

func (h *UserHandler) CreateUserHandler(
	ctx context.Context, request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	var userReq models.UserRequest

	if err := json.Unmarshal([]byte(request.Body), &userReq); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, err)
	}

	if err := userReq.Validate(false); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, err)
	}

	newUser := models.User{
		ID:        uuid.New().String(),
		Name:      userReq.Name,
		Email:     userReq.Email,
		CreatedAt: time.Now(),
	}

	createdUser, err := h.Repo.CreateUser(newUser)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, err)
	}

	return utils.APIResponse(http.StatusCreated, createdUser)
}

func (h *UserHandler) GetUserHandler(
	ctx context.Context, request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	userID := request.PathParameters["id"]
	if userID == "" {
		return utils.ErrorResponse(http.StatusBadRequest, errors.New("user ID is required"))
	}

	user, err := h.Repo.GetUserByID(userID)
	if err != nil {
		return utils.ErrorResponse(http.StatusNotFound, err)
	}

	return utils.APIResponse(http.StatusOK, user)
}

func (h *UserHandler) GetAllUsersHandler(
	ctx context.Context, request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	userList := h.Repo.GetAllUsers()

	return utils.APIResponse(http.StatusOK, userList)
}

func (h *UserHandler) UpdateUserHandler(
	ctx context.Context, request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	userID := request.PathParameters["id"]
	if userID == "" {
		return utils.ErrorResponse(http.StatusBadRequest, errors.New("user ID is required"))
	}

	var userReq models.UserRequest
	if err := json.Unmarshal([]byte(request.Body), &userReq); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, err)
	}

	if err := userReq.Validate(true); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, err)
	}

	existingUser, err := h.Repo.GetUserByID(userID)
	if err != nil {
		return utils.ErrorResponse(http.StatusNotFound, err)
	}

	if userReq.Name != "" {
		existingUser.Name = userReq.Name
	}
	if userReq.Email != "" {
		existingUser.Email = userReq.Email
	}

	updatedUser, err := h.Repo.UpdateUser(existingUser)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, err)
	}

	return utils.APIResponse(http.StatusOK, updatedUser)
}

func (h *UserHandler) DeleteUserHandler(
	ctx context.Context, request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	userID := request.PathParameters["id"]
	if userID == "" {
		return utils.ErrorResponse(http.StatusBadRequest, errors.New("user ID is required"))
	}

	err := h.Repo.DeleteUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			return utils.ErrorResponse(http.StatusNotFound, err)
		}

		return utils.ErrorResponse(http.StatusInternalServerError, err)
	}

	return utils.APIResponse(http.StatusNoContent, nil)
}
