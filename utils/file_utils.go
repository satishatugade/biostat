package utils

import (
	"io"

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
		RecordExt: fileType,
		FileData: fileBytes,
	}

	return record, nil
}
