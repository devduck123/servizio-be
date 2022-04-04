package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/devduck123/servizio-be/internal/clientdao"
)

type Server struct {
	businessDao *businessdao.Dao
	app         *firebase.App
}

func NewServer(businessDao *businessdao.Dao, clientDao *clientdao.Dao, app *firebase.App) *Server {
	return &Server{
		businessDao: businessDao,
		app:         app,
	}
}

type User struct {
	ID string
}

var UserKey struct{}

func (s *Server) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// check req header here
		// ensure JWT is valid
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
		s.GetBusiness(w, r)
		return
	case http.MethodPost:
		s.Authenticate(s.CreateBusiness)(w, r)
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

type BusinessCreateInput struct {
	Name     string               `json:"name"`
	Category businessdao.Category `json:"category"`
}

func (s *Server) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	user, err := UserFromContext(r.Context())
	if err != nil {
		writeErrorJSON(w, http.StatusUnauthorized, err)
		return
	}

	// TODO: review this
	fmt.Println(r.URL.Path)
	var businessCreateInput BusinessCreateInput
	err = json.NewDecoder(r.Body).Decode(&businessCreateInput)
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
	// TODO: review this
	fmt.Println(r.URL.Path)

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
