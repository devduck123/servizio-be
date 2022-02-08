package businessdao

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Business struct {
	ID     string   `json:"id" firestore:"-"`
	Name   string   `json:"name" firestore:"name"`
	Images []string `json:"images,omitempty" firestore:"images,omitempty"`
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
	docRef := dao.fsClient.Collection("businesses").Doc(id)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("business not found")
		}
		return nil, err
	}
	var business Business
	if err := snapshot.DataTo(&business); err != nil {
		return nil, err
	}

	business.ID = docRef.ID
	return &business, nil
}

type GetAllBusinessesInput struct {
	// TODO:
}

func (dao *Dao) GetAllBusinesses(ctx context.Context, input GetAllBusinessesInput) ([]Business, error) {
	// TODO:
	return nil, nil
}

type CreateInput struct {
	Name string
}

func (dao *Dao) Create(ctx context.Context, input CreateInput) (*Business, error) {
	business := Business{
		Name: input.Name,
	}

	doc, _, err := dao.fsClient.Collection("businesses").Add(ctx, business)
	if err != nil {
		return nil, err
	}
	business.ID = doc.ID
	return &business, nil
}
