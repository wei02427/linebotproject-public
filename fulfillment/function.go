package fulfillment

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"

	"cloud.google.com/go/datastore"
	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/thedevsaddam/gojsonq"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"project.com/fulfillment/carouselmessage"
)

//road 路段停車格
type parking struct {
	ID            int     //車格序號
	Name          string  //車格類型
	Day           string  //收費天
	Hour          string  //收費時段
	Pay           string  //收費形式
	PayCash       string  //費率
	Memo          string  //車格備註
	RoadID        string  //路段代碼
	CellStatus    bool    //車格狀態判斷 Y有車 N空位
	IsNowCash     bool    //收費時段判斷
	ParkingStatus int     //車格狀態 　1：有車、2：空位、3：非收費時段、4：時段性禁停、5：施工（民眾申請施工租用車格時使用）
	Lat           float64 //緯度
	Lon           float64 //經度
	Distance      string  //距離
}

// dialogflowProcessor has all the information for connecting with Dialogflow
type dialogflowProcessor struct {
	projectID        string
	authJSONFilePath string
	lang             string
	timeZone         string
	sessionClient    *dialogflow.SessionsClient
	ctx              context.Context
}

// datastoreProcessor 存取 datastore
type datastoreProcessor struct {
	projectID string
	client    *datastore.Client
	ctx       context.Context
}

// nlpResponse is webhook回應
type nlpResponse struct {
	Intent     string
	Confidence float32
	Entities   map[string]string
	Prompts    string
}

const projectID string = "parkingproject-261207"

var dialogflowProc dialogflowProcessor
var datastoreProc datastoreProcessor
var bot *linebot.Client

var err error

//response webhook回應
type response struct {
	FulfillmentText string `json:"fulfillmentText"`
}

// Pair A data structure to hold a key/value pair.
type Pair struct {
	Key   string
	Value float64
}

// PairList A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// sortMapByValue 排序 map
func sortMapByValue(m map[string][]float64) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = Pair{k, v[0]} //k=roadID v=distance
		i++
	}
	sort.Sort(p) //開始排序 詳見:https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter03/03.1.html
	return p
}

// init 初始化權限
func init() {
	bot, err = linebot.New("57cc60c3fc1530cc32ba896e1c4b7856", "GiKIwKk+Lwku0WeGEGnlEDBDDGC67tQVCSIMbcQaKpA2IyZPU6OgVSIdI0h1HUUG2Ky/psNLEEkjfnEZGITnJolxlEScGgLoWT/iKpwyinf/IJDgeB5gnIB0zmuag0vYlcs7WgOYdUg0CwbGXlWKIwdB04t89/1O/w1cDnyilFU=")
	dialogflowProc.init(projectID, "parkingproject-261207-2933e4112308.json", "zh-TW", "Asia/Hong_Kong")
	datastoreProc.init(projectID)

}

func replyUser(resp interface{}, event *linebot.Event) {
	switch resp.(type) { //確認是何種類型訊息
	case string:
		if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(resp.(string))).Do(); err != nil {
			log.Print(err)
		}
	case [5][5]interface{}:
		container := carouselmessage.Carouselmesage(resp.([5][5]interface{})) //建立Carouselmesage
		if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage("車位資訊。", container)).Do(); err != nil {
			log.Print(err)
		}
	}
	//var roads []map[string]string
	// roads = append(roads, map[string]string{"roadName": "五權路", "roadAvail": "10"})
	//linebot.NewFlexMessage("車位資訊。", container),

	//container := carouselmessage.Carouselmesage(roads)

}

