package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/devduck123/servizio-be/internal/clientdao"
)

func (s *Server) GetClient(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetClient called on:", r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/clients/")

	client, err := s.clientDao.GetClient(r.Context(), id)
	if err != nil {
		if errors.Is(err, clientdao.ErrClientNotFound) {
			writeErrorJSON(w, http.StatusNotFound, err)
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, client)
}

func (s *Server) GetAllClients(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAllClients called on:", r.URL.Path)

	allClients, err := s.clientDao.GetAllClients(r.Context())
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, allClients)
}

type ClientCreateInput struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

func (s *Server) CreateClient(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateClient called on :", r.URL.Path)

	user, err := UserFromContext(r.Context())
	if err != nil {
		writeErrorJSON(w, http.StatusUnauthorized, err)
		return
	}

	var clientCreateInput ClientCreateInput
	err = json.NewDecoder(r.Body).Decode(&clientCreateInput)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	fmt.Printf("%+v\n", clientCreateInput)

	if strings.TrimSpace(clientCreateInput.FirstName) == "" || strings.TrimSpace(clientCreateInput.LastName) == "" {
		writeErrorJSON(w, http.StatusBadRequest, errors.New("name cannot be empty"))
		return
	}

	clientToCreateInput := clientdao.CreateInput{
		FirstName: clientCreateInput.FirstName,
		LastName:  clientCreateInput.LastName,
		UserID:    user.ID,
	}
	client, err := s.clientDao.Create(r.Context(), clientToCreateInput)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, client)
}

func (s *Server) DeleteClient(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeleteClient called on:", r.URL.Path)

	id := strings.TrimPrefix(r.URL.Path, "/clients/")
	err := s.clientDao.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, clientdao.ErrClientNotFound) {
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
func (s *Server) UploadImageClient(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UploadImage called on:", r.URL.Path)

	ctx := r.Context()

	trimmedURL := strings.TrimSuffix(r.URL.Path, "/")
	id := strings.TrimSuffix(strings.TrimPrefix(trimmedURL, "/clients/"), "/images")

	fmt.Println("id:", id)
	_, err := s.clientDao.GetClient(ctx, id)
	if err != nil {
		if err == clientdao.ErrClientNotFound {
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
	if err := s.clientDao.AppendImage(ctx, id, imageURL); err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, "success")
}
