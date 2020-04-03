## 目錄結構

    .
    ├── 依功能命名資料夾
    |   ├── test
    |       └── main.go
    |   ├── function.go
    |   └── go.mod
    └── .gitignore

* function.go -> 撰寫cloud function
* go.mod -> function.go依賴
* main.go -> 測試function.go

## 撰寫規則

1. 依開發功能建立資料夾

2. 建立go.mod

    ```bash
    go mod init projet.com/功能名稱
    ```

3. 撰寫完 function.go，以 main.go 測試

4. main.go 引入 function.go 方式

    ```go
    import projet.com/功能名稱
    ```

## 部署到 cloud function

1. commit 至 github

2. cmd 執行
    ```bash
    gcloud functions deploy 執行函數名稱 --source https://source.developers.google.com/projects/parkingproject-261207/repos/github_wei02427_linebotproject/moveable-aliases/master/paths/資料夾名稱 --runtime=go113 --trigger-http --allow-unauthenticated
    ```