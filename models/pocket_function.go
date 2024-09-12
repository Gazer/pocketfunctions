package models

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

type PocketFunction struct {
	Uri  string `json:"path"`
	Code string `json:"code"`
	Deps string `json:"deps"`
	Id   string
}

func PocketFunctionFromRequest(c *gin.Context) *PocketFunction {
	var newFunction PocketFunction
	if err := c.BindJSON(&newFunction); err != nil {
		return nil
	}
	newFunction.Id = GetMD5Hash(newFunction.Uri)
	return &newFunction
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
