package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"

	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/places"
)

// Query is data of user's request
type Query struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Keyword *string `json:"keyword"`
	Radius  int     `json:"radius"`
	Page    int     `json:"page"`
}

type PostbackData struct {
	Action string `json:"action"`
	Query  Query  `json:"query"`
}

func (q *Query) BuildURI(apiType, apiKey string) string {
	return places.BuildURI(apiType, map[string]string{
		"language": "ja",
		"key":      apiKey,
		"type":     "restaurant",
		"location": float64ToString(q.Lat) + "," + float64ToString(q.Lng),
		"radius":   strconv.Itoa(q.Radius),
	})
}

func (q *Query) SearchConfirmButton() *linebot.TemplateMessage {
	jsonBytes, _ := json.Marshal(q)
	actions := []linebot.TemplateAction{
		linebot.NewPostbackAction("距離で絞り込み", `{"action": "narrowDownRadius", "query": `+string(jsonBytes)+`}`, "", ""),
		linebot.NewPostbackAction("キーワードで絞り込み", `{"action": "narrowDownKeyword", "query": `+string(jsonBytes)+`}`, "", ""),
		linebot.NewPostbackAction("検索する", `{"action": "search", "query": `+string(jsonBytes)+`}`, "", ""),
	}
	buttons := linebot.NewButtonsTemplate("", "確認", "絞り込みますか？", actions...)
	return linebot.NewTemplateMessage("確認ボタン", buttons)
}

func (q *Query) RadiusQuickReply() linebot.SendingMessage {
	radiusKey := []string{"100m", "250m", "500m", "1km", "2km", "5km"}
	radiusValue := []int{100, 250, 500, 1000, 2000, 5000}
	buttons := make([]*linebot.QuickReplyButton, 0)
	for i := range radiusKey {
		q.Radius = radiusValue[i]
		jsonBytes, _ := json.Marshal(q)
		queryString := `{"action": "setRadius", "query": ` + string(jsonBytes) + `}`
		b := linebot.NewQuickReplyButton("", linebot.NewPostbackAction(radiusKey[i], queryString, "", radiusKey[i]))
		buttons = append(buttons, b)
	}
	textMsg := linebot.NewTextMessage("検索範囲を選択してネ")
	return textMsg.WithQuickReplies(linebot.NewQuickReplyItems(buttons...))
}

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
					case "キーワード検索":
						reply = linebot.NewTextMessage("ごめんなさい。今はまだ使えません(><)")
					default:
						reply = linebot.NewTextMessage(message.Text)
					}
				case *linebot.LocationMessage:
					query := Query{
						Lat:     message.Latitude,
						Lng:     message.Longitude,
						Keyword: nil,
						Radius:  500,
						Page:    0,
					}
					reply = query.SearchConfirmButton()
				}
				if _, err := bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
					log.Print(err)
				}
			} else if event.Type == linebot.EventTypePostback {
				postbackData := new(PostbackData)
				if err := json.Unmarshal([]byte(event.Postback.Data), postbackData); err != nil {
					log.Print(err)
					continue
				}
				switch postbackData.Action {
				case "narrowDownRadius":
					reply = postbackData.Query.RadiusQuickReply()
				case "narrowDownKeyword":
					reply = linebot.NewTextMessage("ごめんなさい。今はまだ使えません(><)")
				case "setRadius":
					reply = postbackData.Query.SearchConfirmButton()
				case "search":
					uri := postbackData.Query.BuildURI("nearbysearch", placesAPIKey)
					fmt.Println("[URI]", uri)
					nearbyPlaces, err := places.Search(uri)
					if err != nil {
						log.Print(err)
						continue
					}
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
