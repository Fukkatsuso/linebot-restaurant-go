package bot

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
	MaxPlaces int = 10
)

func (bot *Bot) CallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		events, err := bot.LINEBotClient.ParseRequest(r)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		bot.LineEventsController(ctx, events)
	}
}

func (bot *Bot) LineEventsController(ctx context.Context, events []*linebot.Event) {
	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			bot.HandleMessage(ctx, event)
		case linebot.EventTypePostback:
			bot.HandlePostback(ctx, event)
		}
	}
}

func (bot *Bot) HandleMessage(ctx context.Context, event *linebot.Event) {
	switch event.Message.(type) {
	case *linebot.TextMessage:
		bot.HandleTextMessage(ctx, event)
	case *linebot.LocationMessage:
		bot.HandleLocationMessage(ctx, event)
	}
}

func (bot *Bot) HandleTextMessage(ctx context.Context, event *linebot.Event) {
	msg := event.Message.(*linebot.TextMessage)
	text := msg.Text
	switch text {
	case "位置情報検索":
		bot.ReplyMessage(ctx, event, LocationSendButton())
	case "お気に入りを見る":
		bot.ShowFavorite(ctx, event)
	default:
		bot.AddKeyword(ctx, event)
	}
}

// 返信
func (bot *Bot) ReplyMessage(ctx context.Context, event *linebot.Event, messages ...linebot.SendingMessage) {
	replyToken := event.ReplyToken
	if _, err := bot.LINEBotClient.ReplyMessage(replyToken, messages...).Do(); err != nil {
		log.Print(err)
	}
}

// お気に入りを表示
func (bot *Bot) ShowFavorite(ctx context.Context, event *linebot.Event) {
	userID := event.Source.UserID
	f := mystore.Favorite{}
	err := mystore.Get(ctx, bot.DatastoreClient, &f, userID, nil)
	if err == datastore.ErrNoSuchEntity || len(f.List) == 0 {
		bot.ReplyMessage(ctx, event, TextMessage("お気に入りがありません"))
		return
	}
	favoritePlaces := FavoritePlaces(f.List)
	bot.ReplyMessage(ctx, event, CarouselMessage(&favoritePlaces, MaxPlaces))
}

// 検索クエリにキーワードを追加
func (bot *Bot) AddKeyword(ctx context.Context, event *linebot.Event) {
	userID := event.Source.UserID
	q := mystore.Query{}
	if err := mystore.Get(ctx, bot.DatastoreClient, &q, userID, nil); err != nil {
		bot.ReplyMessage(ctx, event, TextMessage("位置情報を送信して「キーワードで絞り込み」を選択してください"))
		return
	}
	keyword := event.Message.(*linebot.TextMessage).Text
	q.Keywords = append(q.Keywords, keyword)
	if err := mystore.Save(ctx, bot.DatastoreClient, &q, userID, nil); err != nil {
		bot.ReplyMessage(ctx, event, TextMessage("キーワードの保存に失敗しました．\nもう一度送信してくださいm(__)m"))
		return
	}
	bot.ReplyMessage(ctx, event, SearchConfirmWindow((*Query)(&q)))
}

func (bot *Bot) HandleLocationMessage(ctx context.Context, event *linebot.Event) {
	msg := event.Message.(*linebot.LocationMessage)
	lat, lng := float64ToString(msg.Latitude), float64ToString(msg.Longitude)
	q := Query{
		Lat:      lat,
		Lng:      lng,
		Keywords: []string{},
		Radius:   "500",
		Page:     0,
	}
	bot.ReplyMessage(ctx, event, SearchConfirmWindow(&q))
}

func float64ToString(s float64) string {
	return strconv.FormatFloat(s, 'f', -1, 64)
}

func (bot *Bot) HandlePostback(ctx context.Context, event *linebot.Event) {
	postback := Postback{}
	if err := json.Unmarshal([]byte(event.Postback.Data), &postback); err != nil {
		log.Print(err)
		return
	}

	data := postback.Data
	switch postback.Action {
	case PostbackActionChangeRadius:
		bot.ChangeRadius(ctx, event, data.(*Query))
	case PostbackActionChangeKeyword:
		bot.ChangeKeyword(ctx, event, data.(*Query))
	case PostbackActionUpdateRadius:
		bot.UpdateRadius(ctx, event, data.(*Query))
	case PostbackActionNearbySearch:
		bot.ShowNearbyPlaces(ctx, event, data.(*Query))
	case PostbackActionAddFavorite:
		bot.AddFavorite(ctx, event, data.(*PlaceInfo))
	case PostbackActionDeleteFavorite:
		bot.DeleteFavorite(ctx, event, data.(*PlaceInfo))
	}
}

