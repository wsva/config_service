package main

import (
	"path"

	wl_crypto "github.com/wsva/lib_go/crypto"
	wl_fs "github.com/wsva/lib_go/fs"
)

// config file on server side, used to encrypt all files
type ConfigFileServer struct {
	DirectoryOnServer string `json:"DirectoryOnServer"`
	FilenameOnServer  string `json:"FilenameOnServer"`
	Content           string `json:"Content"`
}

func (c *ConfigFileServer) Encrypt(aeskey, aesiv string) error {
	encryptedContent, err := wl_crypto.AES256Encrypt(aeskey, aesiv, c.Content)
	if err != nil {
		return err
	}
	c.Content = encryptedContent
	return nil
}

func (c *ConfigFileServer) Write(directory string) error {
	return wl_fs.WriteFile(path.Join(directory, c.DirectoryOnServer),
		c.FilenameOnServer, c.Content)
}
