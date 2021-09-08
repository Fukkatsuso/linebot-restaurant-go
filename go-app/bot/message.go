package bot

import (
	"encoding/json"
	"fmt"

	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/places"
	"github.com/line/line-bot-sdk-go/linebot"
)

var (
	radiusKey   = []string{"100m", "250m", "500m", "1km", "2km", "5km"}
	radiusValue = []string{"100", "250", "500", "1000", "2000", "5000"}
	radiusMap   = map[string]string{}
)

func init() {
	for i, val := range radiusKey {
		radiusMap[radiusValue[i]] = val
	}
}

type PostbackAction string

const (
	PostbackActionChangeRadius   PostbackAction = "changeRadius"
	PostbackActionChangeKeyword  PostbackAction = "changeKeyword"
	PostbackActionUpdateRadius   PostbackAction = "updateRadius"
	PostbackActionNearbySearch   PostbackAction = "nearbySearch"
	PostbackActionAddFavorite    PostbackAction = "addFavorite"
	PostbackActionDeleteFavorite PostbackAction = "deleteFavorite"
)

type PostbackData interface {
	PostbackData()
}

func (q *Query) PostbackData() {}

type PlaceInfo struct {
	PlaceID  string `json:"place_id"`
	PhotoURI string `json:"photo_uri"`
}

func (p *PlaceInfo) PostbackData() {}

type Postback struct {
	Action PostbackAction `json:"action"`
	Data   PostbackData   `json:"data"`
}

func PostbackJSON(action PostbackAction, pbData PostbackData) string {
	b, _ := json.Marshal(&Postback{
		Action: action,
		Data:   pbData,
	})
	return string(b)
}

func (pb *Postback) UnmarshalJSON(b []byte) error {
	type Alias Postback
	a := struct {
		Data json.RawMessage `json:"data"`
		*Alias
	}{
		Alias: (*Alias)(pb),
	}
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}

	switch a.Action {
	case PostbackActionAddFavorite, PostbackActionDeleteFavorite:
		p := new(PlaceInfo)
		if err := json.Unmarshal(a.Data, p); err != nil {
			return err
		}
		pb.Data = p
	default:
		q := new(Query)
		if err := json.Unmarshal(a.Data, q); err != nil {
			return err
		}
		pb.Data = q
	}

	return nil
}

// 位置情報送信ボタン
func LocationSendButton() *linebot.TemplateMessage {
	uriAction := linebot.NewURIAction("送信する", "line://nv/location")
	button := linebot.NewButtonsTemplate("", "", "位置情報を送信してネ", uriAction)
	return linebot.NewTemplateMessage("位置情報送信ボタン", button)
}

// 検索確認ウィンドウ
func SearchConfirmWindow(q *Query) *linebot.TemplateMessage {
	label := map[string]string{}
	if len(q.Keywords) == 0 {
		label["changeKeyword"] = "キーワードで絞り込み"
	} else {
		label["changeKeyword"] = "キーワードを設定し直す"
	}
	actions := []linebot.TemplateAction{
		linebot.NewPostbackAction("距離で絞り込み", PostbackJSON(PostbackActionChangeRadius, q), "", ""),
		linebot.NewPostbackAction(label["changeKeyword"], PostbackJSON(PostbackActionChangeKeyword, q), "", ""),
		linebot.NewPostbackAction("検索する", PostbackJSON(PostbackActionNearbySearch, q), "", ""),
	}
	buttons := linebot.NewButtonsTemplate("", "絞り込みますか？", searchStatus(q), actions...)
	return linebot.NewTemplateMessage("確認ボタン", buttons)
}

func searchStatus(q *Query) string {
	var str string
	str += fmt.Sprintf("距離: %s\n", radiusMap[q.Radius])
	if len(q.Keywords) > 0 {
		str += fmt.Sprintf("キーワード: %v\n", q.Keywords)
	}
	return str
}

// 距離絞り込み用のクイックリプライボタン
func RadiusQuickReply(q *Query) linebot.SendingMessage {
	buttons := make([]*linebot.QuickReplyButton, 0)
	for i := range radiusKey {
		q.Radius = radiusValue[i]
		postbackString := PostbackJSON(PostbackActionUpdateRadius, q)
		b := linebot.NewQuickReplyButton("", linebot.NewPostbackAction(radiusKey[i], postbackString, "", radiusKey[i]))
		buttons = append(buttons, b)
	}
	textMsg := linebot.NewTextMessage("検索範囲を選択してネ")
	return textMsg.WithQuickReplies(linebot.NewQuickReplyItems(buttons...))
}

func TextMessage(text string) *linebot.TextMessage {
	return linebot.NewTextMessage(text)
}

type PlaceBubble interface {
	MarshalBubble() *linebot.BubbleContainer
}

type NearbyPlace places.Place

type FavoritePlace places.Place

