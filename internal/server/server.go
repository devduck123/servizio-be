package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/devduck123/servizio-be/internal/businessdao"
)

type Server struct {
	BusinessDao *businessdao.Dao
}

func NewServer(businessDao *businessdao.Dao) *Server {
	return &Server{
		BusinessDao: businessDao,
	}
}

func (s *Server) BusinessRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.GetBusiness(w, r)
		return
	case http.MethodPost:
		s.CreateBusiness(w, r)
		return
	case http.MethodDelete:
		s.DeleteBusiness(w, r)
		return
	default:
		writeErrorJSON(w, http.StatusNotImplemented, fmt.Errorf("%v not implemented yet", r.Method))
	}
}

func (s *Server) GetBusiness(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/businesses/")

	business, err := s.BusinessDao.GetBusiness(r.Context(), id)
	if err != nil {
		if errors.Is(err, businessdao.ErrBusinessNotFound) {
			writeErrorJSON(w, http.StatusNotFound, err)
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, business)
}

type BusinessCreateInput struct {
	Name     string               `json:"name"`
	Category businessdao.Category `json:"category"`
}

func (s *Server) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	// TODO: review this
	fmt.Println(r.URL.Path)
	var businessCreateInput BusinessCreateInput
	err := json.NewDecoder(r.Body).Decode(&businessCreateInput)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(businessCreateInput.Name) == "" {
		writeErrorJSON(w, http.StatusBadRequest, errors.New("name cannot be empty"))
		return
	}
	// TODO: add valid categories in error message
	if !businessCreateInput.Category.IsValid() {
		writeErrorJSON(w, http.StatusBadRequest, errors.New("invalid category"))
		return
	}


	businessToCreateInput := businessdao.CreateInput{
		Name:     businessCreateInput.Name,
		Category: businessCreateInput.Category,
	}
	business, err := s.BusinessDao.Create(r.Context(), businessToCreateInput)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, business)
}

func (s *Server) DeleteBusiness(w http.ResponseWriter, r *http.Request) {
	// TODO: review this
	fmt.Println(r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/businesses/")
	err := s.BusinessDao.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, businessdao.ErrBusinessNotFound) {
			writeErrorJSON(w, http.StatusNotFound, err)
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, "successful deletion")
}

func writeErrorJSON(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var response struct {
		Error string `json:"error"`
	}
	response.Error = err.Error()
	raw, _ := json.Marshal(response)
	w.Write(raw)
}

func writeJSON(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	raw, _ := json.Marshal(response)
	w.Write(raw)
}
