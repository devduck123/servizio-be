package images

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/google/uuid"
)

var projectID = "servizio-be"

type Image struct {
	SignedURL string
	Key       string
}

type ImageManager struct {
	API        *storage.Client
	BucketName string
}

func (i ImageManager) UploadImage(ctx context.Context, id string, raw []byte) (Image, error) {
	key := uuid.New().String()

	bucket := i.API.Bucket(i.BucketName)
	object := bucket.Object(key)
	w := object.NewWriter(ctx)
	_, err := w.Write(raw)
	if err != nil {
		return Image{}, err
	}
	if err := w.Close(); err != nil {
		return Image{}, fmt.Errorf("Writer.Close: %v", err)
	}

	// AppendImage to business in firestore
	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return Image{}, err
	}
	dao := businessdao.NewDao(fsClient, projectID)
	if err = dao.AppendImage(ctx, id, key); err != nil {
		return Image{}, err
	}

	return Image{
		Key: key,
	}, nil
}

func (i ImageManager) GetImage(ctx context.Context, key string) ([]byte, error) {
	bucket := i.API.Bucket(i.BucketName)
	object := bucket.Object(key)

	reader, err := object.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	err = reader.Close()
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func (i ImageManager) GetImages(ctx context.Context, id string) ([][]byte, error) {
	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dao := businessdao.NewDao(fsClient, projectID)
	business, err := dao.GetBusiness(ctx, id)
	if err != nil {
		return nil, err
	}

	fmt.Println("GetImages called for:", business.Images)

	var raws [][]byte
	for _, key := range business.Images {
		raw, err := i.GetImage(ctx, key)
		if err != nil {
			return nil, err
		}
		raws = append(raws, raw)
	}

	return raws, nil
}

func (i ImageManager) CreateBucket(ctx context.Context, bucketName string) (string, error) {
	bucket := i.API.Bucket(bucketName)
	err := bucket.Create(ctx, projectID, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("created bucket: %s", bucketName), nil
}

func (i ImageManager) DeleteBucket(ctx context.Context, bucketName string) (string, error) {
	bucketToDelete := i.API.Bucket(bucketName)
	err := bucketToDelete.Delete(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("successfully deleted bucket %s", bucketName), nil
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
