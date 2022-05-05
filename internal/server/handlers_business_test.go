package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/google/uuid"
	"github.com/tj/assert"
)

func createTestBusinessDao(ctx context.Context, t *testing.T, keepCollection ...bool) *businessdao.Dao {
	t.Helper()

	client, err := firestore.NewClient(ctx, projectID)
	assert.NoError(t, err)
	businessCollection := fmt.Sprintf("business-%v", uuid.New())
	dao := businessdao.NewDao(client, businessCollection)

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

func TestCreateBusiness_Invalid(t *testing.T) {
	ctx := context.Background()
	dao := createTestBusinessDao(ctx, t)
	server := NewServer(dao, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	body := bytes.NewReader([]byte(`{}`))
	r := httptest.NewRequest(http.MethodPost, "/", body)
	r = r.WithContext(ContextWithUser(ctx, User{}))
	server.CreateBusiness(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	raw, err := ioutil.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"error":"name cannot be empty"}`, string(raw))
}

func TestCreateBusiness_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestBusinessDao(ctx, t)
	server := NewServer(dao, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	body := bytes.NewReader([]byte(`{
		"name": "test",
		"category": "pets"
	}`))
	r := httptest.NewRequest(http.MethodPost, "/", body)
	r = r.WithContext(ContextWithUser(ctx, User{}))
	server.CreateBusiness(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response businessdao.Business
	err := json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "test", response.Name)
	assert.EqualValues(t, "pets", response.Category)
	assert.NotEmpty(t, response.ID)
}

func TestGetBusiness_Invalid(t *testing.T) {
	ctx := context.Background()
	dao := createTestBusinessDao(ctx, t)
	server := NewServer(dao, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/businesses/foo", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetBusiness(w, r)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

	var response businessdao.Business
	err := json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)
}

func TestGetBusiness_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestBusinessDao(ctx, t)
	server := NewServer(dao, nil, nil, nil, nil)

	business, err := dao.Create(ctx, businessdao.CreateInput{
		Name:     "foo",
		Category: businessdao.CategoryPets,
	})
	assert.NoError(t, err)

	businessURL := fmt.Sprintf("/businesses/%s", business.ID)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, businessURL, nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetBusiness(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response businessdao.Business
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, business.ID, response.ID)
}

func TestGetAllBusinesses_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestBusinessDao(ctx, t)
	server := NewServer(dao, nil, nil, nil, nil)

	_, err := dao.Create(ctx, businessdao.CreateInput{
		Name:     "foo",
		Category: businessdao.CategoryPets,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, businessdao.CreateInput{
		Name:     "bar",
		Category: businessdao.CategoryPets,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, businessdao.CreateInput{
		Name:     "baz",
		Category: businessdao.CategoryAutomotive,
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/businesses", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetAllBusinesses(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response []businessdao.Business
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(response))
}

func TestGetAllBusinessesByCategory_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestBusinessDao(ctx, t)
	server := NewServer(dao, nil, nil, nil, nil)

	_, err := dao.Create(ctx, businessdao.CreateInput{
		Name:     "foo",
		Category: businessdao.CategoryPets,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, businessdao.CreateInput{
		Name:     "bar",
		Category: businessdao.CategoryPets,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, businessdao.CreateInput{
		Name:     "baz",
		Category: businessdao.CategoryAutomotive,
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/businesses?category=pets", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetAllBusinesses(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response []businessdao.Business
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response))
}

func TestGetAllBusinessesByCategory_Invalid(t *testing.T) {
	ctx := context.Background()
	dao := createTestBusinessDao(ctx, t)
	server := NewServer(dao, nil, nil, nil, nil)

	_, err := dao.Create(ctx, businessdao.CreateInput{
		Name:     "foo",
		Category: businessdao.CategoryPets,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, businessdao.CreateInput{
		Name:     "bar",
		Category: businessdao.CategoryPets,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, businessdao.CreateInput{
		Name:     "baz",
		Category: businessdao.CategoryAutomotive,
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/businesses?category=pokemon", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetAllBusinesses(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	raw, err := io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"error":"invalid category"}`, string(raw))
}
