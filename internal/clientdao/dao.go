package clientdao

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrClientNotFound = errors.New("client not found")

type Client struct {
	ID        string   `json:"id" firestore:"-"`
	FirstName string   `json:"firstName" firestore:"firstName"`
	LastName  string   `json:"lastName" firestore:"lastName"`
	Images    []string `json:"images,omitempty" firestore:"images,omitempty"`
}

type Dao struct {
	fsClient             *firestore.Client
	clientCollectionName string
}

func NewDao(client *firestore.Client, clientCollectionName string) *Dao {
	return &Dao{
		fsClient:             client,
		clientCollectionName: clientCollectionName,
	}
}

func (dao *Dao) GetClient(ctx context.Context, id string) (*Client, error) {
	docRef := dao.fsClient.Collection(dao.clientCollectionName).Doc(id)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	var client Client
	if err := snapshot.DataTo(&client); err != nil {
		return nil, err
	}

	client.ID = docRef.ID
	return &client, nil
}

func (dao *Dao) GetAllClients(ctx context.Context) ([]Client, error) {
	query := dao.fsClient.Collection(dao.clientCollectionName).Query
	snapshots, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	clients := make([]Client, 0, len(snapshots))
	for _, snapshot := range snapshots {
		var client Client
		if err := snapshot.DataTo(&client); err != nil {
			return nil, err
		}
		client.ID = snapshot.Ref.ID
		clients = append(clients, client)
	}

	return clients, nil
}

type CreateInput struct {
	FirstName string
	LastName  string
	UserID    string
}

func (dao *Dao) Create(ctx context.Context, input CreateInput) (*Client, error) {
	client := Client{
		FirstName: input.FirstName,
		LastName:  input.LastName,
	}

	doc, _, err := dao.fsClient.Collection(dao.clientCollectionName).Add(ctx, client)
	if err != nil {
		return nil, err
	}
	client.ID = doc.ID
	return &client, nil
}

func (dao *Dao) Delete(ctx context.Context, id string) error {
	docRef := dao.fsClient.Collection(dao.clientCollectionName).Doc(id)
	_, err := docRef.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (dao *Dao) AppendImage(ctx context.Context, id string, key string) error {
	// update client in firestore
	docRef := dao.fsClient.Collection(dao.clientCollectionName).Doc(id)
	_, err := docRef.Update(ctx, []firestore.Update{{
		Path:  "images",
		Value: firestore.ArrayUnion(key),
	}})
	if err != nil {
		return err
	}

	return nil
}
