package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/datastore"
	mystore "github.com/Fukkatsuso/linebot-restaurant-go/go-app/datastore"
	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/places"
	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	// MaxPlaces is maximum size of carousel message
	MaxPlaces = 10
)

// ReplyMessage executes sending messages
func ReplyMessage(bot *linebot.Client, replyToken string, messages ...linebot.SendingMessage) {
	if _, err := bot.ReplyMessage(replyToken, messages...).Do(); err != nil {
		log.Print(err)
	}
}

// CallbackHandler handles "/callback"
func CallbackHandler(ctx context.Context, bot *linebot.Client, dsClient *datastore.Client) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		LineEventsController(ctx, events, bot, dsClient)
	}
}

// LineEventsController controller
func LineEventsController(ctx context.Context, events []*linebot.Event, bot *linebot.Client, dsClient *datastore.Client) {
	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			MessageController(ctx, event, bot, dsClient)
		case linebot.EventTypePostback:
			PostbackController(ctx, event, bot, dsClient)
		}
	}
}

// MessageController controller
func MessageController(ctx context.Context, event *linebot.Event, bot *linebot.Client, dsClient *datastore.Client) {
	userID := event.Source.UserID
	replyToken := event.ReplyToken
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		TextMessageController(ctx, message, userID, replyToken, bot, dsClient)
	case *linebot.LocationMessage:
		LocationMessageController(ctx, message, userID, replyToken, bot, dsClient)
	}
}

// TextMessageController controller
func TextMessageController(ctx context.Context, message *linebot.TextMessage, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	text := message.Text
	switch text {
	case "位置情報検索":
		ReplyMessage(bot, replyToken, LocationSendButton())
	case "お気に入りを見る":
		ShowFavorite(ctx, userID, replyToken, bot, dsClient)
	default:
		AddKeyword(ctx, text, userID, replyToken, bot, dsClient)
	}
}

// LocationMessageController controller
func LocationMessageController(ctx context.Context, message *linebot.LocationMessage, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	lat, lng := float64ToString(message.Latitude), float64ToString(message.Longitude)
	query := Query{
		Lat:      lat,
		Lng:      lng,
		Keywords: []string{},
		Radius:   "500",
		Page:     0,
	}
	ReplyMessage(bot, replyToken, SearchConfirmWindow(&query))
}

// PostbackController controller
func PostbackController(ctx context.Context, event *linebot.Event, bot *linebot.Client, dsClient *datastore.Client) {
	postback := new(Postback)
	if err := json.Unmarshal([]byte(event.Postback.Data), postback); err != nil {
		log.Print(err)
		return
	}
	data := postback.Data
	userID := event.Source.UserID
	replyToken := event.ReplyToken

	switch postback.Action {
	case PostbackActionChangeRadius:
		ChangeRadius(ctx, data.(*Query), userID, replyToken, bot, dsClient)
	case PostbackActionChangeKeyword:
		ChangeKeyword(ctx, data.(*Query), userID, replyToken, bot, dsClient)
	case PostbackActionUpdateRadius:
		UpdateRadius(ctx, data.(*Query), userID, replyToken, bot, dsClient)
	case PostbackActionNearbySearch:
		ShowNearbyPlaces(ctx, data.(*Query), userID, replyToken, bot, dsClient)
	case PostbackActionAddFavorite:
		AddFavorite(ctx, data.(*PlaceInfo), userID, replyToken, bot, dsClient)
	case PostbackActionDeleteFavorite:
		DeleteFavorite(ctx, data.(*PlaceInfo), userID, replyToken, bot, dsClient)
	}
}

// ChangeRadius responds to changeRadius postback
func ChangeRadius(ctx context.Context, q *Query, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	if err := mystore.Save(ctx, dsClient, (*mystore.Query)(q), userID, nil); err != nil {
		return
	}
	ReplyMessage(bot, replyToken, RadiusQuickReply(q))
}

// ChangeKeyword responds to changeKeyword postback
func ChangeKeyword(ctx context.Context, q *Query, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	q.Keywords = []string{}
	if err := mystore.Save(ctx, dsClient, (*mystore.Query)(q), userID, nil); err != nil {
		return
	}
	ReplyMessage(bot, replyToken, Text("キーワードを入力してネ\n送ったメッセージの数だけキーワードが追加されます!"))
}

// UpdateRadius updates radius
func UpdateRadius(ctx context.Context, q *Query, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	ReplyMessage(bot, replyToken, SearchConfirmWindow(q))
}

