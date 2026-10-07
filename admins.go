package pocketbase

import "context"

// AdminServiceAPI defines the API operations for admin accounts.
// Deprecated: use Records with the _superusers collection.
type AdminServiceAPI interface {
	GetList(ctx context.Context, opts *ListOptions) (*ListResult, error)
	GetOne(ctx context.Context, adminID string) (*Admin, error)
	Create(ctx context.Context, body any) (*Admin, error)
	Update(ctx context.Context, adminID string, body any) (*Admin, error)
	Delete(ctx context.Context, adminID string) error
}

// AdminService provides API for managing superuser accounts.
// Deprecated: use Records with the _superusers collection.
type AdminService struct{ Client *Client }

var _ AdminServiceAPI = (*AdminService)(nil)

func (s *AdminService) GetList(ctx context.Context, opts *ListOptions) (*ListResult, error) {
	return s.Client.Records.GetList(ctx, "_superusers", opts)
}

func (s *AdminService) GetOne(ctx context.Context, adminID string) (*Admin, error) {
	record, err := s.Client.Records.GetOne(ctx, "_superusers", adminID, nil)
	return adminFromRecord(record, err)
}

func (s *AdminService) Create(ctx context.Context, body any) (*Admin, error) {
	record, err := s.Client.Records.Create(ctx, "_superusers", body)
	return adminFromRecord(record, err)
}

func (s *AdminService) Update(ctx context.Context, adminID string, body any) (*Admin, error) {
	record, err := s.Client.Records.Update(ctx, "_superusers", adminID, body)
	return adminFromRecord(record, err)
}

func (s *AdminService) Delete(ctx context.Context, adminID string) error {
	return s.Client.Records.Delete(ctx, "_superusers", adminID)
}

func adminFromRecord(record *Record, err error) (*Admin, error) {
	if err != nil {
		return nil, err
	}
	return &Admin{ID: record.ID, CollectionID: record.CollectionID,
		CollectionName: record.CollectionName, Email: record.GetString("email"),
		Avatar: int(record.GetFloat("avatar"))}, nil
}
