package controllers

import (

	"fmt"
	"strings"

	"gopkg.in/ini.v1"
)


func GetKey(config *ini.File, section string, key string) (string, error) {
	sectionObj := config.Section(section)
	if sectionObj == nil {
	  return "", fmt.Errorf("section '%s' not found in configuration file", section)
	}
	value := sectionObj.Key(key).String()
	return strings.TrimSpace(value), nil
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