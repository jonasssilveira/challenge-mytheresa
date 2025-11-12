package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOKResponse(t *testing.T) {
	type sampleResponse struct {
		Message string `json:"message"`
	}

	tests := []struct {
		name                string
		data                interface{}
		expectedStatus      int
		expectedBody        string
		expectedContentType string
	}{
		{
			name:                "successful http200 json response",
			data:                sampleResponse{Message: "Success"},
			expectedStatus:      http.StatusOK,
			expectedBody:        `{"message":"Success"}`,
			expectedContentType: "application/json",
		},
		{
			name:                "response with empty struct",
			data:                struct{}{},
			expectedStatus:      http.StatusOK,
			expectedBody:        `{}`,
			expectedContentType: "application/json",
		},
		{
			name:                "response with slice",
			data:                []string{"item1", "item2"},
			expectedStatus:      http.StatusOK,
			expectedBody:        `["item1","item2"]`,
			expectedContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			OKResponse(recorder, tt.data)

			assert.Equal(t, tt.expectedStatus, recorder.Code, "Expected status code to match")
			assert.Equal(t, tt.expectedContentType, recorder.Header().Get("Content-Type"), "Expected Content-Type to match")
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String(), "Response body does not match expected")
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name                string
		statusCode          int
		message             string
		expectedBody        string
		expectedContentType string
	}{
		{
			name:                "internal server error",
			statusCode:          http.StatusInternalServerError,
			message:             "Some error occurred",
			expectedBody:        `{"error":"Some error occurred"}`,
			expectedContentType: "application/json",
		},
		{
			name:                "bad request error",
			statusCode:          http.StatusBadRequest,
			message:             "Invalid input",
			expectedBody:        `{"error":"Invalid input"}`,
			expectedContentType: "application/json",
		},
		{
			name:                "not found error",
			statusCode:          http.StatusNotFound,
			message:             "Resource not found",
			expectedBody:        `{"error":"Resource not found"}`,
			expectedContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ErrorResponse(recorder, tt.statusCode, tt.message)

			assert.Equal(t, tt.statusCode, recorder.Code, "Expected status code to match")
			assert.Equal(t, tt.expectedContentType, recorder.Header().Get("Content-Type"), "Expected Content-Type to match")
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String(), "Response body does not match expected")
		})
	}
}
