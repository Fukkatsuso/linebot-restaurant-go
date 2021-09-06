package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/config"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	port := os.Getenv("PORT")

	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, config.DatastoreProjectID)
	if err != nil {
		log.Fatalf("Could not create datastore client: %v", err)
	}
	defer dsClient.Close()

	bot, err := linebot.New(config.LINEChannelSecret, config.LINEChannelToken)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/callback", CallbackHandler(ctx, bot, dsClient))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
