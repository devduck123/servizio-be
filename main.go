package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/devduck123/servizio-be/internal/businessdao"
	"github.com/devduck123/servizio-be/internal/clientdao"
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
	}

	projectID := "servizio-be"

	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	businessDao := businessdao.NewDao(fsClient, "businesses")
	clientDao := clientdao.NewDao(fsClient, "clients")
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}

	s := server.NewServer(businessDao, clientDao, app)
	http.HandleFunc("/businesses/", s.Logger(s.BusinessRouter))
	fmt.Println("listening on port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		return err
	}

	return nil
}

// TODO: put port in variable
