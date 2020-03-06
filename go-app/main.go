package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"

	"./places"
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
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					var reply linebot.SendingMessage
					switch message.Text {
					case "位置情報検索":
						locationURIAction := linebot.NewURIAction("送信する", "line://nv/location")
						locationButton := linebot.NewButtonsTemplate("", "", "位置情報を送信してネ", locationURIAction)
						reply = linebot.NewTemplateMessage("位置情報送信ボタン", locationButton)
					default:
						reply = linebot.NewTextMessage(message.Text)
					}
					if _, err = bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
						log.Print(err)
					}
				case *linebot.LocationMessage:
					params := map[string]string{
						"language": "ja",
						"type":     "restaurant",
						"key":      os.Getenv("GCP_PLACES_API_KEY"),
						"location": float64ToString(message.Latitude) + "," + float64ToString(message.Longitude),
						"radius":   "500",
					}
					uri := places.BuildURI("nearbysearch", params)
					fmt.Println("[URI]", uri)
					result, err := places.Search(uri)
					if err != nil {
						fmt.Println(err)
						continue
					}
					var rests string
					for i := range result.Results {
						rests += result.Results[i].Name + "\n"
					}
					reply := linebot.NewTextMessage(rests)
					if _, err = bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
						log.Print(err)
					}
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
