package images

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
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
	objectPath := fmt.Sprintf("%s/%s", id, uuid.New().String())
	bucketName := i.BucketName

	// bucketIterator := i.API.Buckets(ctx, projectID)
	// for {
	// 	attrs, err := bucketIterator.Next()
	// 	if err == iterator.Done {
	// 		break
	// 	}
	// 	if err != nil {
	// 		return Image{}, err
	// 	}
	// 	fmt.Println("bucketName:", attrs.Name)
	// }

	fmt.Println("bucketName:", bucketName)
	fmt.Println("invoking Bucket")
	bucket := i.API.Bucket(bucketName)

	fmt.Println("invoking bucket.Object")
	object := bucket.Object(objectPath)
	fmt.Println("invoking object.NewWriter")
	w := object.NewWriter(ctx)
	fmt.Println("invoking w.Write")
	_, err := w.Write(raw)

	if err != nil {
		return Image{}, err
	}
	fmt.Println("invoking w.Close")
	if err := w.Close(); err != nil {
		return Image{}, fmt.Errorf("Writer.Close: %v", err)
	}

	return Image{
		Key: objectPath,
	}, nil
}

func (i ImageManager) GetImage(ctx context.Context, objectPath string) ([]byte, error) {
	bucketName := i.BucketName
	fmt.Println("objectPath:", objectPath)
	bucket := i.API.Bucket(bucketName)
	object := bucket.Object(objectPath)

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

// TODO: review this code to check if useful...
// func (i ImageManager) createBucket(ctx context.Context, bucketName string) (string, error) {
// 	bucket := i.API.Bucket(bucketName)
// 	err := bucket.Create(ctx, projectID, nil)
// 	if err != nil {
// 		return "", err
// 	}
// 	return fmt.Sprintf("created bucket: %s", bucketName), nil
// }

// func (i ImageManager) deleteBucket(ctx context.Context, bucketName string) (string, error) {
// 	bucketToDelete := i.API.Bucket(bucketName)
// 	err := bucketToDelete.Delete(ctx)
// 	if err != nil {
// 		return "", err
// 	}
// 	return fmt.Sprintf("successfully deleted bucket %s", bucketName), nil
// }

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
