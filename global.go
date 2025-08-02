package main

import (
	"encoding/json"
	"os"
	"path"

	wl_fs "github.com/wsva/lib_go/fs"
	wl_http "github.com/wsva/lib_go/http"
	wl_location "github.com/wsva/lib_go/location"
	wl_int "github.com/wsva/lib_go_integration"
)

const (
	RoleRoot   = "Root"
	RoleBranch = "Branch"

	AESKey = "key"
	AESIV  = "iv"
)

/*
FormatSingleLine : it can be easier for shell to parse
*/
const (
	FormatSingleLine = "SingleLine"
)

const (
	KeywordRegexp    = "Regexp"
	PublicConfigFile = "public"
)

const (
	ConfigDataZipFile  = "config_service_data.zip"
	ConfigDataHashFile = "config_service_data.zip.sum"
)

// MainConfig comment
type MainConfig struct {
	ListenList []wl_http.ListenInfo   `json:"ListenList"`
	Role       string                 `json:"Role"`
	SourceList []ConfigSource         `json:"SourceList"`
	SendToList []wl_location.Location `json:"SendToList"`
}

// global
var (
	MainConfigFile = path.Join(wl_int.DirConfig, "config_service_config.json")
	RootDir        = path.Join(wl_int.DirData, "config_service")
	DataDir        = wl_int.DirData
	TmpDir         = wl_int.DirTmp
	PKIPath        = wl_int.DirPKI
	CACrtFile      = path.Join(wl_int.DirPKI, wl_int.CACrtFile)
	ServerCrtFile  = path.Join(wl_int.DirPKI, wl_int.ServerCrtFile)
	ServerKeyFile  = path.Join(wl_int.DirPKI, wl_int.ServerKeyFile)
)

var mainConfig MainConfig
var cc *wl_int.CommonConfig

func initGlobals() error {
	basepath, err := wl_fs.GetExecutableFullpath()
	if err != nil {
		return nil
	}

	MainConfigFile = path.Join(basepath, MainConfigFile)
	contentBytes, err := os.ReadFile(MainConfigFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(contentBytes, &mainConfig)
	if err != nil {
		return nil
	}
	cc, err = wl_int.LoadCommonConfig(basepath, AESKey, AESIV)
	if err != nil {
		return err
	}

	RootDir = path.Join(basepath, RootDir)
	DataDir = path.Join(RootDir, DataDir)
	TmpDir = path.Join(RootDir, TmpDir)

	CACrtFile = path.Join(basepath, CACrtFile)
	ServerCrtFile = path.Join(basepath, ServerCrtFile)
	ServerKeyFile = path.Join(basepath, ServerKeyFile)

	return nil
}
