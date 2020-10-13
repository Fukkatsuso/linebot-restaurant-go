package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	Lat      float64  `json:"lat,omitempty" datastore:"lat,noindex"`
	Lng      float64  `json:"lng,omitempty" datastore:"lng,noindex"`
	Keywords []string `json:"keywords,omitempty" datastore:"keywords,noindex"`
	Radius   int      `json:"radius,omitempty" datastore:"raduis,noindex"`
	Page     int      `json:"page,omitempty" datastore:"page,noindex"`
	PlaceID  string   `json:"place_id,omitempty"`
}

type Restaurant struct {
	PlaceID      string  `json:"place_id" datastore:"place_id,noindex"`
	Name         string  `json:"name" datastore:"name,noindex"`
	Rating       float64 `json:"rating" datastore:"rating,noindex"`
	PhotoURI     string  `json:"photo_uri" datastore:"photo_uri,noindex"`
	GoogleMapURI string  `json:"googlemap_uri" datastore:"googlemap_uri,noindex"`
}

type Favorite struct {
	List []Restaurant `datastore:"list,noindex"`
}

func (favorite Favorite) MarshalMessage(maxBubble int) *linebot.FlexMessage {
	carousel := favorite.MarshalCarousel(maxBubble)
	return linebot.NewFlexMessage("お気に入りリスト", &carousel)
}

func (favorite Favorite) MarshalCarousel(maxBubble int) linebot.CarouselContainer {
	bubbleContainers := make([]*linebot.BubbleContainer, 0)
	for i := 0; i < len(favorite.List) && i < maxBubble; i++ {
		bubble := favorite.List[i].MarshalBubble()
		bubbleContainers = append(bubbleContainers, &bubble)
	}
	carousel := linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbleContainers,
	}
	return carousel
}

func (restaurant *Restaurant) MarshalBubble() linebot.BubbleContainer {
	jsonBytes, _ := json.Marshal(&struct {
		PlaceID string `json:"place_id"`
	}{
		PlaceID: restaurant.PlaceID,
	})
	return linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Size: linebot.FlexBubbleSizeTypeKilo,
		Hero: &linebot.ImageComponent{
			Type:       linebot.FlexComponentTypeImage,
			URL:        restaurant.PhotoURI,
			Size:       linebot.FlexImageSizeTypeFull,
			AspectMode: linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   restaurant.Name,
					Size:   linebot.FlexTextSizeTypeLg,
					Weight: linebot.FlexTextWeightTypeBold,
					Wrap:   true,
				},
				&linebot.BoxComponent{
					Type:     linebot.FlexComponentTypeBox,
					Layout:   linebot.FlexBoxLayoutTypeBaseline,
					Contents: places.RatingStars(restaurant.Rating),
					Margin:   linebot.FlexComponentMarginTypeMd,
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Action: linebot.NewPostbackAction("お気に入りから削除", `{"action": "deleteFavorite", "query": `+string(jsonBytes)+`}`, "", ""),
					Height: linebot.FlexButtonHeightTypeSm,
				},
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Action: linebot.NewURIAction("マップで見る", restaurant.GoogleMapURI),
					Height: linebot.FlexButtonHeightTypeSm,
				},
			},
		},
	}
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
	params := map[string]string{
		"language": "ja",
		"key":      apiKey,
		"type":     "restaurant",
		"location": float64ToString(q.Lat) + "," + float64ToString(q.Lng),
		"radius":   strconv.Itoa(q.Radius),
	}
	if len(q.Keywords) > 0 {
		params["keyword"] = strings.Join(q.Keywords, "+")
	}
	return places.BuildURI(apiType, params)
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
	label := map[string]string{}
	if len(q.Keywords) == 0 {
		label["narrowDownKeyword"] = "キーワードで絞り込み"
	} else {
		label["narrowDownKeyword"] = "キーワードを設定し直す"
	}
	actions := []linebot.TemplateAction{
		linebot.NewPostbackAction("距離で絞り込み", `{"action": "narrowDownRadius", "query": `+string(jsonBytes)+`}`, "", ""),
		linebot.NewPostbackAction(label["narrowDownKeyword"], `{"action": "narrowDownKeyword", "query": `+string(jsonBytes)+`}`, "", ""),
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
					case "お気に入りを見る":
						// ユーザのお気に入りリストを取得
						favorite := &Favorite{}
						err := Get(favorite, ctx, dsClient, event.Source.UserID, nil)
						if err == datastore.ErrNoSuchEntity || len(favorite.List) == 0 {
							reply = linebot.NewTextMessage("お気に入りがありません")
						} else {
							reply = favorite.MarshalMessage(10)
						}
					default:
						query := &Query{}
						if err := Get(query, ctx, dsClient, event.Source.UserID, nil); err != nil {
							continue
						}
						query.Keywords = append(query.Keywords, message.Text)
						if err := Save(query, ctx, dsClient, event.Source.UserID, nil); err != nil {
							continue
						}
						reply = query.SearchConfirmButton()
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
					reply = linebot.NewTextMessage("キーワードを入力してネ\n送ったメッセージの数だけキーワードが追加されます!")
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
					if len(nearbyPlaces) == 0 {
						reply = linebot.NewTextMessage("見つかりませんでした(´・ω・`)")
					} else {
						reply = nearbyPlaces.MarshalMessage(10)
					}
				case "addFavorite":
					placeID := postbackData.Query.PlaceID
					place, _ := places.SearchByID(placeID, placesAPIKey)
					restaurant := &Restaurant{placeID, place.Name, place.Rating, place.PhotoURI(map[string]string{"key": placesAPIKey, "maxwidth": "350"}), place.URL}
					// ユーザのお気に入りリストを取得
					favorite := &Favorite{}
					err := Get(favorite, ctx, dsClient, event.Source.UserID, nil)
					if err == datastore.ErrNoSuchEntity {
						// エンティティがなければ作成
						favorite.List = []Restaurant{}
					} else if err != nil {
						continue
					}
					// お気に入りに追加
					// 登録済みか否か
					has := false
					for _, place := range favorite.List {
						if placeID == place.PlaceID {
							has = true
							break
						}
					}
					if has {
						reply = linebot.NewTextMessage("このお店は登録済みです")
					} else if len(favorite.List) == 10 {
						reply = linebot.NewTextMessage("お気に入りに登録できるのは最大10件です")
					} else {
						favorite.List = append(favorite.List, *restaurant)
						err := Save(favorite, ctx, dsClient, event.Source.UserID, nil)
						if err == nil {
							reply = linebot.NewTextMessage(fmt.Sprintf("お気に入りに登録しました! (%d/10)", len(favorite.List)))
						}
					}
				case "deleteFavorite":
					placeID := postbackData.Query.PlaceID
					// ユーザのお気に入りリストを取得
					favorite := &Favorite{}
					err := Get(favorite, ctx, dsClient, event.Source.UserID, nil)
					if err != nil {
						continue
					}
					newList := []Restaurant{}
					had := false
					for _, place := range favorite.List {
						if placeID == place.PlaceID {
							had = true
						} else {
							newList = append(newList, place)
						}
					}
					if !had {
						reply = linebot.NewTextMessage("すでに削除されています")
					} else {
						favorite.List = newList
						if err := Save(favorite, ctx, dsClient, event.Source.UserID, nil); err != nil {
							reply = linebot.NewTextMessage("削除に失敗しました...")
						} else {
							reply = linebot.NewTextMessage("お気に入り登録から削除しました!")
						}
					}
				}
			}
			if _, err := bot.ReplyMessage(event.ReplyToken, reply).Do(); err != nil {
				log.Print(err)
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
