package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/devduck123/servizio-be/internal/businessdao"
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
		log.Println("connecting to local running firestore database on localhost:8080")
		if err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080"); err != nil {
			log.Fatal("failed to set FIRESTORE_EMULATOR_HOST environment variable", err)
		}
	}

	client, err := firestore.NewClient(ctx, "servizio-be")
	if err != nil {
		return err
	}

	businessDao := businessdao.NewDao(client, "businesses")
	s := server.NewServer(businessDao)
	http.HandleFunc("/businesses/", s.BusinessRouter)
	fmt.Println("listening on port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		return err
	}

	return nil
}

// TODO: put port in variable
