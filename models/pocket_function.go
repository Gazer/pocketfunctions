package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type PocketFunction struct {
	Uri  string `json:"path"`
	Code string `json:"code"`
	Id   string
}

func PocketFunctionFromRequest(c *gin.Context) *PocketFunction {
	var newFunction PocketFunction
	newFunction.Uri = c.Param("path")
	newFunction.Id = GetMD5Hash(newFunction.Uri)
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil
	}
	newFunction.Code = strings.TrimRight(fileHeader.Filename, ".zip")

	// Open or create the file to save the uploaded content
	dst, err := os.Create(fmt.Sprintf("function_repository/%s.zip", newFunction.Code))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}
	return &newFunction
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
