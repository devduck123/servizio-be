package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/devduck123/servizio-be/internal/appointmentdao"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/devduck123/servizio-be/internal/clientdao"
	"github.com/devduck123/servizio-be/internal/images"
)

type Server struct {
	businessDao    *businessdao.Dao
	clientDao      *clientdao.Dao
	appointmentDao *appointmentdao.Dao
	imageManager   *images.ImageManager
	app            *firebase.App
}

func NewServer(businessDao *businessdao.Dao, clientDao *clientdao.Dao, appointmentDao *appointmentdao.Dao, imageManager *images.ImageManager, app *firebase.App) *Server {
	return &Server{
		businessDao:    businessDao,
		clientDao:      clientDao,
		appointmentDao: appointmentDao,
		imageManager:   imageManager,
		app:            app,
	}
}

type User struct {
	ID string
}

var UserKey struct{}

func (s *Server) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		idToken := r.Header.Get("Authorization")

		client, err := s.app.Auth(ctx)
		if err != nil {
			fmt.Printf("error getting Auth client: %v\n", err)
			writeErrorJSON(w, http.StatusInternalServerError, errors.New("something went wrong"))
			return
		}

		token, err := client.VerifyIDToken(ctx, idToken)
		if err != nil {
			fmt.Printf("error verifying ID token: %v\n", err)
			writeErrorJSON(w, http.StatusUnauthorized, errors.New("invalid credentials"))
			return
		}

		expiresAt := time.Unix(token.Expires, 0)
		if time.Now().After(expiresAt) {
			writeErrorJSON(w, http.StatusUnauthorized, errors.New("token expired"))
			return
		}

		emailVerified := token.Claims["email_verified"].(bool)
		if !emailVerified {
			writeErrorJSON(w, http.StatusUnauthorized, errors.New("email not verified"))
			return
		}

		user := User{
			ID: token.Subject,
		}
		// update context with user object and UID
		r = r.WithContext(context.WithValue(ctx, UserKey, user))

		next(w, r)
	}
}

func UserFromContext(ctx context.Context) (User, error) {
	rawUser := ctx.Value(UserKey)
	if rawUser == nil {
		return User{}, errors.New("user object missing")
	}

	user, ok := rawUser.(User)
	if !ok {
		panic("invalid user object")
	}

	return user, nil
}

func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

func (s *Server) Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		next(w, r)
		duration := time.Since(t)

		fmt.Printf("%v: %v %v took %v\n", t.Format(time.RFC3339), r.Method, r.URL.Path, duration)
	}
}

func (s *Server) BusinessRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		businessID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/businesses"), "/")
		fmt.Println("id:", businessID)
		if businessID == "" {
			s.GetAllBusinesses(w, r)
			return
		}
		s.GetBusiness(w, r)
		return
	case http.MethodPost:
		trimmedURL := strings.TrimSuffix(r.URL.Path, "/")
		if strings.HasSuffix(trimmedURL, "/images") {
			s.Authenticate(s.UploadImageBusiness)(w, r)
			return
		}
		s.Authenticate(s.CreateBusiness)(w, r)
		return
	case http.MethodDelete:
		s.Authenticate(s.DeleteBusiness)(w, r)
		return
	default:
		writeErrorJSON(w, http.StatusNotImplemented, fmt.Errorf("%v not implemented yet", r.Method))
	}
}

func (s *Server) ClientRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		clientID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/clients"), "/")
		if clientID == "" {
			s.GetAllClients(w, r)
			return
		}
		s.GetClient(w, r)
		return
	case http.MethodPost:
		trimmedURL := strings.TrimSuffix(r.URL.Path, "/")
		if strings.HasSuffix(trimmedURL, "/images") {
			s.Authenticate(s.UploadImageClient)(w, r)
			return
		}
		s.Authenticate(s.CreateClient)(w, r)
		return
	case http.MethodDelete:
		s.Authenticate(s.DeleteClient)(w, r)
		return
	default:
		writeErrorJSON(w, http.StatusNotImplemented, fmt.Errorf("%v not implemented yet", r.Method))
	}
}

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
	fmt.Println("GetAllBusinesses called on:", r.URL.Path)
	fmt.Println("GetAllBusinesses called on:", r.URL.RawQuery)

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
