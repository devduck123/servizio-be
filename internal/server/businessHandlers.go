package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/devduck123/servizio-be/internal/businessdao"
)

func (s *Server) GetBusiness(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetBusiness called on:", r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/businesses/")

	business, err := s.businessDao.GetBusiness(r.Context(), id)
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

func (s *Server) GetAllBusinesses(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAllBusinesses called on:", r.URL.Path, "on", r.URL.RawQuery)

	category := strings.TrimPrefix(r.URL.RawQuery, "category=")
	input := businessdao.GetAllBusinessesInput{
		Category: businessdao.Category(category),
	}

	allBusinesses, err := s.businessDao.GetAllBusinesses(r.Context(), input)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}
	if !input.Category.IsValid() && input.Category != "" {
		writeErrorJSON(w, http.StatusBadRequest, errors.New("invalid category"))
		return
	}

	writeJSON(w, http.StatusOK, allBusinesses)
}

type BusinessCreateInput struct {
	Name     string               `json:"name"`
	Category businessdao.Category `json:"category"`
}

func (s *Server) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateBusiness called on :", r.URL.Path)

	user, err := UserFromContext(r.Context())
	if err != nil {
		writeErrorJSON(w, http.StatusUnauthorized, err)
		return
	}

	var businessCreateInput BusinessCreateInput
	err = json.NewDecoder(r.Body).Decode(&businessCreateInput)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	fmt.Printf("%+v\n", businessCreateInput)

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
		UserID:   user.ID,
	}
	business, err := s.businessDao.Create(r.Context(), businessToCreateInput)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, business)
}

func (s *Server) DeleteBusiness(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteBusiness called on:", r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/businesses/")
	err := s.businessDao.Delete(r.Context(), id)
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

// TODO: file size limit, file type limit, consider form api
// TODO: review imageURL
func (s *Server) UploadImageBusiness(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UploadImage called on:", r.URL.Path)

	ctx := r.Context()

	trimmedURL := strings.TrimSuffix(r.URL.Path, "/")
	id := strings.TrimSuffix(strings.TrimPrefix(trimmedURL, "/businesses/"), "/images")

	fmt.Println("id:", id)
	_, err := s.businessDao.GetBusiness(ctx, id)
	if err != nil {
		if err == businessdao.ErrBusinessNotFound {
			writeErrorJSON(w, http.StatusNotFound, err)
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	if len(raw) == 0 {
		writeErrorJSON(w, http.StatusBadRequest, errors.New("no image provided"))
		return
	}

	image, err := s.imageManager.UploadImage(ctx, id, raw)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	imageURL := fmt.Sprintf("servizio-be.appspot.com/%v/%v", id, image.Key)
	if err := s.businessDao.AppendImage(ctx, id, imageURL); err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, "success")
}
