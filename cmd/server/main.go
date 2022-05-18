package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/devduck123/servizio-be/internal/appointmentdao"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/devduck123/servizio-be/internal/clientdao"
	"github.com/devduck123/servizio-be/internal/images"
	"github.com/devduck123/servizio-be/internal/server"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("done")
}

func run(ctx context.Context) error {
	local := flag.Bool("local", false, "local connects to a local running firestore database")
	flag.Parse()
	if local != nil && *local {
		if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080"); err != nil {
			log.Fatal("failed to set FIRESTORE_EMULATOR_HOST environment variable", err)
		}
		log.Println("connecting to local running firestore database on localhost:8080")

		if err := os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099"); err != nil {
			log.Fatal("failed to set FIREBASE_AUTH_EMULATOR_HOST environment variable", err)
		}
		log.Println("connecting to firebase auth on localhost:9099")

		if err := os.Setenv("STORAGE_EMULATOR_HOST", "localhost:9199"); err != nil {
			log.Fatal("failed to set STORAGE_EMULATOR_HOST environment variable", err)
		}
		log.Println("connecting to firebase storage on localhost:9199")
	}

	projectID := "servizio-be"

	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	businessDao := businessdao.NewDao(fsClient, "businesses")
	clientDao := clientdao.NewDao(fsClient, "clients")
	appointmentDao := appointmentdao.NewDao(fsClient, "appointments")
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
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
		return err
	}

	s := server.NewServer(businessDao, clientDao, appointmentDao, im, app)
	http.HandleFunc("/businesses/", s.CORS(s.Logger(s.BusinessRouter)))
	http.HandleFunc("/clients/", s.CORS(s.Logger(s.ClientRouter)))
	http.HandleFunc("/appointments/", s.CORS(s.Logger(s.AppointmentRouter)))
	fmt.Println("listening on port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		return err
	}

	return nil
}

// TODO: put port in variable
