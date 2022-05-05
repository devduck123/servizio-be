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
	"time"

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

func TestGetAppointment(t *testing.T) {
	ctx := context.Background()
	dao := createTestAppointmentDao(ctx, t)
	server := NewServer(nil, nil, dao, nil, nil)

	inputTime, err := time.Parse("2006-01-02 15:04", "2023-04-20 04:35")
	assert.NoError(t, err)
	appointment, err := dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "foo",
		BusinessID: "bar",
		Date:       inputTime,
	})
	assert.NoError(t, err)

	appointmentURL := fmt.Sprintf("/appointments/%s", appointment.ID)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, appointmentURL, nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetAppointment(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response appointmentdao.Appointment
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, appointment.ID, response.ID)
	assert.Equal(t, appointment.Date, response.Date)
}

func TestGetAllAppointments(t *testing.T) {
	ctx := context.Background()
	dao := createTestAppointmentDao(ctx, t)
	server := NewServer(nil, nil, dao, nil, nil)

	inputTime, err := time.Parse("2006-01-02 15:04", "2023-04-20 04:35")
	assert.NoError(t, err)
	_, err = dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "foo",
		BusinessID: "bar",
		Date:       inputTime,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "abe",
		BusinessID: "linc",
		Date:       inputTime,
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/appointments", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetAllAppointments(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response []appointmentdao.Appointment
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response))
}

func TestGetAllAppointments_ByClientId(t *testing.T) {
	ctx := context.Background()
	dao := createTestAppointmentDao(ctx, t)
	server := NewServer(nil, nil, dao, nil, nil)

	inputTime, err := time.Parse("2006-01-02 15:04", "2023-04-20 04:35")
	assert.NoError(t, err)
	_, err = dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "foo",
		BusinessID: "bar",
		Date:       inputTime,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "foo",
		BusinessID: "baz",
		Date:       inputTime,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "duck",
		BusinessID: "bar",
		Date:       inputTime,
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/appointments?client=foo", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetAllAppointments(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response []appointmentdao.Appointment
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response))
}

func TestGetAllAppointments_ByClientId_ByBusinessId(t *testing.T) {
	ctx := context.Background()
	dao := createTestAppointmentDao(ctx, t)
	server := NewServer(nil, nil, dao, nil, nil)

	inputTime, err := time.Parse("2006-01-02 15:04", "2023-04-20 04:35")
	assert.NoError(t, err)
	_, err = dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "foo",
		BusinessID: "bar",
		Date:       inputTime,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "foo",
		BusinessID: "bar",
		Date:       inputTime,
	})
	assert.NoError(t, err)
	_, err = dao.Create(ctx, appointmentdao.CreateInput{
		ClientID:   "duck",
		BusinessID: "bar",
		Date:       inputTime,
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/appointments?client=foo&business=bar", nil)
	r = r.WithContext(ContextWithUser(ctx, User{}))

	server.GetAllAppointments(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var response []appointmentdao.Appointment
	err = json.NewDecoder(w.Result().Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response))
}
