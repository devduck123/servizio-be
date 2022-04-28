package images

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
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

func (i ImageManager) UploadImage(ctx context.Context, raw []byte) (Image, error) {
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

	return Image{
		Key: key,
	}, nil
}

// TODO: convert this to return Image instead of the image name
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

// TODO: convert this to return []Image instead of the slice of image names
func (i ImageManager) GetImages(ctx context.Context) ([]string, error) {
	bucket := i.API.Bucket(i.BucketName)

	fmt.Println("bucket name:", i.BucketName)
	fmt.Println("bucket:", bucket)

	var names []string
	it := bucket.Objects(ctx, nil)
	pageInfo := it.PageInfo()
	fmt.Println("remaining:", pageInfo.Remaining())
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		names = append(names, attrs.Name)
	}

	fmt.Println(names)

	return names, nil
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
