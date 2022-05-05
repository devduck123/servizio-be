package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/devduck123/servizio-be/internal/clientdao"
	"github.com/google/uuid"
	"github.com/tj/assert"
)

func createTestClientDao(ctx context.Context, t *testing.T, keepCollection ...bool) *clientdao.Dao {
	t.Helper()

	client, err := firestore.NewClient(ctx, projectID)
	assert.NoError(t, err)
	clientCollection := fmt.Sprintf("client-%v", uuid.New())
	dao := clientdao.NewDao(client, clientCollection)

	cleanUp := true
	for _, kc := range keepCollection {
		if kc {
			cleanUp = false
			break
		}
	}

	if cleanUp {
		t.Cleanup(func() {
			deleteCollection(ctx, t, client, clientCollection)
		})
	}

	return dao
}

func TestCreateClient_Invalid(t *testing.T) {
	ctx := context.Background()
	dao := createTestClientDao(ctx, t)
	server := NewServer(nil, dao, nil, nil, nil)

	w := httptest.NewRecorder()
	body := bytes.NewReader([]byte(`{}`))
	r := httptest.NewRequest(http.MethodPost, "/", body)
	r = r.WithContext(ContextWithUser(ctx, User{}))
	server.CreateClient(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	raw, err := ioutil.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"error":"first name cannot be empty"}`, string(raw))
}

func TestCreateClient_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestClientDao(ctx, t)
	server := NewServer(nil, dao, nil, nil, nil)

	w := httptest.NewRecorder()
	body := bytes.NewReader([]byte(`{
		"firstname":"foo",
		"lastname":"bar"
	}`))
	r := httptest.NewRequest(http.MethodPost, "/", body)
	r = r.WithContext(ContextWithUser(ctx, User{}))
	server.CreateClient(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response clientdao.Client
	err := json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "foo", response.FirstName)
	assert.Equal(t, "bar", response.LastName)
}

func TestGetClient_Invalid(t *testing.T) {
	ctx := context.Background()
	dao := createTestClientDao(ctx, t)
	server := NewServer(nil, dao, nil, nil, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/clients/foo", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetClient(w, r)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

	var response clientdao.Client
	err := json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)
}

func TestGetClient_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestClientDao(ctx, t)
	server := NewServer(nil, dao, nil, nil, nil)

	client, err := dao.Create(ctx, clientdao.CreateInput{
		FirstName: "foo",
		LastName:  "bar",
	})
	assert.NoError(t, err)

	clientURL := fmt.Sprintf("/clients/%s", client.ID)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, clientURL, nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetClient(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response clientdao.Client
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, client.ID, response.ID)
}

func TestGetAllClients_Valid(t *testing.T) {
	ctx := context.Background()
	dao := createTestClientDao(ctx, t)
	server := NewServer(nil, dao, nil, nil, nil)

	_, err := dao.Create(ctx, clientdao.CreateInput{
		FirstName: "abraham",
		LastName:  "lincoln",
	})
	assert.NoError(t, err)

	_, err = dao.Create(ctx, clientdao.CreateInput{
		FirstName: "george",
		LastName:  "washington",
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/businesses", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetAllClients(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response []clientdao.Client
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response))
}
