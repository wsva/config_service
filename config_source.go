package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"

	wl_fs "github.com/wsva/lib_go/fs"
	wl_location "github.com/wsva/lib_go/location"
)

type ConfigSource struct {
	Enable bool `json:"Enable"`
	//Role root: MongoDB, Directory
	//Role branch: ZipFile
	SourceType string          `json:"SourceType"`
	Source     json.RawMessage `json:"Source"`
}

// ConfigIndexList is used by Role root
func (s *ConfigSource) ConfigIndexList() (ConfigIndexList, error) {
	if !s.Enable {
		return nil, errors.New("not enabled")
	}
	if s.SourceType != "Directory" {
		return nil, errors.New("SourceType for Root must be Directory")
	}
	var source SourceDirectory
	err := json.Unmarshal(s.Source, &source)
	if err != nil {
		return nil, err
	}
	return source.ConfigIndexList()
}

// DownloadZipFile is used by Role branch
func (s *ConfigSource) DownloadZipFile(dest string) error {
	if !s.Enable {
		return errors.New("not enabled")
	}
	if s.SourceType != "ZipFile" {
		return errors.New("SourceType for Branch must be ZipFile")
	}
	var location wl_location.Location
	err := json.Unmarshal(s.Source, &location)
	if err != nil {
		return err
	}
	return location.Download(dest)
}

type SourceDirectory struct {
	Directory string `json:"Directory"`
	IndexFile string `json:"IndexFile"`
}

func (s *SourceDirectory) init() {
	basepath, _ := wl_fs.GetExecutableFullpath()
	s.Directory = strings.ReplaceAll(s.Directory, "{BasePath}", basepath)
}

func (s *SourceDirectory) ConfigIndexList() (ConfigIndexList, error) {
	s.init()
	contentBytes, err := os.ReadFile(path.Join(s.Directory, s.IndexFile))
	if err != nil {
		return nil, err
	}
	var result ConfigIndexList
	err = json.Unmarshal(contentBytes, &result)
	if err != nil {
		return nil, err
	}
	for k := range result {
		result[k].source = s
	}
	return result, nil
}

func (s *SourceDirectory) ConfigFileServerList(index ConfigIndex) ([]*ConfigFileServer, error) {
	fileInfoList, err := os.ReadDir(path.Join(s.Directory, index.Name))
	if err != nil {
		return nil, err
	}
	var result []*ConfigFileServer
	for _, v := range fileInfoList {
		contentBytes, err := os.ReadFile(path.Join(s.Directory, index.Name, v.Name()))
		if err != nil {
			return nil, err
		}
		result = append(result, &ConfigFileServer{
			DirectoryOnServer: index.DirectoryOnServer,
			FilenameOnServer:  v.Name(),
			Content:           string(contentBytes),
		})
	}
	return result, nil
}
