package pocketbase

import (
	"context"
	"fmt"
	"net/http"
)

// SettingServiceAPI defines operations for viewing and modifying settings.
type SettingServiceAPI interface {
	GetAll(ctx context.Context) (map[string]interface{}, error)
	Update(ctx context.Context, body interface{}) (map[string]interface{}, error)
	TestS3(ctx context.Context) (map[string]interface{}, error)
	TestEmail(ctx context.Context, toEmail string) (map[string]interface{}, error)
	GenerateAppleClientSecret(ctx context.Context, params interface{}) (map[string]interface{}, error)
}

// SettingService provides functionality for viewing and modifying server settings.
type SettingService struct {
	Client *Client
}

var _ SettingServiceAPI = (*SettingService)(nil)

// GetAll retrieves all setting values.
func (s *SettingService) GetAll(ctx context.Context) (map[string]interface{}, error) {
	path := "/api/settings"
	var result map[string]interface{}
	err := s.Client.send(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: get settings: %w", err)
	}
	return result, nil
}

// Update modifies settings.
func (s *SettingService) Update(ctx context.Context, body interface{}) (map[string]interface{}, error) {
	path := "/api/settings"
	var result map[string]interface{}
	err := s.Client.send(ctx, http.MethodPatch, path, body, &result)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: update settings: %w", err)
	}
	return result, nil
}

// TestS3 tests the S3 file storage settings.
func (s *SettingService) TestS3(ctx context.Context) (map[string]interface{}, error) {
	path := "/api/settings/test/s3"
	var result map[string]interface{}
	err := s.Client.send(ctx, http.MethodPost, path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: test s3: %w", err)
	}
	return result, nil
}

// TestEmail tests the SMTP settings.
func (s *SettingService) TestEmail(ctx context.Context, toEmail string) (map[string]interface{}, error) {
	path := "/api/settings/test/email"
	body := map[string]string{"email": toEmail}
	var result map[string]interface{}
	err := s.Client.send(ctx, http.MethodPost, path, body, &result)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: test email: %w", err)
	}
	return result, nil
}

// GenerateAppleClientSecret generates a client secret for Apple OAuth.
func (s *SettingService) GenerateAppleClientSecret(ctx context.Context, params interface{}) (map[string]interface{}, error) {
	path := "/api/settings/apple/generate-client-secret"
	var result map[string]interface{}
	err := s.Client.send(ctx, http.MethodPost, path, params, &result)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: generate apple client secret: %w", err)
	}
	return result, nil
}
