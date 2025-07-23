package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"regexp"
	"strings"

	"biostat/models"

	"github.com/gin-gonic/gin"
)

func ProcessFileUpload(ctx *gin.Context) (*models.TblMedicalRecord, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fileName := header.Filename
	fileSize := header.Size
	fileType := header.Header.Get("Content-Type")

	record := &models.TblMedicalRecord{
		RecordName: fileName,
		RecordSize: fileSize,
		FileType:   fileType,
		FileData:   fileBytes,
	}

	return record, nil
}

func ReadFileBytes(file multipart.File) ([]byte, error) {
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return fileData, nil
}

func GenerateHMAC(fileData []byte, clientSecret string) string {
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write(fileData)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func SanitizeFileName(name string) string {
	decodedName, err := url.QueryUnescape(name)
	if err != nil {
		decodedName = name
	}
	decodedName = strings.ReplaceAll(decodedName, " ", "_")
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	return re.ReplaceAllString(decodedName, "_")
}

func SaveFile(header *multipart.FileHeader, path string) error {
	return os.WriteFile(path, func() []byte {
		file, _ := header.Open()
		defer file.Close()
		data, _ := io.ReadAll(file)
		return data
	}(), 0644)
}
