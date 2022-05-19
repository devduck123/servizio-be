package images

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

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

	im := ImageManager{
		BucketName: "servizio-be.appspot.com",
	}

	raw := []byte("hello")
	image, err := im.UploadImage(ctx, "foo", raw)
	assert.NoError(t, err)

	fmt.Println("image key:", image.Key)

	gotRaw, err := im.GetImage(ctx, image.Key)
	assert.NoError(t, err)
	assert.Equal(t, raw, gotRaw)
}

func TestGetImage_Exists(t *testing.T) {
	t.SkipNow()

	ctx := context.Background()

	im := ImageManager{
		BucketName: "servizio-be.appspot.com",
	}

	gotRaw, err := im.GetImage(ctx, "39afdab7-1ef8-4e4e-be9b-b8254cb45c69")
	assert.NoError(t, err)

	assert.Equal(t, []byte("hello"), gotRaw)
	fmt.Println(string(gotRaw))
}
