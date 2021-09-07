package bot

import (
	"cloud.google.com/go/datastore"
	"github.com/line/line-bot-sdk-go/linebot"
)

type Bot struct {
	LINEBotClient   *linebot.Client
	DatastoreClient *datastore.Client
	GCPPlacesAPIKey string
}

func NewBot(linebotClient *linebot.Client, datastoreClient *datastore.Client, gcpPlacesAPIKey string) *Bot {
	return &Bot{
		LINEBotClient:   linebotClient,
		DatastoreClient: datastoreClient,
		GCPPlacesAPIKey: gcpPlacesAPIKey,
	}
}
