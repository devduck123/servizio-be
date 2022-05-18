package servizio

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/devduck123/servizio-be/internal/appointmentdao"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/devduck123/servizio-be/internal/clientdao"
	"github.com/devduck123/servizio-be/internal/images"
	"github.com/devduck123/servizio-be/internal/server"
)

func setupServer() (http.Handler, error) {
	ctx := context.Background()

	projectID := "servizio-be"

	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	businessDao := businessdao.NewDao(fsClient, "businesses")
	clientDao := clientdao.NewDao(fsClient, "clients")
	appointmentDao := appointmentdao.NewDao(fsClient, "appointments")
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	im := &images.ImageManager{
		API:        client,
		BucketName: "servizio-be.appspot.com",
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}

	s := server.NewServer(businessDao, clientDao, appointmentDao, im, app)

	sm := http.NewServeMux()
	sm.Handle("/businesses/", s.Logger(s.CORS(s.BusinessRouter)))
	sm.Handle("/clients/", s.Logger(s.CORS(s.ClientRouter)))
	sm.Handle("/appointments/", s.Logger(s.CORS(s.AppointmentRouter)))

	return sm, nil
}

func init() {
	server, err := setupServer()
	if err != nil {
		panic(err)
	}

	functions.HTTP("servizio", server.ServeHTTP)
}
