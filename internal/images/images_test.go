package images

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/storage"
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

func TestImageManager(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	assert.NoError(t, err)
	defer client.Close()

	im := ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com",
	}

	image, err := im.UploadImage(ctx, []byte("hello"))
	assert.NoError(t, err)

	fmt.Println(image.Key)
	fmt.Println(image.SignedURL)

}
