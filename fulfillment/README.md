# Fulfillment 撰寫

## Webhook要求

>*如果啟用執行要求的意圖比對相符，Dialogflow 就會使用 JSON 物件向 Webhook 提出 HTTP POST 要求，其中包含有關相符意圖的資訊。*

* JSON 要求格式

    ```json
    {
        "responseId": "...",
        "session": "projects/your-agents-project-id/agent/sessions/88d13aa8-2999-4f71-b233-39cbf3a824a0",
        "queryResult": {
            "queryText": "使用者原始查詢語句",
            "parameters": {
            "參數名": "參數值"
            },
            "allRequiredParamsPresent": true,
            "fulfillmentText": "Text defined in Dialogflow's console for the intent that was matched",
            "fulfillmentMessages": [
            {
                "text": {
                "text": [
                    "Text defined in Dialogflow's console for the intent that was matched"
                ]
                }
            }
            ],
            "outputContexts": [
            {
                "name": "projects/your-agents-project-id/agent/sessions/88d13aa8-2999-4f71-b233-39cbf3a824a0/contexts/generic",
                "lifespanCount": 5,
                "parameters": {
                "param": "param value"
                }
            }
            ],
            "intent": {
            "name": "projects/your-agents-project-id/agent/intents/29bcd7f8-f717-4261-a8fd-2d3e451b8af8",
            "displayName": "Matched Intent Name"
            },
            "intentDetectionConfidence": 1,
            "diagnosticInfo": {},
            "languageCode": "en"
        },
        "originalDetectIntentRequest": {}
        }
    ```

* 取得 JSON 值

    >*使用第三方套件 gjson*

    ```go
    const json = `{"name":{"first":"Janet","last":"Prichard"},"age":47}`

    func main() {
        value := gjson.Get(json, "name.last")
        //value = Prichard
    }
    ```

    [Read More](https://github.com/tidwall/gjson)

## Webhook回應

* JSON 回應格式

    ```json
    {
      "fulfillmentText": "This is a text response",
      "fulfillmentMessages": [
        {
          "card": {
            "title": "card title",
            "subtitle": "card text",
            "imageUri": "https://assistant.google.com/static/images/molecule/Molecule-Formation-stop.png",
            "buttons": [
              {
                "text": "button text",
                "postback": "https://assistant.google.com/"
              }
            ]
          }
        }
      ],
      "source": "example.com",
      "payload": {
        "google": {
          "expectUserResponse": true,
          "richResponse": {
            "items": [
              {
                "simpleResponse": {
                  "textToSpeech": "this is a simple response"
                }
              }
            ]
          }
        },
        "facebook": {
          "text": "Hello, Facebook!"
        },
        "slack": {
          "text": "This is a text response for Slack."
        }
      },
      "outputContexts": [
        {
          "name": "projects/${PROJECT_ID}/agent/sessions/${SESSION_ID}/contexts/context name",
          "lifespanCount": 5,
          "parameters": {
            "param": "param value"
          }
        }
      ],
      "followupEventInput": {
        "name": "event name",
        "languageCode": "en-US",
        "parameters": {
          "param": "param value"
        }
      }
    }
    ```

## 本地除錯

>*先將 ngrok.exe 放到 fulfillment/*

1. 執行 main.go

2. 將 localhost 部署到 ngrok

    ``` bash
    ./ngrok http 3000
    ```

3. 將連結貼至 dialogflow Webhook

4. 利用 DEBUG CONSOLE 除錯
