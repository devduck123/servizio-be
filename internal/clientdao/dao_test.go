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

func TestGetAllClients(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	createInput := CreateInput{
		FirstName: "foo",
		LastName:  "dog",
	}
	createInput2 := CreateInput{
		FirstName: "bar",
		LastName:  "cat",
	}
	_, err := dao.Create(ctx, createInput)
	assert.NoError(t, err)
	_, err = dao.Create(ctx, createInput2)
	assert.NoError(t, err)

	allClients, err := dao.GetAllClients(ctx)
	assert.NoError(t, err)
	assert.Len(t, allClients, 2)
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

func TestDeleteClient(t *testing.T) {
	// first create the client
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

	// then delete the client
	err = dao.Delete(ctx, client.ID)
	assert.NoError(t, err)
}

func TestAppendImage(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	input := CreateInput{
		FirstName: "foo",
		LastName:  "bar",
	}
	client, err := dao.Create(ctx, input)
	assert.NoError(t, err)

	dao.AppendImage(ctx, client.ID, "test1")
	dao.AppendImage(ctx, client.ID, "test2")
	dao.AppendImage(ctx, client.ID, "test3")

	client, err = dao.GetClient(ctx, client.ID)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(client.Images))
}
