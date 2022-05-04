package appointmentdao

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
	appointmentCollection := fmt.Sprintf("appointment-%v", uuid.New())

	dao := NewDao(fsClient, appointmentCollection)

	t.Cleanup(func() {
		firestoretest.DeleteCollection(ctx, t, fsClient, appointmentCollection)
	})

	return dao
}

func TestGetAppointmentByID(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	input := CreateInput{
		ClientID:   "foo",
		BusinessID: "bar",
	}
	appointment, err := dao.Create(ctx, input)
	assert.NoError(t, err)

	gotAppointment, err := dao.GetAppointment(ctx, appointment.ID)
	assert.NoError(t, err)
	assert.NotNil(t, gotAppointment)
	assert.Equal(t, appointment.ID, gotAppointment.ID)
	assert.Equal(t, input.BusinessID, gotAppointment.BusinessID)
	assert.Equal(t, input.ClientID, gotAppointment.ClientID)
}

func TestGetAppointment_NotExists(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	gotAppointment, err := dao.GetAppointment(ctx, "notexists")
	assert.Equal(t, "appointment not found", err.Error())
	assert.Nil(t, gotAppointment)
}

func TestGetAllAppointments_ByClientID(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	createInput := CreateInput{
		ClientID:   "client1",
		BusinessID: "google",
	}
	createInput2 := CreateInput{
		ClientID:   "client2",
		BusinessID: "google",
	}
	createInput3 := CreateInput{
		ClientID:   "client3",
		BusinessID: "apple",
	}
	_, err := dao.Create(ctx, createInput)
	assert.NoError(t, err)
	_, err = dao.Create(ctx, createInput2)
	assert.NoError(t, err)
	_, err = dao.Create(ctx, createInput3)
	assert.NoError(t, err)

	getAllAppointmentsInput := GetAllAppointmentsInput{
		BusinessID: "google",
	}
	allAppointments, err := dao.GetAllAppointments(ctx, getAllAppointmentsInput)
	assert.NoError(t, err)
	assert.Len(t, allAppointments, 2)
}

// TODO: consider.. should all appointments ever be viewable??? no?
func TestGetAllAppointments(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	createInput := CreateInput{
		ClientID:   "client1",
		BusinessID: "google",
	}
	createInput2 := CreateInput{
		ClientID:   "client2",
		BusinessID: "google",
	}
	createInput3 := CreateInput{
		ClientID:   "client3",
		BusinessID: "apple",
	}
	_, err := dao.Create(ctx, createInput)
	assert.NoError(t, err)
	_, err = dao.Create(ctx, createInput2)
	assert.NoError(t, err)
	_, err = dao.Create(ctx, createInput3)
	assert.NoError(t, err)

	getAllAppointmentsInput := GetAllAppointmentsInput{}
	allAppointments, err := dao.GetAllAppointments(ctx, getAllAppointmentsInput)
	assert.NoError(t, err)
	assert.Len(t, allAppointments, 3)
}

func TestCreateAppointment(t *testing.T) {
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	input := CreateInput{
		ClientID:   "foo",
		BusinessID: "bar",
	}
	appointment, err := dao.Create(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.NotEmpty(t, appointment.ID)
	assert.Equal(t, input.ClientID, appointment.ClientID)
	assert.Equal(t, input.BusinessID, appointment.BusinessID)
}

func TestDeleteAppointment(t *testing.T) {
	// first create the appointment
	ctx := context.Background()
	dao := createTestDao(ctx, t)

	input := CreateInput{
		ClientID:   "",
		BusinessID: "",
	}
	appointment, err := dao.Create(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.NotEmpty(t, appointment.ID)
	assert.Equal(t, input.ClientID, appointment.ClientID)
	assert.Equal(t, input.BusinessID, appointment.BusinessID)

	// then delete the appointment
	err = dao.Delete(ctx, appointment.ID)
	assert.NoError(t, err)
}
