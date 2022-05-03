package authtest

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/tj/assert"
)

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

// use this for getting a JWT locally
func TestGetJWT(t *testing.T) {
	t.SkipNow()
	
	token, err := signIn(t, "test@email.com", "password")
	assert.NoError(t, err)

	fmt.Println("token:", token)
}

// use this test for creating a JWT locally
func TestCreateJWT(t *testing.T) {
	t.SkipNow()

	token, err := GetJWT(t, "servizio-be")
	assert.NoError(t, err)

	fmt.Println("token:", token)
}
