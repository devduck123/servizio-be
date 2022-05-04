package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/devduck123/servizio-be/internal/appointmentdao"
)

func (s *Server) GetAppointment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAppointment called on:", r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/appointments/")

	appointment, err := s.appointmentDao.GetAppointment(r.Context(), id)
	if err != nil {
		if errors.Is(err, appointmentdao.ErrAppointmentNotFound) {
			writeErrorJSON(w, http.StatusNotFound, err)
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, appointment)
}

// TODO: NOTE THAT GETALLAPPOINTMENTS MAY NOT WANT TO
// TO RETURN ALL APPOINTMENTS IN GENERAL???
func (s *Server) GetAllAppointments(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAllAppointments called on:", r.URL.Path, "on", r.URL.RawQuery)

	clientID := r.URL.Query().Get("client")
	businessID := r.URL.Query().Get("business")
	fmt.Println("clientID:", clientID)
	fmt.Println("businessID:", businessID)

	input := appointmentdao.GetAllAppointmentsInput{
		BusinessID: businessID,
		ClientID:   clientID,
	}

	allAppointments, err := s.appointmentDao.GetAllAppointments(r.Context(), input)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, allAppointments)
}

// TODO: give it a time field
type AppointmentCreateInput struct {
	ClientID   string    `json:"clientId"`
	BusinessID string    `json:"businessId"`
	Date       time.Time `json:"date"`
}

func (s *Server) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateAppointment called on:", r.URL.Path)

	_, err := UserFromContext(r.Context())
	if err != nil {
		writeErrorJSON(w, http.StatusUnauthorized, err)
		return
	}

	var appointmentCreateInput AppointmentCreateInput
	err = json.NewDecoder(r.Body).Decode(&appointmentCreateInput)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	fmt.Printf("%+v\n", appointmentCreateInput)

	if strings.TrimSpace(appointmentCreateInput.ClientID) == "" {
		writeErrorJSON(w, http.StatusBadRequest, errors.New("clientID cannot be empty"))
		return
	}
	if strings.TrimSpace(appointmentCreateInput.BusinessID) == "" {
		writeErrorJSON(w, http.StatusBadRequest, errors.New("businessID cannot be empty"))
		return
	}
	if appointmentCreateInput.Date.Before(time.Now()) {
		writeErrorJSON(w, http.StatusBadRequest, errors.New("date invalid"))
		return
	}

	appointmentToCreateInput := appointmentdao.CreateInput{
		ClientID:   appointmentCreateInput.ClientID,
		BusinessID: appointmentCreateInput.BusinessID,
		Date:       appointmentCreateInput.Date,
	}
	appointment, err := s.appointmentDao.Create(r.Context(), appointmentToCreateInput)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, appointment)
}

func (s *Server) DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteAppointment called on:", r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/appointments/")
	err := s.appointmentDao.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, appointmentdao.ErrAppointmentNotFound) {
			writeErrorJSON(w, http.StatusNotFound, err)
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, "successful deletion")
}
