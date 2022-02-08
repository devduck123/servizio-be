package businessdao

import (
	"context"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/tj/assert"
)

func TestMain(m *testing.M) {
	if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080"); err != nil {
		log.Fatal("failed to set FIRESTORE_EMULATOR_HOST environment variable", err)
	}

	m.Run()
}

func TestGetBusinessByID(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "test")
	assert.NoError(t, err)
	dao := NewDao(client)
	input := CreateInput{
		Name: "foobarbaz",
	}
	business, err := dao.Create(ctx, input)
	assert.NoError(t, err)

	gotBusiness, err := dao.GetBusiness(ctx, business.ID)
	assert.NoError(t, err)
	assert.NotNil(t, gotBusiness)
	assert.Equal(t, business.ID, gotBusiness.ID)
	assert.Equal(t, input.Name, gotBusiness.Name)
}

func TestGetBusiness_NotExists(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "test")
	assert.NoError(t, err)
	dao := NewDao(client)

	gotBusiness, err := dao.GetBusiness(ctx, "notexists")
	assert.Equal(t, "business not found", err.Error())
	assert.Nil(t, gotBusiness)
}

func TestCreateBusiness(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "servizio-be")
	assert.NoError(t, err)
	dao := NewDao(client)
	input := CreateInput{
		Name: "foobarbaz",
	}
	business, err := dao.Create(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, business)
	assert.NotEmpty(t, business.ID)
	assert.Equal(t, input.Name, business.Name)
}
