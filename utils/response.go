package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

var jsonMarshal = json.Marshal

// APIResponse generates a consistent APIGatewayProxyResponse.
func APIResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	var respBody []byte
	var err error

	if body != nil {
		respBody, err = jsonMarshal(body)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode:        http.StatusInternalServerError,
				Headers:           headers,
				MultiValueHeaders: map[string][]string{},
				IsBase64Encoded:   false,
				Body:              fmt.Sprintf(`{"error": "%s"}`, err.Error()),
			}, fmt.Errorf("failed to marshal response body: %w", err)
		}
	}

	response := events.APIGatewayProxyResponse{
		StatusCode:        statusCode,
		Headers:           headers,
		Body:              string(respBody),
		MultiValueHeaders: map[string][]string{},
		IsBase64Encoded:   false,
	}

	EnsureHeaders(&response)

	return response, nil
}

// ErrorResponse generates a consistent error APIGatewayProxyResponse.
func ErrorResponse(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	errMessage := ""
	if err != nil {
		log.Printf("Error: %v", err.Error())
		errMessage = err.Error()
	}

	respBody, jsonErr := jsonMarshal(map[string]string{"error": errMessage})
	if jsonErr != nil {
		// Fallback if marshaling error also fails
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: "{\"error\": \"failed to marshal error response\"}"}, nil
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(respBody),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}

	EnsureHeaders(&response)

	return response, nil
}

// EnsureHeaders ensures that essential headers are present in the response.
func EnsureHeaders(response *events.APIGatewayProxyResponse) {
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}

	if _, ok := response.Headers["Content-Type"]; !ok {
		response.Headers["Content-Type"] = "application/json"
	}
}
