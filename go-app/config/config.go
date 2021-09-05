package config

import (
	"log"
	"os"
)

// LINE
var (
	LINEChannelID     string
	LINEChannelSecret string
	LINEChannelToken  string
)

func initEnvLINE() {
	LINEChannelID = os.Getenv("LINE_CHANNEL_ID")
	LINEChannelSecret = os.Getenv("LINE_CHANNEL_SECRET")
	LINEChannelToken = os.Getenv("LINE_CHANNEL_TOKEN")
}

// GCP
var (
	GCPPlacesAPIKey    string
	DatastoreProjectID string
)

func initEnvGCP() {
	GCPPlacesAPIKey = os.Getenv("GCP_PLACES_API_KEY")
	DatastoreProjectID = os.Getenv("DATASTORE_PROJECT_ID")
	if DatastoreProjectID == "" {
		log.Fatal(`You need to set the environment variable "DATASTORE_PROJECT_ID"`)
	}
}

func init() {
	initEnvLINE()
	initEnvGCP()
}
