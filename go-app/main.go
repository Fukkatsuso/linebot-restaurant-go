package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/bot"
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

	lineBot, err := linebot.New(config.LINEChannelSecret, config.LINEChannelToken)
	if err != nil {
		log.Fatal(err)
	}

	bot := bot.NewBot(lineBot, dsClient, config.GCPPlacesAPIKey)

	http.HandleFunc("/callback", bot.CallbackHandler())

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
