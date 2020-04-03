package main

import (
	"context"

	linebotProject "project.com/datastore"
)

//main 測試function執行
func main() {
	//linebotProject.A(context.Background())
	linebotProject.UpdateParkingInfo(context.Background(), linebotProject.PubSubMessage{Data: []byte("update")})
}

//部署指令
//gcloud functions deploy PutParkingInfo --source https://source.developers.google.com/projects/parkingproject-261207/repos/github_wei02427_linebotproject/moveable-aliases/master/paths/datastore --runtime=go113 --trigger-topic=updateInfo
