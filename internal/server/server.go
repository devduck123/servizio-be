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

func (s *Server) CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		next(w, r)
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

func (s *Server) AppointmentRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appointmentID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/appointments"), "/")
		fmt.Println("id:", appointmentID)
		if appointmentID == "" {
			s.GetAllAppointments(w, r)
			return
		}
		s.GetAppointment(w, r)
		return
	case http.MethodPost:
		s.Authenticate(s.CreateAppointment)(w, r)
		return
	case http.MethodDelete:
		s.Authenticate(s.DeleteAppointment)(w, r)
		return
	default:
		writeErrorJSON(w, http.StatusNotImplemented, fmt.Errorf("%v not implemented yet", r.Method))
	}
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
