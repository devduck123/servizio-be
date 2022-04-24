package images

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type Image struct {
	SignedURL string
	Key       string
}

type ImageManager struct {
	API        *storage.Client
	BucketName string
}

func (i ImageManager) UploadImage(ctx context.Context, raw []byte) (Image, error) {
	key := uuid.New().String()

	object := i.API.Bucket(i.BucketName).Object(key)
	writer := object.NewWriter(ctx)
	_, err := writer.Write(raw)
	if err != nil {
		return Image{}, err
	}
	if err := writer.Close(); err != nil {
		return Image{}, fmt.Errorf("Writer.Close: %v", err)
	}

	return Image{
		Key: key,
	}, nil
}

func (i ImageManager) GetSignedURL(ctx context.Context) (Image, error) {
	key := uuid.New().String()

	url, err := i.API.Bucket(i.BucketName).SignedURL(key, &storage.SignedURLOptions{
		Method:         "PUT",
		Expires:        time.Now().Add(15 * time.Minute),
		Scheme:         storage.SigningSchemeV4,
		GoogleAccessID: "xxx@developer.gserviceaccount.com",
	})
	if err != nil {
		return Image{}, err
	}

	return Image{
		SignedURL: url,
		Key:       key,
	}, nil
}