// メッセージバブルに変換
func (p *NearbyPlace) MarshalBubble() *linebot.BubbleContainer {
	info := PlaceInfo{
		PlaceID:  p.PlaceID,
		PhotoURI: p.PhotoURI,
	}
	bubble := linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Size: linebot.FlexBubbleSizeTypeKilo,
		Hero: &linebot.ImageComponent{
			Type:       linebot.FlexComponentTypeImage,
			URL:        p.PhotoURI,
			Size:       linebot.FlexImageSizeTypeFull,
			AspectMode: linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   p.Name,
					Size:   linebot.FlexTextSizeTypeLg,
					Weight: linebot.FlexTextWeightTypeBold,
					Wrap:   true,
				},
				&linebot.BoxComponent{
					Type:     linebot.FlexComponentTypeBox,
					Layout:   linebot.FlexBoxLayoutTypeBaseline,
					Contents: RatingStars(p.Rating),
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
					Action: linebot.NewPostbackAction("お気に入りに登録", PostbackJSON(PostbackActionAddFavorite, &info), "", ""),
					Height: linebot.FlexButtonHeightTypeSm,
				},
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Action: linebot.NewURIAction("マップで見る", p.GooglemapURI),
					Height: linebot.FlexButtonHeightTypeSm,
				},
			},
		},
	}
	return &bubble
}

// メッセージバブルに変換
func (p *FavoritePlace) MarshalBubble() *linebot.BubbleContainer {
	info := PlaceInfo{
		PlaceID: p.PlaceID,
	}
	bubble := linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Size: linebot.FlexBubbleSizeTypeKilo,
		Hero: &linebot.ImageComponent{
			Type:       linebot.FlexComponentTypeImage,
			URL:        p.PhotoURI,
			Size:       linebot.FlexImageSizeTypeFull,
			AspectMode: linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   p.Name,
					Size:   linebot.FlexTextSizeTypeLg,
					Weight: linebot.FlexTextWeightTypeBold,
					Wrap:   true,
				},
				&linebot.BoxComponent{
					Type:     linebot.FlexComponentTypeBox,
					Layout:   linebot.FlexBoxLayoutTypeBaseline,
					Contents: RatingStars(p.Rating),
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
					Action: linebot.NewPostbackAction("お気に入りから削除", PostbackJSON(PostbackActionDeleteFavorite, &info), "", ""),
					Height: linebot.FlexButtonHeightTypeSm,
				},
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Action: linebot.NewURIAction("マップで見る", p.GooglemapURI),
					Height: linebot.FlexButtonHeightTypeSm,
				},
			},
		},
	}
	return &bubble
}

type PlacesCarousel interface {
	PlaceBubbles(maxBubble int) []PlaceBubble
	AltText() string
	Len() int
}

// カルーセルメッセージ
func CarouselMessage(p PlacesCarousel, maxBubble int) *linebot.FlexMessage {
	carousel := MarshalCarousel(p, maxBubble)
	altText := p.AltText()
	return linebot.NewFlexMessage(altText, carousel)
}

// カルーセルに変換
func MarshalCarousel(p PlacesCarousel, maxBubble int) *linebot.CarouselContainer {
	placeBubbles := p.PlaceBubbles(maxBubble)
	bubbleContainers := make([]*linebot.BubbleContainer, 0)
	for i := range placeBubbles {
		bubble := placeBubbles[i].MarshalBubble()
		bubbleContainers = append(bubbleContainers, bubble)
	}
	carousel := linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbleContainers,
	}
	return &carousel
}

type NearbyPlaces places.Places

type FavoritePlaces places.Places

// 複数のメッセージバブルに変換
func (p *NearbyPlaces) PlaceBubbles(maxBubble int) []PlaceBubble {
	bubbles := make([]PlaceBubble, 0)
	for i := 0; i < p.Len() && i < maxBubble; i++ {
		placePtr := &(*p)[i]
		bubbles = append(bubbles, (*NearbyPlace)(placePtr))
	}
	return bubbles
}

// 複数のメッセージバブルに変換
func (p *FavoritePlaces) PlaceBubbles(maxBubble int) []PlaceBubble {
	bubbles := make([]PlaceBubble, 0)
	for i := 0; i < p.Len() && i < maxBubble; i++ {
		placePtr := &(*p)[i]
		bubbles = append(bubbles, (*FavoritePlace)(placePtr))
	}
	return bubbles
}

// 代替テキスト
func (p *NearbyPlaces) AltText() string {
	return "検索結果"
}

// 代替テキスト
func (p *FavoritePlaces) AltText() string {
	return "お気に入りリスト"
}

func (p *NearbyPlaces) Len() int {
	return len(*p)
}

func (p *FavoritePlaces) Len() int {
	return len(*p)
}

// レートを表す5つ星
func RatingStars(rating float64) []linebot.FlexComponent {
	maxRating := 5
	stars := make([]linebot.FlexComponent, maxRating)
	for i := 0; i < maxRating; i++ {
		star := linebot.IconComponent{
			Type: linebot.FlexComponentTypeIcon,
			URL:  StarIconURI(i < int(rating)),
			Size: linebot.FlexIconSizeTypeMd,
		}
		stars[i] = &star
	}
	stars = append(stars, &linebot.TextComponent{
		Type:   linebot.FlexComponentTypeText,
		Text:   fmt.Sprintf("%1.1f", rating),
		Margin: linebot.FlexComponentMarginTypeMd,
		Size:   linebot.FlexTextSizeTypeMd,
		Weight: linebot.FlexTextWeightTypeRegular,
		Color:  "#999999",
	})
	return stars
}

// 星アイコンのURI
func StarIconURI(gold bool) string {
	base := "https://scdn.line-apps.com/n/channel_devcenter/img/fx/"
	if gold {
		return base + "review_gold_star_28.png"
	}
	return base + "review_gray_star_28.png"
}
