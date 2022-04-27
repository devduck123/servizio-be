package images

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/tj/assert"
	"google.golang.org/api/option"
)

// var projectID = "servizio-be"

func TestMain(m *testing.M) {
	if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080"); err != nil {
		log.Fatal("failed to set FIRESTORE_EMULATOR_HOST environment variable", err)
	}
	if err := os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099"); err != nil {
		log.Fatal("failed to set FIREBASE_AUTH_EMULATOR_HOST environment variable", err)
	}

	if err := os.Setenv("STORAGE_EMULATOR_HOST", "localhost:9199"); err != nil {
		log.Fatal("failed to set STORAGE_EMULATOR_HOST environment variable", err)
	}

	m.Run()
}

func TestUploadImage(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	assert.NoError(t, err)
	defer client.Close()

	im := ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com/foo",
	}

	raw := []byte("hello")
	image, err := im.UploadImage(ctx, raw)
	assert.NoError(t, err)

	fmt.Println("image key:", image.Key)

	gotRaw, err := im.GetImage(ctx, image.Key)
	assert.NoError(t, err)
	assert.Equal(t, raw, gotRaw)
}

func TestGetImage(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	assert.NoError(t, err)
	defer client.Close()

	im := ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com",
	}

	imageName, err := im.GetImage(ctx, "351e3547-eee4-405e-a3e9-41b317dd3915")
	assert.NoError(t, err)

	fmt.Println(imageName)
}

func TestGetImages(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx,
		option.WithEndpoint("http://localhost:9199/storage/v1/"))
	assert.NoError(t, err)
	defer client.Close()

	im := ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com",
	}

	imageNames, err := im.GetImages(ctx)
	assert.NoError(t, err)

	fmt.Println(imageNames)
}

func TestCreateBucket(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	assert.NoError(t, err)
	defer client.Close()

	im := ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com",
	}

	result, err := im.CreateBucket(ctx, "servizio-be.appspot.com/world")
	assert.NoError(t, err)

	fmt.Println(result)
}

func TestDeleteBucket(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	assert.NoError(t, err)
	defer client.Close()

	im := ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com",
	}

	result, err := im.DeleteBucket(ctx, "servizio-be.appspot.com/hello")
	assert.NoError(t, err)

	fmt.Println(result)
}
