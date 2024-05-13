package controller

import (
	"encoding/json"
	"fmt"
	"gocroot/config"
	"gocroot/helper"
	"gocroot/model"
	"log"
	"net/http"
)

func HandleRequest(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	var msg model.IteungMessage
	if GetSecretFromHeader(req) == config.WebhookSecret {
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			resp.Response = err.Error()
		}
		resp = helper.WebHook(config.WAKeyword, config.WAPhoneNumber, config.WAAPIQRLogin, config.WAAPIMessage, msg, config.Mongoconn)
	} else {
		dt := &model.WebHook{
			URL:    config.WebhookURL,
			Secret: config.WebhookSecret,
		}
		res, err := helper.RefreshToken(dt, config.WAPhoneNumber, config.WAAPIGetToken, config.Mongoconn)
		if err != nil {
			resp.Response = err.Error()
		}
		resp.Response = jsonstr(res.ModifiedCount)
	}
	fmt.Fprintf(respw, resp.Response)
}

func jsonstr(strc interface{}) string {
	jsonData, err := json.Marshal(strc)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonData)
}

func GetSecretFromHeader(r *http.Request) (secret string) {
	if r.Header.Get("secret") != "" {
		secret = r.Header.Get("secret")
	} else if r.Header.Get("Secret") != "" {
		secret = r.Header.Get("Secret")
	}
	return
}