//Fulfillment 查詢車位
func Fulfillment(w http.ResponseWriter, r *http.Request) {

	var events []*linebot.Event
	events, err = bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
			log.Print(err)
		} else {
			w.WriteHeader(500)
			log.Print(err)
		}

	} else {
		w.WriteHeader(200)
	}

	var resp interface{} //回傳的訊息，可能為text、Carouselmesage，故用interface

	//可能不只一位使用者傳送訊息
	for _, event := range events {
		//訊息事件 https://developers.line.biz/en/reference/messaging-api/#common-properties
		if event.Type == linebot.EventTypeMessage {
			//訊息種類
			switch message := event.Message.(type) {
			case *linebot.TextMessage: //文字訊息
				log.Println("UserID", event.Source.UserID)
				response := dialogflowProc.processNLP(message.Text, event.Source.UserID) //解析使用者所傳文字
				if response.Intent == "FindParking" {
					if _, ok := response.Entities["location"]; ok {
						// log.Printf("@@@@@@@@", response.Entities["location"])
						lat, lon := getGPS(response.Entities["location"]) //路名轉GPS
						resp = getData(lat, lon)                          //查詢車格資訊

					} else {
						resp = response.Prompts //如果偵測到intent卻沒有entity，回傳提示輸入訊息
					}
				} else {
					resp = "我聽不太懂"
				}
			case *linebot.LocationMessage: //位置訊息
				fmt.Printf("gps %f,%f\n", message.Latitude, message.Longitude)

				resp = getData(message.Latitude, message.Longitude) //查詢車格資訊
			}
			//加好友事件
		} else if event.Type == linebot.EventTypeFollow {
			resp = "還敢加我好友啊"
		}

		replyUser(resp, event) //回復使用者訊息
	}
}

//初始化 dialogflow (pointer receiver)
func (dp *dialogflowProcessor) init(data ...string) (err error) {
	dp.projectID = data[0]
	dp.authJSONFilePath = data[1]
	dp.lang = data[2]
	dp.timeZone = data[3]
	// Auth process: https://dialogflow.com/docs/reference/v2-auth-setup

	dp.ctx = context.Background()
	dp.sessionClient, err = dialogflow.NewSessionsClient(dp.ctx, option.WithCredentialsFile(dp.authJSONFilePath))

	return
}

func (ds *datastoreProcessor) init(data string) (err error) {
	ds.projectID = data
	ds.ctx = context.Background()
	ds.client, err = datastore.NewClient(ds.ctx, ds.projectID)
	return
}

//dialogflow 分析語意 (pointer receiver)
func (dp *dialogflowProcessor) processNLP(rawMessage string, username string) (r nlpResponse) {
	//DetectIntentRequest struct https://godoc.org/google.golang.org/genproto/googleapis/cloud/dialogflow/v2#StreamingDetectIntentRequest
	sessionID := username
	request := dialogflowpb.DetectIntentRequest{
		Session: fmt.Sprintf("projects/%s/agent/sessions/%s", dp.projectID, sessionID),
		QueryInput: &dialogflowpb.QueryInput{
			Input: &dialogflowpb.QueryInput_Text{
				Text: &dialogflowpb.TextInput{
					Text:         rawMessage,
					LanguageCode: dp.lang,
				},
			},
		},
		QueryParams: &dialogflowpb.QueryParameters{
			TimeZone: dp.timeZone,
		},
	}
	//DetectIntent https://godoc.org/cloud.google.com/go/dialogflow/apiv2#SessionsClient.DetectIntent
	response, err := dp.sessionClient.DetectIntent(dp.ctx, &request)
	if err != nil {
		log.Fatalf("Error in communication with Dialogflow %s", err.Error())
		return
	}
	queryResult := response.GetQueryResult()
	if queryResult.Intent != nil {
		//The name of this Intent
		r.Intent = queryResult.Intent.DisplayName
		//Values range from 0.0 (completely uncertain) to 1.0 (completely certain).
		// This value is for informational purpose only and is only used to
		// help match the best intent within the classification threshold.
		r.Confidence = float32(queryResult.IntentDetectionConfidence)

	}
	r.Entities = make(map[string]string)
	//The collection of extracted parameters.
	params := queryResult.Parameters.GetFields()
	if len(params) > 0 {
		for paramName, entity := range params {
			extractedValue := extractDialogflowEntities(entity) //解析entities type
			log.Printf("paramName= %s, entity= %s\n", paramName, extractedValue)
			if extractedValue != "" {
				r.Entities[paramName] = extractedValue
			} else {
				r.Prompts = queryResult.GetFulfillmentText() //因entity為必要參數，若為空則取得prompts提示文字
			}
		}
	}
	return
}

