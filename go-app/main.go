package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"

	"linebot-restaurant-go/go-app/places"
)

func main() {
	port := os.Getenv("PORT")

	bot, err := linebot.New(
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		placesAPIKey := os.Getenv("GCP_PLACES_API_KEY")
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			var reply linebot.SendingMessage
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					switch message.Text {
					case "位置情報検索":
						reply = places.LocationSendButton()
					default:
						reply = linebot.NewTextMessage(message.Text)
					}
				case *linebot.LocationMessage:
					// URI組み立て
					params := map[string]string{
						"language": "ja",
						"type":     "restaurant",
						"key":      placesAPIKey,
						"location": float64ToString(message.Latitude) + "," + float64ToString(message.Longitude),
						"radius":   "500",
					}
					uri := places.BuildURI("nearbysearch", params)
					fmt.Println("[URI]", uri)
					// Places検索実行
					nearbyPlaces, err := places.Search(uri)
					if err != nil {
						log.Print(err)
						continue
					}
					// 返信の形式に整える
					reply = nearbyPlaces.MarshalMessage(10)
				}
				if _, err := bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func float64ToString(s float64) string {
	return strconv.FormatFloat(s, 'f', -1, 64)
}
