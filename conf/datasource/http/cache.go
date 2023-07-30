package http

import (
	"fmt"
	"io/ioutil"
	"os"

	"miopkg/log"
	"miopkg/util/xfile"
)

// GetFileName ...
func GetFileName(cacheKey string, cacheDir string) string {
	return cacheDir + string(os.PathSeparator) + cacheKey
}

// WriteConfigToFile ...
func WriteConfigToFile(cacheKey string, cacheDir string, content string) {
	if err := xfile.MkdirIfNecessary(cacheDir); err != nil {
		log.Errorf("[ERROR]:faild to MkdirIfNecessary config ,value:%s ,err:%s \n", string(content), err.Error())
		return
	}
	fileName := GetFileName(cacheKey, cacheDir)
	err := ioutil.WriteFile(fileName, []byte(content), 0666)
	if err != nil {
		log.Errorf("[ERROR]:faild to write config  cache:%s ,value:%s ,err:%s \n", fileName, string(content), err.Error())
	}
}

// ReadConfigFromFile ...
func ReadConfigFromFile(cacheKey string, cacheDir string) (string, error) {
	fileName := GetFileName(cacheKey, cacheDir)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to read config cache file:%s,err:%s! ", fileName, err.Error())
	}
	return string(b), nil
}
