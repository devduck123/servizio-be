package server

import (
	"encoding/json"
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

func (s *Server) GetBusiness(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	id := strings.TrimPrefix(r.URL.Path, "/businesses/")
	business, err := s.BusinessDao.GetBusiness(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	raw, _ := json.Marshal(business)
	w.Write(raw)
}
