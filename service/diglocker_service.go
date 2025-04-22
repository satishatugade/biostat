package service

import (
	"biostat/models"
	"biostat/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func GetDigiLockerToken(code string) (map[string]interface{}, error) {
	apiUrl := "https://digilocker.meripehchaan.gov.in/public/oauth2/1/token"

	data := url.Values{}
	data.Set("grant_type", os.Getenv("DIGILOCKER_GRANT_TYPE"))
	data.Set("code", code)
	data.Set("redirect_uri", os.Getenv("DIGILOCKER_REDIRECT_URI"))
	data.Set("code_verifier", os.Getenv("DIGILOCKER_CODE_VERIFIER"))
	data.Set("client_id", os.Getenv("DIGI_LOCKER_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("DIGITLOCKER_CLIENT_SECRET"))

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
		log.Println("Error response:", string(body))
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

func FetchDirItemsRecursively(token string, dirId string, digiLockerId string, userId uint64) ([]models.TblMedicalRecord, error) {
	var allDocs []models.TblMedicalRecord

	dirRes, err := GetDigiLockerDocumentsList(token, dirId)
	if err != nil {
		log.Println("Error fetching documents:", err)
		return nil, err
	}

	dirItems, ok := dirRes["items"].([]interface{})
	if !ok {
		return nil, errors.New("invalid response format for items")
	}

	for _, dirItem := range dirItems {
		record, ok := dirItem.(map[string]interface{})
		if !ok {
			continue
		}

		switch record["type"] {
		case "file":
			newRecord := models.TblMedicalRecord{
				RecordName:     record["name"].(string),
				RecordSize:     utils.ParseIntField(record["size"].(string)),
				FileType:       record["mime"].(string),
				UploadSource:   "DigiLocker",
				SourceAccount:  digiLockerId,
				RecordCategory: "Report",
				Description:    record["description"].(string),
				UploadedBy:     userId,
				RecordUrl:      "https://digilocker.meripehchaan.gov.in/public/oauth2/1/file/" + record["uri"].(string),
				FetchedAt:      time.Now(),
				CreatedAt:      utils.ParseDateField(record["date"]),
			}
			allDocs = append(allDocs, newRecord)
		case "dir":
			log.Printf("Entering sub-directory: %v", record["name"])
			subDocs, err := FetchDirItemsRecursively(token, record["id"].(string), digiLockerId, userId)
			if err != nil {
				log.Printf("Error fetching sub-directory %v: %v", record["name"], err)
				continue
			}
			allDocs = append(allDocs, subDocs...)
		}
	}
	return allDocs, nil
}

func SaveRecordToDigiLocker(accessToken string, fileData []byte, fileName string, contentType string) (*models.TblMedicalRecord, error) {
	clientSecret := os.Getenv("DIGITLOCKER_CLIENT_SECRET")
	hmac := utils.GenerateHMAC(fileData, clientSecret)
	apiUrl := "https://digilocker.meripehchaan.gov.in/public/oauth2/1/file/upload"
	req, err := http.NewRequest("POST", apiUrl, bytes.NewReader(fileData))
	if err != nil {
		return nil, errors.New("Erro sending:" + err.Error())
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("hmac", hmac)
	req.Header.Set("path", "Health/"+fileName)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Erro in response:" + err.Error())
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("Status: %s\n", resp.Status)
	log.Println("Response: %s\n", string(body))
	var uploadResult map[string]interface{}
	if err := json.Unmarshal(body, &uploadResult); err != nil {
		return nil, errors.New("Error while unmarshal:" + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("Error response:", string(body))
		return nil, errors.New(uploadResult["error_description"].(string))
	}

	fullPath, ok := uploadResult["path"].(string)
	if !ok {
		return nil, errors.New("invalid upload result format: path missing")
	}
	parts := strings.Split(fullPath, "/")
	uploadedFileName := parts[len(parts)-1]

	fileList, err := GetDigiLockerDocumentsList(accessToken, "SGVhbHRo")
	if err != nil {
		return nil, errors.New("error fetching updated directory: " + err.Error())
	}
	if items, ok := fileList["items"].([]interface{}); ok {
		for _, item := range items {
			file, ok := item.(map[string]interface{})
			if !ok || file["type"] != "file" {
				continue
			}
			if file["name"] == uploadedFileName {
				log.Println("Found uploaded file:", file)
				newRecord := models.TblMedicalRecord{
					RecordName:   file["name"].(string),
					RecordSize:   utils.ParseIntField(file["size"].(string)),
					FileType:     contentType,
					UploadSource: "DigiLocker",
					RecordUrl:    "https://digilocker.meripehchaan.gov.in/public/oauth2/1/file/" + file["uri"].(string),
					FetchedAt:    time.Now(),
					CreatedAt:    utils.ParseDateField(file["date"]),
				}
				return &newRecord, nil
			}
		}
	}

	return nil, errors.New("uploaded file not found in directory")
}

func ReadDigiLockerFile(accessToken string, fileUrl string) (*models.DigiLockerFile, error) {
	req, err := http.NewRequest("GET", fileUrl, nil)
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file, status: %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	hmacHeader := resp.Header.Get("hmac")

	body, _ := ioutil.ReadAll(resp.Body)

	return &models.DigiLockerFile{Data: body, ContentType: contentType, HMAC: hmacHeader}, nil
}
