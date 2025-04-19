package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func GetDigiLockerToken(code string) (map[string]interface{}, error) {
	apiUrl := "https://digilocker.meripehchaan.gov.in/public/oauth2/1/token"

	data := url.Values{}
	data.Set("grant_type", os.Getenv("DIGILOCKER_GRANT_TYPE"))
	data.Set("code", code)
	data.Set("redirect_uri", os.Getenv("DIGILOCKER_REDIRECT_URI"))
	data.Set("code_verifier", os.Getenv("DIGILOCKER_CODE_VERIFIER"))
	data.Set("client_id", os.Getenv("DIGILOCKER_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("DIGILOCKER_CLIENT_SECRET"))

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch token: " + result["error_description"].(string))
	}

	return result, nil
}

func GetDigiLockerDirs(accessToken string) (map[string]interface{}, error) {
	apiUrl := "https://digilocker.meripehchaan.gov.in/public/oauth2/1/files/"

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch dirs: " + result["error_description"].(string))
	}

	return result, nil
}

func GetDigiLockerDocumentsList(accessToken string, dir_code string) (map[string]interface{}, error) {
	apiUrl := "https://digilocker.meripehchaan.gov.in/public/oauth2/1/files/" + dir_code

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch documents: " + result["error_description"].(string))
	}

	return result, nil
}

func FetchDirItemsRecursively(token string, dirId string) {
	dirRes, err := GetDigiLockerDocumentsList(token, dirId)
	if err != nil {
		log.Println("Error fetching documents:", err)
		return
	}

	dirItems, ok := dirRes["items"].([]interface{})
	if !ok {
		log.Println("Items is not a list")
		return
	}

	for _, dirItem := range dirItems {
		record, ok := dirItem.(map[string]interface{})
		if !ok {
			continue
		}

		switch record["type"] {
		case "file":
			log.Printf("Save:: Name: %v, Size: %v Type: %v Uri: %v mime: %v\n",
				record["name"], record["size"], record["type"], record["uri"], record["mime"])
		case "dir":
			log.Printf("📂 Entering sub-directory: %v", record["name"])
			// Recursive call
			FetchDirItemsRecursively(token, record["id"].(string))
		}
	}
}
