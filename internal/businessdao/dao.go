package businessdao

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

type Business struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Images []string `json:"images"`
}

type Dao struct {
	fsClient *firestore.Client
}

func NewDao(client *firestore.Client) *Dao {
	return &Dao{
		fsClient: client,
	}
}

func (dao *Dao) GetBusiness(ctx context.Context, id string) (*Business, error) {
	business := Business{
		ID: id,
	}

	return &business, nil
}

type CreateInput struct {
	Name string
}

func (dao *Dao) Create(ctx context.Context, input CreateInput) (*Business, error) {
	business := Business{
		ID:   uuid.New().String(),
		Name: input.Name,
	}

	return &business, nil

	// TODO: persist business to database
}