// func (ds *datastoreProcessor) processDB()
// 解碼 Protobuf 格式
func extractDialogflowEntities(p *structpb.Value) (extractedEntity string) {
	kind := p.GetKind()
	switch kind.(type) {
	case *structpb.Value_StringValue:
		return p.GetStringValue()
	case *structpb.Value_NumberValue:
		return strconv.FormatFloat(p.GetNumberValue(), 'f', 6, 64)
	case *structpb.Value_BoolValue:
		return strconv.FormatBool(p.GetBoolValue())
	case *structpb.Value_StructValue: //sys.location 為內建entity，回傳格式為struct，可以在dialogflow上輸入測試地址看完整結構
		s := p.GetStructValue()
		fields := s.GetFields()

		// for key, value := range fields {
		// 	log.Printf("key: %s, value: %s", key, value)
		// 	// @TODO: Other entity types can be added here
		// }
		extractedEntity := fields["street-address"].GetStringValue() //取得地址這欄
		return extractedEntity

	case *structpb.Value_ListValue:
		list := p.GetListValue()
		if len(list.GetValues()) > 1 {
			// @TODO: Extract more values
		}
		extractedEntity = extractDialogflowEntities(list.GetValues()[0])
		return extractedEntity
	default:
		return ""
	}
}

func floatToString(num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(num, 'f', 6, 64)
}

func getDist(userLat float64, userLon float64, lat float64, lon float64) (dist float64) {
	dist = math.Abs(userLat-lat) + math.Abs(userLon-lon)
	return
}

func getDistText(userLat float64, userLon float64, lat float64, lon float64) (distText string) {
	origins := floatToString(userLat) + "," + floatToString(userLon)
	destinations := floatToString(lat) + "," + floatToString(lon)
	// log.Printf("origins===",origins)
	// log.Printf("destinations===",destinations)
	url := "https://maps.googleapis.com/maps/api/distancematrix/json?origins=" + origins + "&destinations=" + destinations + "&key=AIzaSyAhsij-kCTyOzK9Vq83zemmxJXTdNJVkV8"
	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	jq := gojsonq.New().FromString(string(body)) //gojsonq解析json
	res := jq.Find("rows.[0].elements.[0].distance")
	dis := res.(map[string]interface{})
	distText = dis["text"].(string)
	// distValue = dis["value"].(float64)
	// log.Print("valueeeee=", distValue)
	return
}

func getGPS(roadName string) (lat float64, lon float64) {

	geocoding := "https://maps.googleapis.com/maps/api/geocode/json?address=" + roadName + "&key=AIzaSyAhsij-kCTyOzK9Vq83zemmxJXTdNJVkV8"
	resp, _ := http.Get(geocoding)
	body, _ := ioutil.ReadAll(resp.Body)
	jq := gojsonq.New().FromString(string(body))    //gojsonq解析json
	res := jq.Find("results.[0].geometry.location") //可以直接點網址了解json結構
	gps := res.(map[string]interface{})             //interface型態轉回map
	lat = gps["lat"].(float64)
	lon = gps["lng"].(float64)
	return
}

type roadName struct {
	RoadID   string
	RoadName string
}

func getRoadName(id string) (name string) {

	query := datastore.NewQuery("NTPCRoadName").
		Filter("RoadID =", id)
	it := datastoreProc.client.Run(datastoreProc.ctx, query)
	for {
		var data roadName
		_, err := it.Next(&data)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching next task: %v", err)
		}
		name = data.RoadName
	}
	return
}