func (bot *Bot) ChangeRadius(ctx context.Context, event *linebot.Event, q *Query) {
	userID := event.Source.UserID
	if err := mystore.Save(ctx, bot.DatastoreClient, (*mystore.Query)(q), userID, nil); err != nil {
		return
	}
	bot.ReplyMessage(ctx, event, RadiusQuickReply(q))
}
func (bot *Bot) ChangeKeyword(ctx context.Context, event *linebot.Event, q *Query) {
	userID := event.Source.UserID
	q.Keywords = []string{}
	if err := mystore.Save(ctx, bot.DatastoreClient, (*mystore.Query)(q), userID, nil); err != nil {
		return
	}
	bot.ReplyMessage(ctx, event, TextMessage("キーワードを入力してネ\n送ったメッセージの数だけキーワードが追加されます!"))
}

func (bot *Bot) UpdateRadius(ctx context.Context, event *linebot.Event, q *Query) {
	bot.ReplyMessage(ctx, event, SearchConfirmWindow(q))
}

func (bot *Bot) ShowNearbyPlaces(ctx context.Context, event *linebot.Event, q *Query) {
	p, err := bot.NearbySearch(q)
	if err != nil {
		log.Print(err)
		bot.ReplyMessage(ctx, event, TextMessage("検索に失敗しました..."))
		return
	}
	if len(*p) == 0 {
		bot.ReplyMessage(ctx, event, TextMessage("見つかりませんでした(´・ω・`)"))
	} else {
		bot.ReplyMessage(ctx, event, CarouselMessage((*NearbyPlaces)(p), MaxPlaces))
	}
}

func (bot *Bot) AddFavorite(ctx context.Context, event *linebot.Event, info *PlaceInfo) {
	placeID := info.PlaceID
	p, err := bot.DetailsSearch(placeID)
	if err != nil {
		log.Print(err)
		bot.ReplyMessage(ctx, event, TextMessage("お気に入り登録に失敗しました..."))
		return
	}

	userID := event.Source.UserID

	// Datastoreからリストを取得してお気に入り追加
	f := mystore.Favorite{}
	err = mystore.Get(ctx, bot.DatastoreClient, &f, userID, nil)
	if err == datastore.ErrNoSuchEntity {
		// エンティティがなければ作成
		f.List = []places.Place{}
	} else if err != nil {
		bot.ReplyMessage(ctx, event, TextMessage("お気に入り登録に失敗しました..."))
		return
	}
	// お気に入りに追加
	// 登録済みか否かチェック
	for _, place := range f.List {
		if placeID == place.PlaceID {
			bot.ReplyMessage(ctx, event, TextMessage("このお店は登録済みです"))
			return
		}
	}
	if len(f.List) == MaxPlaces {
		text := fmt.Sprintf("お気に入りに登録できるのは最大%d件です", MaxPlaces)
		bot.ReplyMessage(ctx, event, TextMessage(text))
		return
	}
	// 検索結果表示に使ったものと同じ画像
	p.PhotoURI = info.PhotoURI
	f.List = append(f.List, *p)
	if err := mystore.Save(ctx, bot.DatastoreClient, &f, userID, nil); err != nil {
		bot.ReplyMessage(ctx, event, TextMessage("お気に入り登録に失敗しました..."))
		return
	}

	text := fmt.Sprintf("お気に入りに登録しました! (%d/%d)", len(f.List), MaxPlaces)
	bot.ReplyMessage(ctx, event, TextMessage(text))
}

func (bot *Bot) DeleteFavorite(ctx context.Context, event *linebot.Event, info *PlaceInfo) {
	userID := event.Source.UserID
	// お気に入りリストを取得
	f := mystore.Favorite{}
	err := mystore.Get(ctx, bot.DatastoreClient, &f, userID, nil)
	if err != nil {
		bot.ReplyMessage(ctx, event, TextMessage("お気に入り削除に失敗しました..."))
		return
	}
	// 削除操作後の新たなリスト
	placeID := info.PlaceID
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
		bot.ReplyMessage(ctx, event, TextMessage("すでに削除されています"))
		return
	}
	f.List = newList
	if err := mystore.Save(ctx, bot.DatastoreClient, &f, userID, nil); err != nil {
		bot.ReplyMessage(ctx, event, TextMessage("お気に入り削除に失敗しました..."))
		return
	}
	bot.ReplyMessage(ctx, event, TextMessage("お気に入り登録から削除しました!"))
}
