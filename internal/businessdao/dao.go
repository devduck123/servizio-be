package businessdao

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrBusinessNotFound = errors.New("business not found")

type Business struct {
	ID       string   `json:"id" firestore:"-"`
	Name     string   `json:"name" firestore:"name"`
	Images   []string `json:"images,omitempty" firestore:"images,omitempty"`
	Category Category `json:"category" firestore:"category"`
	UserID   string   `json:"userId" firestore:"userId"`
}

type Dao struct {
	fsClient               *firestore.Client
	businessCollectionName string
}

func NewDao(client *firestore.Client, businessCollectionName string) *Dao {
	return &Dao{
		fsClient:               client,
		businessCollectionName: businessCollectionName,
	}
}

func (dao *Dao) GetBusiness(ctx context.Context, id string) (*Business, error) {
	docRef := dao.fsClient.Collection(dao.businessCollectionName).Doc(id)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrBusinessNotFound
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
	// category is optional
	Category Category
}

func (dao *Dao) GetAllBusinesses(ctx context.Context, input GetAllBusinessesInput) ([]Business, error) {
	query := dao.fsClient.Collection(dao.businessCollectionName).Query
	if input.Category != "" {
		query = query.Where("category", "==", input.Category)
	}
	snapshots, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	businesses := make([]Business, 0, len(snapshots))
	for _, snapshot := range snapshots {
		var business Business
		if err := snapshot.DataTo(&business); err != nil {
			return nil, err
		}
		business.ID = snapshot.Ref.ID
		businesses = append(businesses, business)
	}

	return businesses, nil
}

type CreateInput struct {
	Name     string
	Category Category
	UserID   string
}

func (dao *Dao) Create(ctx context.Context, input CreateInput) (*Business, error) {
	business := Business{
		Name:     input.Name,
		Category: input.Category,
		UserID:   input.UserID,
	}

	doc, _, err := dao.fsClient.Collection(dao.businessCollectionName).Add(ctx, business)
	if err != nil {
		return nil, err
	}
	business.ID = doc.ID
	return &business, nil
}

func (dao *Dao) Delete(ctx context.Context, id string) error {
	docRef := dao.fsClient.Collection(dao.businessCollectionName).Doc(id)
	_, err := docRef.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (dao *Dao) AppendImage(ctx context.Context, id string, imageURL string) error {
	// update business in firestore
	docRef := dao.fsClient.Collection(dao.businessCollectionName).Doc(id)
	_, err := docRef.Update(ctx, []firestore.Update{{
		Path:  "images",
		Value: firestore.ArrayUnion(imageURL),
	}})
	if err != nil {
		return err
	}

	return nil
}
