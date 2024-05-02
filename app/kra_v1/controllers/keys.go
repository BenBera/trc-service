package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"gopkg.in/ini.v1"
)

func calculateSHA256(input string) string {
	// Convert the input string to a byte slice
	data := []byte(input)

	hash := sha256.New()

	hash.Write(data)

	hashSum := hash.Sum(nil)

	hashHex := hex.EncodeToString(hashSum)

	return hashHex
}



func GetKey(config *ini.File,section string, key string) (string,error)  {

	return  strings.TrimSpace(config.Section(section).Key(key).String()),nil

}


func GetKeyWithDefault(config *ini.File,section string, key string, defaults string) string  {

	v, e := GetKey(config,section,key)

	if e != nil {

		return  defaults
	}

	if len(v) == 0 {

		return v
	}

	return v;

}