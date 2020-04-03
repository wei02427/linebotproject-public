package carouselmessage

import (
	"fmt"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"

)

func bubbleContainer(roadName string, lat float64, lon float64, avail int,dis string, num int) (container *linebot.BubbleContainer) {
	container = &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Header: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   "No. " + strconv.Itoa(num),
					Size:   linebot.FlexTextSizeTypeXxl,
					Weight: linebot.FlexTextWeightTypeBold,
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: roadName,
					Size: linebot.FlexTextSizeTypeXl,
				},
			},
		},
		Styles: &linebot.BubbleStyle{
			Header: &linebot.BlockStyle{
				BackgroundColor: "#FF6B6E",
			},
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: "距離" + dis,
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: "剩餘 " + strconv.Itoa(avail) + " 個",
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:  linebot.FlexComponentTypeButton,
					Style: linebot.FlexButtonStyleTypeLink,
					Action: &linebot.URIAction{
						Label: "導航",
						URI:   "https://www.google.com/maps/search/?api=1&query=" + fmt.Sprintf("%f", lat) + "," + fmt.Sprintf("%f", lon),
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeNano,
	}

	return
}

//Carouselmesage 產生訊息
func Carouselmesage(roads [5][5]interface{}) (container *linebot.CarouselContainer) {
	var bubbleConts []*linebot.BubbleContainer

	for i,info := range roads {
		bubbleConts = append(bubbleConts, bubbleContainer(info[0].(string), info[1].(float64), info[2].(float64), info[3].(int), info[4].(string),i+1))
	}
	container = &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbleConts,
	}
	return
}
