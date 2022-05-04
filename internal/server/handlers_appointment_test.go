package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/devduck123/servizio-be/internal/appointmentdao"
	"github.com/google/uuid"
	"github.com/tj/assert"
)

func createTestAppointmentDao(ctx context.Context, t *testing.T, keepCollection ...bool) *appointmentdao.Dao {
	t.Helper()

	client, err := firestore.NewClient(ctx, projectID)
	assert.NoError(t, err)
	appointmentCollection := fmt.Sprintf("appointment-%v", uuid.New())
	dao := appointmentdao.NewDao(client, appointmentCollection)

	cleanUp := true
	for _, kc := range keepCollection {
		if kc {
			cleanUp = false
			break
		}
	}

	if cleanUp {
		t.Cleanup(func() {
			deleteCollection(ctx, t, client, appointmentCollection)
		})
	}

	return dao
}

func TestCreateAppointment(t *testing.T) {
	ctx := context.Background()
	dao := createTestAppointmentDao(ctx, t)
	server := NewServer(nil, nil, dao, nil, nil)

	w := httptest.NewRecorder()
	body := bytes.NewReader([]byte(`{
		"clientId":"foo",
		"businessId":"bar",
		"date":"2022-12-02T15:04:05+07:00"
	}`))
	r := httptest.NewRequest(http.MethodPost, "/", body)
	r = r.WithContext(ContextWithUser(ctx, User{}))
	server.CreateAppointment(w, r)
	
	raw, err := ioutil.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	fmt.Println(string(raw))

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}
