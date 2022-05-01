package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/devduck123/servizio-be/internal/authtest"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/devduck123/servizio-be/internal/images"
	"github.com/google/uuid"
	"github.com/tj/assert"
)

var projectID = "servizio-be"

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

func createTestDao(ctx context.Context, t *testing.T, keepCollection ...bool) *businessdao.Dao {
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

func deleteCollection(ctx context.Context, t *testing.T, fsClient *firestore.Client, collection string) {
	documentRefs, err := fsClient.Collection(collection).DocumentRefs(ctx).GetAll()
	assert.NoError(t, err)

	for _, document := range documentRefs {
		_, err := document.Delete(ctx)
		assert.NoError(t, err)
	}
}

func TestCreateBusiness_Invalid(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	server := NewServer(dao, nil, nil, nil)

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
	dao := createTestDao(ctx, t)
	server := NewServer(dao, nil, nil, nil)

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
	dao := createTestDao(ctx, t)
	server := NewServer(dao, nil, nil, nil)

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
	dao := createTestDao(ctx, t)
	server := NewServer(dao, nil, nil, nil)

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
	dao := createTestDao(ctx, t)
	server := NewServer(dao, nil, nil, nil)

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

	assert.Equal(t, len(response), 3)
}

func TestGetAllBusinessesByCategory_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	server := NewServer(dao, nil, nil, nil)

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

	assert.Equal(t, len(response), 2)
}

func TestGetAllBusinessesByCategory_Invalid(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	server := NewServer(dao, nil, nil, nil)

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

func createTestImageManager(ctx context.Context, t *testing.T) (*images.ImageManager, func()) {
	client, err := storage.NewClient(ctx)
	assert.NoError(t, err)

	im := &images.ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com",
	}

	cleanUp := func() {
		client.Close()
	}

	return im, cleanUp
}

func TestUploadImage(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	client, err := storage.NewClient(ctx)
	assert.NoError(t, err)
	defer client.Close()

	business, err := dao.Create(ctx, businessdao.CreateInput{
		Name: "foo",
	})
	assert.NoError(t, err)

	im, cleanUp := createTestImageManager(ctx, t)
	defer cleanUp()
	server := NewServer(dao, nil, im, nil)
	body := bytes.NewReader([]byte("hello"))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/businesses/%v/images/", business.ID), body)

	server.UploadImage(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	gotBody, err := ioutil.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, `"success"`, string(gotBody))
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	assert.NoError(t, err)
	server := NewServer(dao, nil, nil, app) // This is the constructor that creates a "server"

	httpServer := httptest.NewServer(http.HandlerFunc(server.BusinessRouter)) // This spins up a HTTP test server.
	defer httpServer.Close()

	body := bytes.NewReader([]byte(`{
		"name": "test",
		"category": "pets"
	}`))
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/businesses/", body)
	assert.NoError(t, err)
	token, err := authtest.GetJWT(t, projectID)
	assert.NoError(t, err)
	req.Header.Set("Authorization", token)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response businessdao.Business
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "test", response.Name)
	assert.EqualValues(t, "pets", response.Category)
	assert.NotEmpty(t, response.ID)
}

func TestServer_GetBusiness(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	assert.NoError(t, err)
	server := NewServer(dao, nil, nil, app) // This is the constructor that creates a "server"

	input := businessdao.CreateInput{
		Name:     "foo",
		Category: businessdao.CategoryPets,
	}
	business, err := dao.Create(ctx, input)
	assert.NoError(t, err)

	httpServer := httptest.NewServer(http.HandlerFunc(server.BusinessRouter)) // This spins up a HTTP test server.
	defer httpServer.Close()

	req, err := http.NewRequest(http.MethodGet, httpServer.URL+"/businesses/"+business.ID, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response businessdao.Business
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "foo", response.Name)
	assert.EqualValues(t, businessdao.CategoryPets, response.Category)
	assert.Equal(t, business.ID, response.ID)
}

func TestServer_GetAllBusinesses(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	assert.NoError(t, err)
	server := NewServer(dao, nil, nil, app) // This is the constructor that creates a "server"

	input := businessdao.CreateInput{
		Name:     "foo",
		Category: businessdao.CategoryPets,
	}
	_, err = dao.Create(ctx, input)
	assert.NoError(t, err)

	httpServer := httptest.NewServer(http.HandlerFunc(server.BusinessRouter)) // This spins up a HTTP test server.
	defer httpServer.Close()

	req, err := http.NewRequest(http.MethodGet, httpServer.URL+"/businesses", nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response []businessdao.Business
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(response))
}

func TestUserFromContext_NotFound(t *testing.T) {
	ctx := context.Background()
	_, err := UserFromContext(ctx)
	assert.EqualError(t, err, "user object missing")
}

func TestUserFromContext_Found(t *testing.T) {
	ctx := context.Background()
	user := User{
		ID: "123",
	}
	ctx = context.WithValue(ctx, UserKey, user)
	gotUser, err := UserFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, gotUser.ID)

}
