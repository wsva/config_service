package main

import (
	"errors"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

	wl_crypto "github.com/wsva/lib_go/crypto"
)

// config file for client, will be provided to clients
type ConfigFile struct {
	Filename string `json:"Filename"`
	Content  string `json:"Content"`
}

func GetConfigFileList(typeList []string, ip string) ([]ConfigFile, error) {
	indexfile := path.Join(DataDir, ConfigIndexFile)
	indexlist, err := LoadConfigIndexListFromFile(indexfile)
	if err != nil {
		return nil, err
	}
	var result []ConfigFile
	for _, v := range indexlist {
		//如果typeList为空，则默认为所有，而不是没有
		if len(typeList) > 0 && !slices.Contains(typeList, v.DirectoryOnServer) {
			continue
		}
		content, err := getConfigByType(v.DirectoryOnServer, ip)
		if err != nil {
			continue
		}
		result = append(result, ConfigFile{
			Filename: v.FilenameOnClient,
			Content:  content,
		})
	}
	return result, nil
}

/*
优先使用准确IP的，其次使用IP正则表达式的，最后使用public的
*/
func getConfigByType(configType, ip string) (string, error) {
	contentBytes, err := getConfigByExactIP(configType, ip)
	if err != nil {
		contentBytes, err = getConfigByIPRegexp(configType, ip)
		if err != nil {
			contentBytes, err = getConfigPublic(configType)
			if err != nil {
				return "", errors.New("config not found")
			}
		}
	}
	decryptedContent, err := wl_crypto.AES256Decrypt(AESKey, AESIV, string(contentBytes))
	if err != nil {
		return "", err
	}
	return decryptedContent, nil
}

func getConfigByExactIP(configType, ip string) ([]byte, error) {
	configFile := path.Join(DataDir, configType, ip)
	_, err := os.Stat(configFile)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(configFile)
}

func getConfigByIPRegexp(configType, ip string) ([]byte, error) {
	fileinfoList, err := os.ReadDir(path.Join(DataDir, configType))
	if err != nil {
		return nil, err
	}
	for _, v := range fileinfoList {
		if strings.Contains(v.Name(), KeywordRegexp) {
			exp := strings.ReplaceAll(v.Name(), KeywordRegexp, "")
			reg := regexp.MustCompile(exp)
			if reg.MatchString(ip) {
				configFile := path.Join(DataDir, configType, v.Name())
				return os.ReadFile(configFile)
			}
		}
	}
	return nil, errors.New("config not found by regexp")
}

func getConfigPublic(configType string) ([]byte, error) {
	configFile := path.Join(DataDir, configType, PublicConfigFile)
	return os.ReadFile(configFile)
}
