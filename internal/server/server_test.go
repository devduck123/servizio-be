package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/devduck123/servizio-be/internal/businessdao"
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
	server := NewServer(dao, nil)

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
	server := NewServer(dao, nil)

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
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	assert.NoError(t, err)
	server := NewServer(dao, app) // This is the constructor that creates a "server"

	httpServer := httptest.NewServer(http.HandlerFunc(server.BusinessRouter)) // This spins up a HTTP test server.
	defer httpServer.Close()

	body := bytes.NewReader([]byte(`{
		"name": "test",
		"category": "pets"
	}`))
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/businesses/", body)
	assert.NoError(t, err)
	token, err := getJWT(t)
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

func getJWT(t *testing.T) (string, error) {
	t.Helper()

	email := fmt.Sprintf("test-%v@test.com", uuid.New())
	password := "tester"

	token, err := signUpAndVerify(t, email, password)
	assert.NoError(t, err)

	return token, nil
}

func signUpAndVerify(t *testing.T, email, password string) (string, error) {
	t.Helper()

	err := signUp(t, email, password)
	assert.NoError(t, err)

	err = verifyEmail(t, email)
	assert.NoError(t, err)

	token, err := signIn(t, email, password)
	assert.NoError(t, err)

	return token, nil
}

func signUp(t *testing.T, email, password string) error {
	t.Helper()

	signUpURL := "http://localhost:9099/identitytoolkit.googleapis.com/v1/accounts:signUp?key=test"
	body := fmt.Sprintf(`{
		"email": "%v",
		"password": "%v"
		}`, email, password)

	res, err := http.Post(signUpURL, "application/json", bytes.NewReader([]byte(body)))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	return nil
}

// returns no error if successful
func verifyEmail(t *testing.T, email string) error {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	assert.NoError(t, err)
	authClient, err := app.Auth(ctx)
	assert.NoError(t, err)
	emailVerificationLink, err := authClient.EmailVerificationLink(ctx, email)
	assert.NoError(t, err)

	res, err := http.Get(emailVerificationLink)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	return nil
}

func signIn(t *testing.T, email, password string) (string, error) {
	signInURL := "http://localhost:9099/identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=test"
	body := fmt.Sprintf(`{
		"email": "%v",
		"password": "%v"
		}`, email, password)

	res, err := http.Post(signInURL, "application/json", bytes.NewReader([]byte(body)))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)

	token := response["idToken"].(string)
	if token == "" {
		return "", errors.New("no token present")
	}

	return token, nil
}
