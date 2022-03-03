package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/google/uuid"
	"github.com/tj/assert"
)

func TestMain(m *testing.M) {
	if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080"); err != nil {
		log.Fatal("failed to set FIRESTORE_EMULATOR_HOST environment variable", err)
	}

	m.Run()
}

func createTestDao(ctx context.Context, t *testing.T, keepCollection ...bool) *businessdao.Dao {
	t.Helper()

	client, err := firestore.NewClient(ctx, "servizio-be")
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
	server := NewServer(dao)

	w := httptest.NewRecorder()
	body := bytes.NewReader([]byte(`{}`))
	r := httptest.NewRequest(http.MethodPost, "/", body)
	server.CreateBusiness(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	raw, err := ioutil.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"error":"name cannot be empty"}`, string(raw))
}

func TestCreateBusiness_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	server := NewServer(dao)

	w := httptest.NewRecorder()
	body := bytes.NewReader([]byte(`{
		"name": "test",
		"category": "pets"
	}`))
	r := httptest.NewRequest(http.MethodPost, "/", body)
	server.CreateBusiness(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response businessdao.Business
	err := json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "test", response.Name)
	assert.EqualValues(t, "pets", response.Category)
	assert.NotEmpty(t, response.ID)
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)
	server := NewServer(dao) // This is the constructor that creates a "server"

	httpServer := httptest.NewServer(http.HandlerFunc(server.BusinessRouter)) // This spins up a HTTP test server.
	defer httpServer.Close()

	body := bytes.NewReader([]byte(`{
		"name": "test",
		"category": "pets"
	}`))
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/businesses/", body)
	assert.NoError(t, err)

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
