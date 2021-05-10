package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type OutfitManager struct {
	AuthenticationCookie string // Used to wear outfits & retrieve CSRF-TOKENs
}

// TO DO:
// Warn user about 429s and other import server statusCodes

// JSON parsed endpoint response
type UserOutfits struct {
	FilteredCount int `json:"filteredCount"`

	Data []struct {
		Id         int
		Name       string
		IsEditable bool
	} `json:"data"`
}

// Struct representing an outfit
type UserOutfit struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	Assets []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`

		AssetType struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"assetType"`
	} `json:"assets"`

	BodyColors struct {
		HeadColorId     int
		TorsoColorId    int
		RightArmColorId int
		LeftArmColorId  int
		RightLegColorId int
		LeftLegColorId  int
	}
}

// Get outfit IDs for Roblox user
func (manager OutfitManager) getUserOutfitIds(userId int) UserOutfits {
	response, err := http.Get(
		fmt.Sprintf("https://avatar.roblox.com/v1/users/%v/outfits?page=1&itemsPerPage=100", userId),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println(err.Error())
	}

	outfits := UserOutfits{}

	json.Unmarshal(body, &outfits)

	return outfits
}

// Download Roblox outfits from endpoint response
func (manager OutfitManager) getUserOutfits(userOutfits UserOutfits) UserOutfit {
	outfit := userOutfits.Data[0]

	response, err := http.Get(fmt.Sprintf("https://avatar.roblox.com/v1/outfits/%v/details", outfit.Id))

	if err != nil {
		fmt.Println(err.Error())
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println(err.Error())
	}

	userOutfit := UserOutfit{}

	err = json.Unmarshal(body, &userOutfit)

	if err != nil {
		fmt.Println(err.Error())
	}

	return userOutfit

	/* 	for _, outfit := range userOutfits.Data {
		// TO DO:
		// Create JSON file:
		//	Remove spaces from outfit name and replace with _
		// 	Write data to JSON file

		fmt.Println(outfit.Name)
	} */
}

// Save outfit (to JSON file)
func (manager OutfitManager) saveOutfit(outfit UserOutfit) {
	if !folderExists("outfits") {
		err := os.Mkdir("outfits", os.ModePerm)

		if err != nil {
			fmt.Println(err.Error())
		}
	}

	outfitEncoded, err := json.Marshal(&outfit)

	if err != nil {
		fmt.Println(err.Error())
	}

	name := strings.ReplaceAll(outfit.Name, " ", "_")

	err = ioutil.WriteFile(fmt.Sprintf("outfits/%v.json", name), outfitEncoded, os.ModePerm)

	if err != nil {
		fmt.Println(err.Error())
	}
}

// Load outfit (from JSON file)
func (manager OutfitManager) loadOutfit() {}

// Wear outfit
func (manager OutfitManager) wearOutfit(outfit UserOutfit) {
	if manager.AuthenticationCookie == "" {
		fmt.Println("wearOutfit requires a valid authentication cookie!")
		return
	}

	type requestAssetIds struct {
		AssetIds []int
	}

	requestData := requestAssetIds{
		AssetIds: make([]int, len(outfit.Assets)),
	}

	for i := 0; i < len(requestData.AssetIds); i++ {
		requestData.AssetIds[i] = outfit.Assets[i].Id
	}

	requestBody, err := json.Marshal(requestData)

	if err != nil {
		fmt.Println(err.Error())
	}

	request, err := http.NewRequest("POST", "https://avatar.roblox.com/v1/avatar/set-wearing-assets", bytes.NewBuffer(requestBody))

	if err != nil {
		fmt.Println(err.Error())
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Csrf-Token", manager.getCSRFToken())

	request.AddCookie(&http.Cookie{
		Name:  ".ROBLOSECURITY",
		Value: manager.AuthenticationCookie,
	})

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)

	fmt.Println("Status code:", response.StatusCode)
	fmt.Println(string(body))
}

func (manager *OutfitManager) getCSRFToken() string {
	if manager.AuthenticationCookie == "" {
		fmt.Println("Cannot retrieve CSRF Token. No cookie provided.")
		return ""
	}

	request, err := http.NewRequest("POST", "https://auth.roblox.com/v2/logout", nil)

	if err != nil {
		fmt.Println(err.Error())
	}

	request.AddCookie(&http.Cookie{
		Name:  ".ROBLOSECURITY",
		Value: manager.AuthenticationCookie,
	})

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer response.Body.Close()

	if err != nil {
		fmt.Println(err.Error())
	}

	switch response.StatusCode {
	case 401:
		fmt.Println("Invalid cookie!")
		manager.AuthenticationCookie = ""
	case 403:
		return response.Header.Get("X-Csrf-Token")
	default:
		fmt.Println("Unexpected status code: ", response.StatusCode)
	}

	return ""
}
