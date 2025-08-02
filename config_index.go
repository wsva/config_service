package main

import (
	"encoding/json"
	"os"

	wl_fs "github.com/wsva/lib_go/fs"
	wl_json "github.com/wsva/lib_go/json"
)

const (
	ConfigIndexFile = "sourceindex.json"
)

/*
ConfigIndex is the index in ConfigSource for Role root,
and is used for generating config data.
*/
type ConfigIndex struct {
	Name              string `json:"Name"`
	Version           int64  `json:"Version"`
	DirectoryOnServer string `json:"DirectoryOnServer"`
	FilenameOnClient  string `json:"FilenameOnClient"`
	Format            string `json:"Format"`

	source *SourceDirectory `json:"-"`
}

type ConfigIndexList []ConfigIndex

/*
Compare is used to decide whether to regenerate config data from source
*/
func (s ConfigIndexList) Compare(s1 ConfigIndexList) bool {
	if len(s) != len(s1) {
		return false
	}
	m1 := make(map[string]ConfigIndex)
	m2 := make(map[string]ConfigIndex)
	for _, v := range s {
		m1[v.DirectoryOnServer] = v
	}
	for _, v := range s1 {
		m2[v.DirectoryOnServer] = v
	}
	for k, v := range m1 {
		if _, exist := m2[k]; !exist {
			return false
		}
		if v.Version != m2[k].Version {
			return false
		}
	}
	for k, v := range m2 {
		if _, exist := m1[k]; !exist {
			return false
		}
		if v.Version != m1[k].Version {
			return false
		}
	}
	return true
}

func (s ConfigIndexList) Merge(s1 ConfigIndexList) ConfigIndexList {
	return append(s, s1...)
}

func (s ConfigIndexList) CheckConflict(s1 ConfigIndexList) bool {
	m1 := make(map[string]ConfigIndex)
	for _, v := range s {
		m1[v.DirectoryOnServer] = v
	}
	for _, v := range s1 {
		if _, exist := m1[v.DirectoryOnServer]; exist {
			return true
		}
	}
	return false
}

// destPath is a fullpath(without filename) to write to
func (s ConfigIndexList) Write(destPath string) error {
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}
	contentBytes := wl_json.Indent(jsonBytes)
	return wl_fs.WriteFile(destPath, ConfigIndexFile, string(contentBytes))
}

func (s ConfigIndexList) ConfigFileServerList() ([]*ConfigFileServer, error) {
	var result []*ConfigFileServer
	for _, v := range s {
		l, err := v.source.ConfigFileServerList(v)
		if err != nil {
			return nil, err
		}
		result = append(result, l...)
	}
	return result, nil
}

func LoadConfigIndexListFromJSON(jsonBytes []byte) (ConfigIndexList, error) {
	var list ConfigIndexList
	err := json.Unmarshal(jsonBytes, &list)
	return list, err
}

func LoadConfigIndexListFromFile(filename string) (ConfigIndexList, error) {
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return LoadConfigIndexListFromJSON(contentBytes)
}