// getData  找車位資料-`-`
func getData(lat float64, lon float64) (parkings [5][5]interface{}) {

	/*查詢各路段 ID*/
	// query := datastore.NewQuery("NTPCParkings").
	// 	Project("RoadID").
	// 	DistinctOn("RoadID").
	// 	Order("RoadID")
	// id := []string{}

	//datastore 查詢剩餘車位

	list := make(map[string][]float64) //儲存各路段離使用者最近且為空位的車格(一個路段一個) ex:[RoadID][distance,lat,lon,剩餘數量]

	//var parkings []parking
	for _, i := range []int{2, 3} { //2為空位,3為非收費時段,datastore查詢沒有or的方法，所以須查詢兩次
		query := datastore.NewQuery("NTPCParkings").
			Filter("CellStatus =", false). //false代表沒有車，但必須確認ParkingStatus必須為2或3才可停
			Filter("ParkingStatus =", i)

		it := datastoreProc.client.Run(datastoreProc.ctx, query)

		for {
			var parking parking
			_, err := it.Next(&parking) //查詢後的結果一一迭代儲存到車格的struct

			if err == iterator.Done {
				break
			} else if err != nil {
				log.Fatalf("Error fetching road: %v", err)
			}
			//fmt.Printf("RoadID %s\n", parking.RoadID)

			if dist1, ok := list[parking.RoadID]; ok { //確認車格是否已在list內，有則比較直線距離，無則直接儲存

				dist2 := getDist(lat, lon, parking.Lat, parking.Lon) //計算距離
				if dist2 < dist1[0] {                                //比較同路段車格距離，若距離較小，則復寫到list
					info := []float64{dist2, parking.Lat, parking.Lon, list[parking.RoadID][3] + 1}
					list[parking.RoadID] = info
				} else {
					list[parking.RoadID][3]++
				}

			} else {
				dist := getDist(lat, lon, parking.Lat, parking.Lon)
				info := []float64{dist, parking.Lat, parking.Lon, 1}
				list[parking.RoadID] = info

			}

			//parkings = append(parkings, parking)
			//id = append(id, road.RoadID)
		}
	}

	for i, v := range sortMapByValue(list)[:5] { //依照距離排序路段車格，並取前五
		text := getDistText(lat, lon, list[v.Key][1], list[v.Key][2])
		// fmt.Printf("%s %f,%f %d %s\n", ID2Name(v.Key), list[v.Key][1], list[v.Key][2], int(list[v.Key][3]), text)
		parkings[i] = [5]interface{}{getRoadName(v.Key), list[v.Key][1], list[v.Key][2], int(list[v.Key][3]), text} //儲存距離前五近車格，並回傳

	}

	return

	/*查詢各路段 ID*/
	// for _, i := range id {

	// 	query = datastore.NewQuery("NTPCParkings").
	// 		Filter("RoadID =", i).
	// 		Order("RoadID").
	// 		Limit(1)

	// 	it = datastoreProc.client.Run(datastoreProc.ctx, query)
	// 	for {
	// 		var road road
	// 		_, err := it.Next(&road)
	// 		if err == iterator.Done {
	// 			break
	// 		}
	// 		if err != nil {
	// 			log.Fatalf("Error fetching road: %v", err)
	// 		}

	/*geocoding gps 轉路名*/

	// 		fmt.Printf("RoadID %s ,%f ,%f ", road.RoadID, road.Lat, road.Lon)
	// 		geo := "https://maps.googleapis.com/maps/api/geocode/json?latlng=" + fmt.Sprintf("%f", road.Lat) + "," + fmt.Sprintf("%f", road.Lon) + "&result_type=route&language=zh-tw&key=AIzaSyAhsij-kCTyOzK9Vq83zemmxJXTdNJVkV8"
	// 		resp, _ := http.Get(geo)
	// 		body, _ := ioutil.ReadAll(resp.Body)
	// 		jq := gojsonq.New().FromString(string(body))
	// 		res := jq.From("results.[0].address_components").Where("types.[0]", "=", "route").Get()
	// 		fmt.Println(res.([]interface{})[0].(map[string]interface{})["long_name"].(string))

	// 	}

	// }

}
