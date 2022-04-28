package images

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/tj/assert"
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
		BucketName: "servizio-be.appspot.com",
	}

	fsClient, err := firestore.NewClient(ctx, projectID)
	assert.NoError(t, err)
	dao := businessdao.NewDao(fsClient, projectID)
	business, err := dao.Create(ctx, businessdao.CreateInput{
		Name: "foo",
	})
	assert.NoError(t, err)

	raw := []byte("hello")
	image, err := im.UploadImage(ctx, business.ID, raw)
	assert.NoError(t, err)

	gotBusiness, err := dao.GetBusiness(ctx, business.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(gotBusiness.Images))

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

	gotRaw, err := im.GetImage(ctx, "9520f5d1-6fe8-4d18-8546-9909bbbbe22d")
	assert.NoError(t, err)

	assert.Equal(t, []byte("hello"), gotRaw)
	fmt.Println(string(gotRaw))
}

func TestGetImages(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	assert.NoError(t, err)
	defer client.Close()

	im := ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com",
	}

	// we need to create a business,
	// then give it some images,
	// then call GetImages(ctx, business.ID)
	fsClient, err := firestore.NewClient(ctx, projectID)
	assert.NoError(t, err)
	dao := businessdao.NewDao(fsClient, projectID)
	input := businessdao.CreateInput{
		Name: "foobarbaz",
	}
	business, err := dao.Create(ctx, input)
	assert.NoError(t, err)

	gotRaws, err := im.GetImages(ctx, business.ID)
	assert.NoError(t, err)

	fmt.Println(gotRaws)
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
