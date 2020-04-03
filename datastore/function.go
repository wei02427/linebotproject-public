package datastore

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	xml2json "github.com/basgys/goxml2json"
	parking "project.com/datastore/parkingstruct"
)

const projectID string = "parkingproject-261207"

//PubSubMessage gcp pub/sub payload
type PubSubMessage struct {
	Data []byte `json:"data"`
}

//UpdateParkingInfo consumes a Pub/Sub message 並更新停車位資訊
func UpdateParkingInfo(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data))
	//取open data

	//fileURL := "https://tcgbusfs.blob.core.windows.net/blobtcmsv/TCMSV_roadquery.gz"
	//TPEParkingInfo, err := getParkingInfo(fileURL)
	NTPCParkingInfo, err := getParkingInfo("https://data.ntpc.gov.tw/od/data/api/1A71BA9C-EF21-4524-B882-6D085DF2877A?$format=json")
	//fmt.Printf(*TPEParkingInfo)
	if err != nil {
		log.Print(err)
	}

	var NTPC parking.NTPC
	roadKeys := []*datastore.Key{}

	//json轉struct
	if err := json.Unmarshal([]byte(*NTPCParkingInfo), &NTPC.Cells); err != nil {
		log.Fatalf("error: %v", err)
	}

	//以roadID產生entity key
	for _, cell := range NTPC.Cells {

		parentKey := datastore.NameKey("NTPCRoadName", cell.RoadID, nil)
		roadKey := datastore.NameKey("NTPCParkings", strconv.Itoa(cell.ID), parentKey)
		roadKeys = append(roadKeys, roadKey)

	}
	//log.Print(&NTPC)
	putParkingInfo(ctx, roadKeys, &NTPC)

	return nil
}

//put路段資訊
func putParkingInfo(ctx context.Context, roadKeys []*datastore.Key, parkings interface{}) {

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	//fmt.Print(parkings.(*parking.NTPC))

	var tmp = 0
	n := math.Ceil(float64(len(roadKeys)) / 500) //一次最多put500筆

	for i := 1; i <= int(n); i++ {
		var size int
		if size = len(roadKeys); i*500 < len(roadKeys) {
			size = i * 500
		}

		switch parkings.(type) {
		case *parking.TPE:
			if _, err := client.PutMulti(ctx, roadKeys[tmp:size-1], parkings.(*parking.TPE).Roads[tmp:size-1]); err != nil {
				log.Fatalf("PutMulti TPE: %v", err)
			}
		case *parking.NTPC:
			//fmt.Print("----------", i, parkings.(*parking.NTPC).Cells[tmp:size-1])
			if _, err := client.PutMulti(ctx, roadKeys[tmp:size-1], parkings.(*parking.NTPC).Cells[tmp:size-1]); err != nil {
				log.Fatalf("PutMulti NTPC: %v", err)
			}
		}

		tmp = size - 1
	}

	fmt.Printf("Info Saved sucess")

}

//GetParkingInfo 取得停車格資訊(TPE)
func getParkingInfo(url string) (*string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var data string
	if strings.Contains(url, ".gz") {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		body, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		var xml = strings.NewReader(string(body))
		json, err := xml2json.Convert(xml)
		if err != nil {
			log.Print("Failed to convert xml to json")
		}

		data = json.String()
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		data = string(body)
		data = strings.ReplaceAll(data, "\"CELLSTATUS\""+":"+"\"Y\"", "\"CELLSTATUS\""+":"+"\"true\"")
		data = strings.ReplaceAll(data, "\"CELLSTATUS\""+":"+"\"N\"", "\"CELLSTATUS\""+":"+"\"false\"")
		fmt.Print()
	}

	defer resp.Body.Close()
	return &data, nil
}

//gcloud functions deploy PutParkingInfo --source https://source.developers.google.com/projects/parkingproject-261207/repos/github_wei02427_linebotproject/moveable-aliases/master/paths/datastore --runtime=go113 --trigger-topic=updateInfo

/*單一車格資訊，因缺少座標故先不用*/

// var cells CellList
// if err := json.Unmarshal(road.CellStatusList, &cells); err == nil {
// 	cellKeys := []*datastore.Key{}
// 	if len(cells.Cells) != 0 {
// 		for i := 0; i < len(cells.Cells); i++ {
// 			cellKey := datastore.IncompleteKey("Cells", roadKey)
// 			cellKeys = append(cellKeys, cellKey)
// 		}

// 		if _, err := client.PutMulti(ctx, cellKeys, cells.Cells); err != nil {
// 			log.Fatalf("PutMulti: %v", err)
// 		} else {
// 			fmt.Printf("%s cells Saved sucess", road.RoadSegName)
// 		}
// 	}
// }

// type ID2Name struct {
// 	RoadID   string
// 	RoadName string
// }
// type TTT struct {
// 	IDs []*ID2Name
// }

// //a
// func A(ctx context.Context) error {

// 	var datas TTT
// 	roadKeys := []*datastore.Key{}

// 	file, err := os.Open("data.txt")

// 	if err != nil {
// 		log.Fatalf("failed opening file: %s", err)
// 	}

// 	scanner := bufio.NewScanner(file)
// 	scanner.Split(bufio.ScanLines)
// 	var txtlines []string

// 	for scanner.Scan() {
// 		txtlines = append(txtlines, scanner.Text())
// 	}

// 	file.Close()

// 	for _, eachline := range txtlines {
// 		dataSlice := strings.Split(eachline, " ")
// 		var a ID2Name
// 		a.RoadID = dataSlice[1]
// 		a.RoadName = dataSlice[4]
// 		// log.Print(a.RoadID, a.RoadName)
// 		datas.IDs = append(datas.IDs, &a)
// 	}

// 	//以roadID產生entity key
// 	for _, ID := range datas.IDs {
// 		roadKey := datastore.NameKey("NTPCRoadName", ID.RoadID, nil)
// 		roadKeys = append(roadKeys, roadKey)

// 	}
// 	B(ctx, roadKeys, &datas)
// 	return nil
// }

// //put路段資訊
// func B(ctx context.Context, roadKeys []*datastore.Key, datas interface{}) {
// 	// fmt.Print("@@@", datas)
// 	client, err := datastore.NewClient(ctx, projectID)
// 	if err != nil {
// 		log.Fatalf("Failed to create client: %v", err)
// 	}
// 	if _, err := client.PutMulti(ctx, roadKeys[0:], datas.(*TTT).IDs[0:]); err != nil {
// 		log.Fatalf("PutMulti ID: %v", err)
// 	}
// 	fmt.Printf("Info Saved sucess")

// }
