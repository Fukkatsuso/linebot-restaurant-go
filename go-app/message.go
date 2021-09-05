package main

import (
	"encoding/json"
	"fmt"

	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/datastore"
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

// PostbackData is used in Postback
type PostbackData interface {
	PostbackData()
}

type Query datastore.Query

func (q *Query) PostbackData() {}

type PlaceInfo struct {
	PlaceID  string `json:"place_id"`
	PhotoURI string `json:"photo_uri"`
}

func (p *PlaceInfo) PostbackData() {}

// Postback is used to pustback
type Postback struct {
	Action PostbackAction `json:"action"`
	Data   PostbackData   `json:"data"`
}

// PostbackJSON converts PostbackData to json string
func PostbackJSON(action PostbackAction, pbData PostbackData) string {
	b, _ := json.Marshal(&Postback{
		Action: action,
		Data:   pbData,
	})
	return string(b)
}

// UnmarshalJSON override
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

// LocationSendButton message
func LocationSendButton() *linebot.TemplateMessage {
	URIAction := linebot.NewURIAction("送信する", "line://nv/location")
	button := linebot.NewButtonsTemplate("", "", "位置情報を送信してネ", URIAction)
	return linebot.NewTemplateMessage("位置情報送信ボタン", button)
}

// SearchConfirmWindow message
func SearchConfirmWindow(q *Query) *linebot.TemplateMessage {
	// jsonBytes, _ := json.Marshal(q)
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
	buttons := linebot.NewButtonsTemplate("", "絞り込みますか？", status(q), actions...)
	return linebot.NewTemplateMessage("確認ボタン", buttons)
}

func status(q *Query) string {
	var str string
	str += fmt.Sprintf("距離: %s\n", radiusMap[q.Radius])
	if len(q.Keywords) > 0 {
		str += fmt.Sprintf("キーワード: %v\n", q.Keywords)
	}
	return str
}

// RadiusQuickReply message
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

// Text message
func Text(text string) *linebot.TextMessage {
	return linebot.NewTextMessage(text)
}

// PlaceBubble is for conversion from Place to a bubble
type PlaceBubble interface {
	MarshalBubble() *linebot.BubbleContainer
}

// NearbyPlace implements PlaceBubble interface
type NearbyPlace places.Place

// FavoritePlace implements PlaceBubble interface
type FavoritePlace places.Place

// MarshalBubble is PlaceBubble interface
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

// MarshalBubble is PlaceBubble interface
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

// PlacesCarousel is for conversion from Places to a carousel message
type PlacesCarousel interface {
	PlaceBubbles(maxBubble int) []PlaceBubble
	AltText() string
	Len() int
}

// Carousel message
func Carousel(p PlacesCarousel, maxBubble int) *linebot.FlexMessage {
	carousel := MarshalCarousel(p, maxBubble)
	altText := p.AltText()
	return linebot.NewFlexMessage(altText, carousel)
}

// MarshalCarousel returns *linebot.CarouselContainer
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

// NearbyPlaces is Places
type NearbyPlaces places.Places

// FavoritePlaces is Places
type FavoritePlaces places.Places

// PlaceBubbles is PlacesCarousel interface
func (p *NearbyPlaces) PlaceBubbles(maxBubble int) []PlaceBubble {
	bubbles := make([]PlaceBubble, 0)
	for i := 0; i < p.Len() && i < maxBubble; i++ {
		placePtr := &(*p)[i]
		bubbles = append(bubbles, (*NearbyPlace)(placePtr))
	}
	return bubbles
}

// PlaceBubbles is PlacesCarousel interface
func (p *FavoritePlaces) PlaceBubbles(maxBubble int) []PlaceBubble {
	bubbles := make([]PlaceBubble, 0)
	for i := 0; i < p.Len() && i < maxBubble; i++ {
		placePtr := &(*p)[i]
		bubbles = append(bubbles, (*FavoritePlace)(placePtr))
	}
	return bubbles
}

// AltText is PlacesCarousel interface
func (p *NearbyPlaces) AltText() string {
	return "検索結果"
}

// AltText is PlacesCarousel interface
func (p *FavoritePlaces) AltText() string {
	return "お気に入りリスト"
}

// Len is PlacesCarousel interface
func (p *NearbyPlaces) Len() int {
	return len(*p)
}

// Len is PlacesCarousel interface
func (p *FavoritePlaces) Len() int {
	return len(*p)
}

// RatingStars returns star view
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

// StarIconURI is uri of star icon
func StarIconURI(gold bool) string {
	base := "https://scdn.line-apps.com/n/channel_devcenter/img/fx/"
	if gold {
		return base + "review_gold_star_28.png"
	}
	return base + "review_gray_star_28.png"
}
