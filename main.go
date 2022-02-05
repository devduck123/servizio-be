package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
	client, err := firestore.NewClient(ctx, "")
	if err != nil {
		return err
	}
	businessDao := businessdao.NewDao(client)
	s := server.NewServer(businessDao)
	http.HandleFunc("/businesses/", s.GetBusiness)
	fmt.Println("listening on port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		return err
	}

	return nil
}

// TODO: put port in variable