// AddKeyword adds keyword
func AddKeyword(ctx context.Context, keyword, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	q := new(Query)
	if err := mystore.Get(ctx, dsClient, (*mystore.Query)(q), userID, nil); err != nil {
		ReplyMessage(bot, replyToken, Text("位置情報を送信して「キーワードで絞り込み」を選択してください"))
		return
	}
	q.Keywords = append(q.Keywords, keyword)
	if err := mystore.Save(ctx, dsClient, (*mystore.Query)(q), userID, nil); err != nil {
		ReplyMessage(bot, replyToken, Text("キーワードの保存に失敗しました．\nもう一度送信してくださいm(__)m"))
		return
	}
	ReplyMessage(bot, replyToken, SearchConfirmWindow(q))
}

// ShowNearbyPlaces shows nearby search result
func ShowNearbyPlaces(ctx context.Context, q *Query, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	p := new(NearbyPlaces)
	uri, err := NearbySearch(q, (*places.Places)(p))
	log.Println("[URI]", uri)
	if err != nil {
		log.Print(err)
		ReplyMessage(bot, replyToken, Text("検索に失敗しました..."))
		return
	}
	if len(*p) == 0 {
		ReplyMessage(bot, replyToken, Text("見つかりませんでした(´・ω・`)"))
	} else {
		ReplyMessage(bot, replyToken, Carousel(p, MaxPlaces))
	}
}

// ShowFavorite shows user's favorite
func ShowFavorite(ctx context.Context, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	f := new(mystore.Favorite)
	err := mystore.Get(ctx, dsClient, f, userID, nil)
	if err == datastore.ErrNoSuchEntity || len(f.List) == 0 {
		ReplyMessage(bot, replyToken, Text("お気に入りがありません"))
		return
	}
	favoritePlaces := (*FavoritePlaces)(&f.List)
	ReplyMessage(bot, replyToken, Carousel(favoritePlaces, MaxPlaces))
}

// AddFavorite adds favorite
func AddFavorite(ctx context.Context, info *PlaceInfo, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	placeID := info.PlaceID
	p := new(places.Place)
	uri, err := DetailsSearch(placeID, p)
	log.Println("[URI]", uri)
	if err != nil {
		log.Print(err)
		ReplyMessage(bot, replyToken, Text("お気に入り登録に失敗しました..."))
		return
	}
	// Datastoreからリストを取得してお気に入り追加
	f := new(mystore.Favorite)
	err = mystore.Get(ctx, dsClient, f, userID, nil)
	if err == datastore.ErrNoSuchEntity {
		// エンティティがなければ作成
		f.List = []places.Place{}
	} else if err != nil {
		ReplyMessage(bot, replyToken, Text("お気に入り登録に失敗しました..."))
		return
	}
	// お気に入りに追加
	// 登録済みか否かチェック
	for _, place := range f.List {
		if placeID == place.PlaceID {
			ReplyMessage(bot, replyToken, Text("このお店は登録済みです"))
			return
		}
	}
	if len(f.List) == MaxPlaces {
		text := fmt.Sprintf("お気に入りに登録できるのは最大%d件です", MaxPlaces)
		ReplyMessage(bot, replyToken, Text(text))
		return
	}
	// 検索結果表示に使ったものと同じ画像
	p.PhotoURI = info.PhotoURI
	f.List = append(f.List, *p)
	if err := mystore.Save(ctx, dsClient, f, userID, nil); err != nil {
		ReplyMessage(bot, replyToken, Text("お気に入り登録に失敗しました..."))
		return
	}
	text := fmt.Sprintf("お気に入りに登録しました! (%d/%d)", len(f.List), MaxPlaces)
	ReplyMessage(bot, replyToken, Text(text))
}

// DeleteFavorite deletes user's favorite
func DeleteFavorite(ctx context.Context, info *PlaceInfo, userID, replyToken string, bot *linebot.Client, dsClient *datastore.Client) {
	placeID := info.PlaceID
	// お気に入りリストを取得
	f := new(mystore.Favorite)
	err := mystore.Get(ctx, dsClient, f, userID, nil)
	if err != nil {
		ReplyMessage(bot, replyToken, Text("お気に入り削除に失敗しました..."))
		return
	}
	// 削除操作後の新たなリスト
	newList := []places.Place{}
	had := false
	for _, place := range f.List {
		if placeID == place.PlaceID {
			had = true
		} else {
			newList = append(newList, place)
		}
	}
	if !had {
		ReplyMessage(bot, replyToken, Text("すでに削除されています"))
		return
	}
	f.List = newList
	if err := mystore.Save(ctx, dsClient, f, userID, nil); err != nil {
		ReplyMessage(bot, replyToken, Text("お気に入り削除に失敗しました..."))
		return
	}
	ReplyMessage(bot, replyToken, Text("お気に入り登録から削除しました!"))
}

func float64ToString(s float64) string {
	return strconv.FormatFloat(s, 'f', -1, 64)
}
