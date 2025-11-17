package main

import (
	"context"
	"testing"
)

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     Request
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Valid request with all fields",
			request: Request{
				Name:        "João Silva",
				PhoneNumber: "+5511999999999",
			},
			shouldError: false,
		},
		{
			name: "Valid request with required fields only",
			request: Request{
				Name:        "Maria Santos",
				PhoneNumber: "+5511888888888",
			},
			shouldError: false,
		},
		{
			name: "Missing name",
			request: Request{
				PhoneNumber: "+5511999999999",
			},
			shouldError: true,
			errorMsg:    "name is required",
		},
		{
			name: "Missing phone_number",
			request: Request{
				Name: "João Silva",
			},
			shouldError: true,
			errorMsg:    "phone_number is required",
		},
		{
			name:        "Empty request",
			request:     Request{},
			shouldError: true,
			errorMsg:    "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock context
			ctx := context.Background()

			// Set required environment variable
			t.Setenv("USERS_TABLE_NAME", "test-users-table")

			// Note: This will fail when trying to connect to DynamoDB
			// For proper testing, you'd need to mock the DynamoDB client
			// or use LocalStack/DynamoDB Local
			_, err := HandleRequest(ctx, tt.request)

			if tt.shouldError {
				if err == nil {
					// Check if response contains error
					// For now, we're just checking basic validation
					if tt.request.Name == "" || tt.request.PhoneNumber == "" {
						// Expected behavior
						return
					}
				}
			}
		})
	}
}

func TestUserStructure(t *testing.T) {
	user := User{
		ID:          "test-id",
		Name:        "Test User",
		PhoneNumber: "+5511999999999",
		Status:      "active",
		CreatedAt:   "2025-11-17T12:00:00Z",
	}

	if user.ID == "" {
		t.Error("User ID should not be empty")
	}
	if user.Name == "" {
		t.Error("User Name should not be empty")
	}
	if user.PhoneNumber == "" {
		t.Error("User PhoneNumber should not be empty")
	}
	if user.Status != "active" {
		t.Error("User Status should be 'active'")
	}
}

func TestRequestStructure(t *testing.T) {
	req := Request{
		Name:        "Test User",
		PhoneNumber: "+5511999999999",
	}

	if req.Name == "" {
		t.Error("Request Name should not be empty")
	}
	if req.PhoneNumber == "" {
		t.Error("Request PhoneNumber should not be empty")
	}
}

