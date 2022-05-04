package appointmentdao

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrAppointmentNotFound = errors.New("appointment not found")

type Appointment struct {
	ID         string    `json:"id" firestore:"-"`
	ClientID   string    `json:"clientId" firestore:"clientId"`
	BusinessID string    `json:"businessId" firestore:"businessId"`
	Date       time.Time `json:"date" firestore:"date"`
}

type Dao struct {
	fsClient                  *firestore.Client
	appointmentCollectionName string
}

func NewDao(client *firestore.Client, appointmentCollectionName string) *Dao {
	return &Dao{
		fsClient:                  client,
		appointmentCollectionName: appointmentCollectionName,
	}
}

func (dao *Dao) GetAppointment(ctx context.Context, id string) (*Appointment, error) {
	docRef := dao.fsClient.Collection(dao.appointmentCollectionName).Doc(id)
	snapshot, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}
	var appointment Appointment
	if err := snapshot.DataTo(&appointment); err != nil {
		return nil, err
	}

	appointment.ID = docRef.ID
	return &appointment, nil
}

type GetAllAppointmentsInput struct {
	// OPTIONAL to filter by businesses
	BusinessID string
	ClientID   string
}

func (dao *Dao) GetAllAppointments(ctx context.Context, input GetAllAppointmentsInput) ([]Appointment, error) {
	query := dao.fsClient.Collection(dao.appointmentCollectionName).Query
	if input.BusinessID != "" {
		query = query.Where("businessId", "==", input.BusinessID)
	}
	if input.ClientID != "" {
		query = query.Where("clientId", "==", input.ClientID)
	}
	snapshots, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	appointments := make([]Appointment, 0, len(snapshots))
	for _, snapshot := range snapshots {
		var appointment Appointment
		if err := snapshot.DataTo(&appointment); err != nil {
			return nil, err
		}
		appointment.ID = snapshot.Ref.ID
		appointments = append(appointments, appointment)
	}

	return appointments, nil
}

// TODO: give it a time field
type CreateInput struct {
	ClientID   string
	BusinessID string
}

func (dao *Dao) Create(ctx context.Context, input CreateInput) (*Appointment, error) {
	appointment := Appointment{
		ClientID:   input.ClientID,
		BusinessID: input.BusinessID,
	}

	doc, _, err := dao.fsClient.Collection(dao.appointmentCollectionName).Add(ctx, appointment)
	if err != nil {
		return nil, err
	}
	appointment.ID = doc.ID
	return &appointment, nil
}

func (dao *Dao) Delete(ctx context.Context, id string) error {
	docRef := dao.fsClient.Collection(dao.appointmentCollectionName).Doc(id)
	_, err := docRef.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}
