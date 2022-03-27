package authtest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	firebase "firebase.google.com/go/v4"
	"github.com/google/uuid"
	"github.com/tj/assert"
)

func GetJWT(t *testing.T, projectID string) (string, error) {
	t.Helper()

	email := fmt.Sprintf("test-%v@test.com", uuid.New())
	password := "tester"

	token, err := signUpAndVerify(t, email, password, projectID)
	assert.NoError(t, err)

	return token, nil
}

func signUpAndVerify(t *testing.T, email, password, projectID string) (string, error) {
	t.Helper()

	err := signUp(t, email, password)
	assert.NoError(t, err)

	err = verifyEmail(t, email, projectID)
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
func verifyEmail(t *testing.T, email, projectID string) error {
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
