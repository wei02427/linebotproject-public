package parkingstruct

import "encoding/json"

//cellTPE 單一停車狀態
type cellTPE struct {
	CellStatus string `json:"cellStatus"` //停車格位狀態(1：車格有車輛停放；2：車格無車輛停放；3：無訊息)
	CoordX     string `json:"coord_X"`    //停車格位X座標
	CoordY     string `json:"coord_Y"`    //停車格位Y座標
	DataDt     string `json:"data_Dt"`    //?
	PsID       string `json:"psId"`       //停車格位格號
}

//cellList 停車格清單
type cellList struct {
	Cells []*cellTPE `json:"cell"`
}

//road 路段停車格
type roadTPE struct {
	RoadSegAvail    string          `json:"roadSegAvail"`                 //路段剩餘格位數
	RoadSegFee      string          `json:"roadSegFee"`                   //收費標準
	RoadSegID       string          `json:"roadSegID"`                    //路段ID
	RoadSegName     string          `json:"roadSegName"`                  //路段名稱
	RoadSegTmEnd    string          `json:"roadSegTmEnd"`                 //收費結束時間
	RoadSegTmStart  string          `json:"roadSegTmStart"`               //收費開始時間
	RoadSegTotal    string          `json:"roadSegTotal"`                 //路段總格位數
	RoadSegUpdateTm string          `json:"roadSegUpdateTm"`              //資料更新時間
	RoadSegUsage    string          `json:"roadSegUsage"`                 //路段使用率
	CellStatusList  json.RawMessage `json:"cellStatusList" datastore:"-"` //單一停車格資訊
}

//TPE 台北市車格
type TPE struct {
	Roads []*roadTPE `json:"ROAD"`
}

type cellNTPC struct {
	ID            int     `json:"ID,string"`            //車格序號
	Name          string  `json:"NAME"`                 //車格類型
	Day           string  `json:"DAY"`                  //收費天
	Hour          string  `json:"Hour"`                 //收費時段
	Pay           string  `json:"PAY"`                  //收費形式
	PayCash       string  `json:"PAYCASH"`              //費率
	Memo          string  `json:"MEMO"`                 //車格備註
	RoadID        string  `json:"ROADID"`               //路段代碼
	CellStatus    bool    `json:"CELLSTATUS,string"`    //車格狀態判斷 Y有車 N空位
	IsNowCash     bool    `json:"ISNOWCASH,string"`     //收費時段判斷
	ParkingStatus int     `json:"ParkingStatus,string"` //車格狀態 　1：有車、2：空位、3：非收費時段、4：時段性禁停、5：施工（民眾申請施工租用車格時使用）
	Lat           float64 `json:"lat,string"`           //緯度
	Lon           float64 `json:"lon,string"`           //經度
}

//NTPC 新北市車格
type NTPC struct {
	Cells []*cellNTPC
}
