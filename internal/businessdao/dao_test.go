package businessdao

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
	businessCollection := fmt.Sprintf("business-%v", uuid.New())

	dao := NewDao(fsClient, businessCollection)

	t.Cleanup(func() {
		firestoretest.DeleteCollection(ctx, t, fsClient, businessCollection)
	})

	return dao
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
