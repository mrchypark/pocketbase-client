package pocketbase

import (
	"context"
	"net/http"
)

// SettingServiceAPI defines operations for viewing and modifying settings.
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
	var result map[string]interface{}
	if err := s.Client.send(ctx, http.MethodGet, "/api/settings", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Update modifies settings.
func (s *SettingService) Update(ctx context.Context, body interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := s.Client.send(ctx, http.MethodPatch, "/api/settings", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TestS3 tests the S3 file storage settings.
func (s *SettingService) TestS3(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := s.Client.send(ctx, http.MethodPost, "/api/settings/test/s3", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TestEmail tests the SMTP settings.
func (s *SettingService) TestEmail(ctx context.Context, toEmail string) (map[string]interface{}, error) {
	body := map[string]string{"email": toEmail}
	var result map[string]interface{}
	if err := s.Client.send(ctx, http.MethodPost, "/api/settings/test/email", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateAppleClientSecret generates a client secret for Apple OAuth.
func (s *SettingService) GenerateAppleClientSecret(ctx context.Context, params interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := s.Client.send(ctx, http.MethodPost, "/api/settings/apple/generate-client-secret", params, &result); err != nil {
		return nil, err
	}
	return result, nil
}
