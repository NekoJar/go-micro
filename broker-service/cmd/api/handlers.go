package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type AuthPayload struct{
	Email string `json:" email"`
	Password string `json:"password"`
}

type RequestPayload struct{
	Action string `json:"action"`
	Auth AuthPayload `json:"auth,omitempty"`
}

func (app * Config) Broker(w http.ResponseWriter, r *http.Request){
	payload := jsonRespose{
		Error: false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)

}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request){
	var requestPayload RequestPayload

	err := app.readJSON(w,r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action{
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	default:
		app.errorJSON(w, errors.New("unkown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload){
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	request, err := http.NewRequest("POST","http://authentication-service/authenticated", bytes.NewBuffer(jsonData))
	if err != nil{
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil{
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized{
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	}else if response.StatusCode != http.StatusAccepted{
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsonRespose

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error{
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonRespose
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)

}