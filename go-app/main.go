package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/line/line-bot-sdk-go/linebot"

	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/places"
)

var (
	radiusKey   = []string{"100m", "250m", "500m", "1km", "2km", "5km"}
	radiusValue = []int{100, 250, 500, 1000, 2000, 5000}
	radiusMap   = map[int]string{}
)

func init() {
	for i, val := range radiusKey {
		radiusMap[radiusValue[i]] = val
	}
}

// Query is data of user's request
type Query struct {
	Lat      float64  `json:"lat" datastore:"lat,noindex"`
	Lng      float64  `json:"lng" datastore:"lng,noindex"`
	Keywords []string `json:"keywords" datastore:"keywords,noindex"`
	Radius   int      `json:"radius" datastore:"raduis,noindex"`
	Page     int      `json:"page" datastore:"page,noindex"`
}

type Restaurant struct {
	PlaceID      string    `datastore:"place_id,noindex"`
	RegisteredAt time.Time `datastore:"registered_at"`
}

type Favorite struct {
	List []Restaurant `datastore:"list,noindex"`
}

type PostbackData struct {
	Action string `json:"action"`
	Query  Query  `json:"query"`
}

type Entity interface {
	NameKey(name string, parent *datastore.Key) *datastore.Key
}

func (query *Query) NameKey(name string, parent *datastore.Key) *datastore.Key {
	return datastore.NameKey("Query", name, parent)
}

func (restaurant *Restaurant) NameKey(name string, parent *datastore.Key) *datastore.Key {
	return datastore.NameKey("Restaurant", name, parent)
}

func (favorite *Favorite) NameKey(name string, parent *datastore.Key) *datastore.Key {
	return datastore.NameKey("Favorite", name, parent)
}

func Get(entity Entity, ctx context.Context, client *datastore.Client, name string, parent *datastore.Key) error {
	key := entity.NameKey(name, parent)
	err := client.Get(ctx, key, entity)
	log.Println("[Get]", entity, err)
	return err
}

func Save(entity Entity, ctx context.Context, client *datastore.Client, name string, parent *datastore.Key) error {
	key := entity.NameKey(name, parent)
	_, err := client.Put(ctx, key, entity)
	log.Println("[Save]", entity, err)
	return err
}

func Delete(entity Entity, ctx context.Context, client *datastore.Client, name string, parent *datastore.Key) error {
	key := entity.NameKey(name, parent)
	err := client.Delete(ctx, key)
	log.Println("[Delete]", entity, err)
	return err
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

func (q *Query) Status() string {
	var str string
	str += fmt.Sprintf("距離: %s\n", radiusMap[q.Radius])
	if len(q.Keywords) > 0 {
		str += fmt.Sprintf("キーワード: %v\n", q.Keywords)
	}
	return str
}

func (q *Query) SearchConfirmButton() *linebot.TemplateMessage {
	jsonBytes, _ := json.Marshal(q)
	actions := []linebot.TemplateAction{
		linebot.NewPostbackAction("距離で絞り込み", `{"action": "narrowDownRadius", "query": `+string(jsonBytes)+`}`, "", ""),
		linebot.NewPostbackAction("キーワードで絞り込み", `{"action": "narrowDownKeyword", "query": `+string(jsonBytes)+`}`, "", ""),
		linebot.NewPostbackAction("検索する", `{"action": "search", "query": `+string(jsonBytes)+`}`, "", ""),
	}
	buttons := linebot.NewButtonsTemplate("", "絞り込みますか？", q.Status(), actions...)
	return linebot.NewTemplateMessage("確認ボタン", buttons)
}

func (q *Query) RadiusQuickReply() linebot.SendingMessage {
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

	projID := os.Getenv("DATASTORE_PROJECT_ID")
	if projID == "" {
		log.Fatal(`You need to set the environment variable "DATASTORE_PROJECT_ID"`)
	}
	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, projID)
	if err != nil {
		log.Fatalf("Could not create datastore client: %v", err)
	}

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
						query := &Query{}
						if err := Get(query, ctx, dsClient, event.Source.UserID, nil); err != nil {
							continue
						}
						// if query does not exists
						//
						//
						query.Keywords = append(query.Keywords, message.Text)
						if err := Save(query, ctx, dsClient, event.Source.UserID, nil); err != nil {
							continue
						}
						reply = query.SearchConfirmButton()
						// linebot.NewTextMessage(message.Text)
					}
				case *linebot.LocationMessage:
					query := Query{
						Lat:      message.Latitude,
						Lng:      message.Longitude,
						Keywords: []string{},
						Radius:   500,
						Page:     0,
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
					query := &postbackData.Query
					if err := Save(query, ctx, dsClient, event.Source.UserID, nil); err != nil {
						continue
					}
					reply = query.RadiusQuickReply()
				case "narrowDownKeyword":
					query := &postbackData.Query
					query.Keywords = []string{}
					if err := Save(query, ctx, dsClient, event.Source.UserID, nil); err != nil {
						continue
					}
					reply = linebot.NewTextMessage("キーワードを入力してネ")
				case "setRadius":
					reply = postbackData.Query.SearchConfirmButton()
				case "search":
					query := &postbackData.Query
					if err := Save(query, ctx, dsClient, event.Source.UserID, nil); err != nil {
						continue
					}
					uri := query.BuildURI("nearbysearch", placesAPIKey)
					log.Println("[URI]", uri)
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
