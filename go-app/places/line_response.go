package places

import (
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

// LocationSendButton template
func LocationSendButton() *linebot.TemplateMessage {
	URIAction := linebot.NewURIAction("送信する", "line://nv/location")
	button := linebot.NewButtonsTemplate("", "", "位置情報を送信してネ", URIAction)
	return linebot.NewTemplateMessage("位置情報送信ボタン", button)
}

// MarshalMessage build a FlexMessage
func (places Places) MarshalMessage(maxBubble int) *linebot.FlexMessage {
	carousel := places.MarshalCarousel(maxBubble)
	return linebot.NewFlexMessage("検索結果", &carousel)
}

// MarshalCarousel build a CarouselContainer
func (places Places) MarshalCarousel(maxBubble int) linebot.CarouselContainer {
	bubbleContainers := make([]*linebot.BubbleContainer, 0)
	for i := 0; i < len(places) && i < maxBubble; i++ {
		bubble := places[i].MarshalBubble()
		bubbleContainers = append(bubbleContainers, &bubble)
	}
	carousel := linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbleContainers,
	}
	return carousel
}

// MarshalBubble builds a BubbleContainer
func (place *Place) MarshalBubble() linebot.BubbleContainer {
	return linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Size: linebot.FlexBubbleSizeTypeKilo,
		Hero: &linebot.ImageComponent{
			Type: linebot.FlexComponentTypeImage,
			URL: place.PhotoURI(map[string]string{
				"key":      os.Getenv("GCP_PLACES_API_KEY"),
				"maxwidth": "350",
			}),
			Size:       linebot.FlexImageSizeTypeFull,
			AspectMode: linebot.FlexImageAspectModeTypeCover,
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   place.Name,
					Size:   linebot.FlexTextSizeTypeMd,
					Weight: linebot.FlexTextWeightTypeBold,
					Wrap:   true,
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeHorizontal,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Action: linebot.NewURIAction("マップで見る", place.GoogleMapURI()),
				},
			},
		},
	}
}
