package clientdao

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/devduck123/servizio-be/internal/firestoretest"
	"github.com/google/uuid"
	"github.com/tj/assert"
)

func TestMain(m *testing.M) {
	if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080"); err != nil {
		log.Fatal("failed to set FIRESTORE_EMULATOR_HOST environment variable", err)
	}

	m.Run()
}

func createTestDao(ctx context.Context, t *testing.T) *Dao {
	fsClient := firestoretest.CreateTestClient(ctx, t)
	clientCollection := fmt.Sprintf("client-%v", uuid.New())

	dao := NewDao(fsClient, clientCollection)

	t.Cleanup(func() {
		firestoretest.DeleteCollection(ctx, t, fsClient, clientCollection)
	})

	return dao
}

func TestGetClientByID(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	input := CreateInput{
		FirstName: "foo",
		LastName:  "bar",
	}
	client, err := dao.Create(ctx, input)
	assert.NoError(t, err)

	gotClient, err := dao.GetClient(ctx, client.ID)
	assert.NoError(t, err)
	assert.NotNil(t, gotClient)
	assert.Equal(t, client.ID, gotClient.ID)
	assert.Equal(t, input.FirstName, gotClient.FirstName)
	assert.Equal(t, input.LastName, gotClient.LastName)
}

func TestGetClient_NotExists(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	gotClient, err := dao.GetClient(ctx, "notexists")
	assert.Equal(t, "client not found", err.Error())
	assert.Nil(t, gotClient)
}

func TestCreateClient(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	input := CreateInput{
		FirstName: "foo",
		LastName:  "bar",
	}
	client, err := dao.Create(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotEmpty(t, client.ID)
	assert.Equal(t, input.FirstName, client.FirstName)
	assert.Equal(t, input.LastName, client.LastName)
}
