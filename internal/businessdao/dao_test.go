package businessdao

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/tj/assert"
)

func TestMain(m *testing.M) {
	if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080"); err != nil {
		log.Fatal("failed to set FIRESTORE_EMULATOR_HOST environment variable", err)
	}

	m.Run()
}

func createTestDao(ctx context.Context, t *testing.T, keepCollection ...bool) *Dao {
	t.Helper()

	client, err := firestore.NewClient(ctx, "servizio-be")
	assert.NoError(t, err)
	businessCollection := fmt.Sprintf("business-%v", uuid.New())
	dao := NewDao(client, businessCollection)

	cleanUp := true
	for _, kc := range keepCollection {
		if kc {
			cleanUp = false
			break
		}
	}

	if cleanUp {
		t.Cleanup(func() {
			deleteCollection(ctx, t, client, businessCollection)
		})
	}

	return dao
}

func deleteCollection(ctx context.Context, t *testing.T, fsClient *firestore.Client, collection string) {
	documentRefs, err := fsClient.Collection(collection).DocumentRefs(ctx).GetAll()
	assert.NoError(t, err)

	for _, document := range documentRefs {
		_, err := document.Delete(ctx)
		assert.NoError(t, err)
	}
}

func TestGetBusinessByID(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

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
	dao := createTestDao(ctx, t)

	gotBusiness, err := dao.GetBusiness(ctx, "notexists")
	assert.Equal(t, "business not found", err.Error())
	assert.Nil(t, gotBusiness)
}

func TestCreateBusiness(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	input := CreateInput{
		Name: "foobarbaz",
	}
	business, err := dao.Create(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, business)
	assert.NotEmpty(t, business.ID)
	assert.Equal(t, input.Name, business.Name)
}

func TestGetAllBusinesses(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	createInput := CreateInput{
		Name:     "foo",
		Category: CategoryAutomotive,
	}
	_, err := dao.Create(ctx, createInput)
	assert.NoError(t, err)

	getAllBusinessesInput := GetAllBusinessesInput{
		Category: CategoryAutomotive,
	}
	allBusinesses, err := dao.GetAllBusinesses(ctx, getAllBusinessesInput)
	assert.NoError(t, err)
	assert.Len(t, allBusinesses, 1)
}
