package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	wl_compress "github.com/wsva/lib_go/compress"
	wl_fs "github.com/wsva/lib_go/fs"
)

func GenerateConfigDataRoot(list []ConfigSource) error {
	//遍历所有的源，生成汇总合并的index
	var indexlist ConfigIndexList
	for _, v := range list {
		l, err := v.ConfigIndexList()
		if err != nil {
			return err
		}
		if indexlist.CheckConflict(l) {
			return errors.New("conflict found among sources")
		}
		indexlist = append(indexlist, l...)
	}

	//比对index，如果相同，说明没有必要更新
	//旧文件可能还不存在，因此此处有错误也正常，也继续往下走
	indexfileold := path.Join(DataDir, ConfigIndexFile)
	indexlistold, _ := LoadConfigIndexListFromFile(indexfileold)
	if indexlist.Compare(indexlistold) {
		return nil
	}

	//清空temp目录
	err := initDataDirectoryTemp()
	if err != nil {
		return err
	}

	//获取所有config文件，先放在内存中，暂不写入磁盘
	filelist, err := indexlist.ConfigFileServerList()
	if err != nil {
		return err
	}

	//将config文件内容加密
	for _, v := range filelist {
		err = v.Encrypt(AESKey, AESIV)
		if err != nil {
			return err
		}
	}

	//将config文件写入磁盘，这里是写入temp目录
	for _, v := range filelist {
		err = v.Write(TmpDir)
		if err != nil {
			return err
		}
	}

	//将index文件写入temp目录
	err = indexlist.Write(TmpDir)
	if err != nil {
		return err
	}

	//移除data目录
	err = removeDataDirectory()
	if err != nil {
		return err
	}

	//将temp目录重命名为data目录
	err = moveTempToDataDirectory()
	if err != nil {
		return err
	}

	return nil
}

func GenerateConfigDataBranch(list []ConfigSource) error {
	//只使用第一个下载成功的zip文件，其他的不再使用
	zipfle := path.Join(RootDir, ConfigDataZipFile)
	for _, v := range list {
		if !v.Enable {
			continue
		}
		err := v.DownloadZipFile(zipfle)
		if err != nil {
			return err
		} else {
			break
		}
	}

	//比对zip文件的hash，如果跟data目录中的hash相同，就不必更新了
	hashnew, err := wl_fs.GetFileHashSHA256(zipfle)
	if err != nil {
		return err
	}
	hasholdfile := path.Join(DataDir, ConfigDataHashFile)
	hashold, _ := os.ReadFile(hasholdfile)
	if hashnew == string(hashold) {
		return nil
	}

	//删除data目录
	removeDataDirectory()

	//解压zip文件
	err = wl_compress.ZipDecompressFile(zipfle, RootDir)
	if err != nil {
		return err
	}

	//更新data目录中的hash文件
	err = wl_fs.WriteFile(DataDir, ConfigDataHashFile, string(hashold))
	if err != nil {
		return err
	}

	return nil
}

// 初始化temp目录，其实就是先删除，再新建一个空目录
func initDataDirectoryTemp() error {
	os.RemoveAll(TmpDir)
	_, err := os.Stat(TmpDir)
	if err == nil {
		return fmt.Errorf("remove " + TmpDir + " failed")
	}
	os.MkdirAll(TmpDir, 0777)
	_, err = os.Stat(TmpDir)
	if err != nil {
		return fmt.Errorf("create " + TmpDir + " failed")
	}
	return nil
}

// 删除data目录
func removeDataDirectory() error {
	os.RemoveAll(DataDir)
	_, err := os.Stat(DataDir)
	if err == nil {
		return fmt.Errorf("remove " + DataDir + " failed")
	}
	return nil
}

// 将temp目录重命名为data目录
func moveTempToDataDirectory() error {
	err := removeDataDirectory()
	if err != nil {
		return err
	}
	os.Rename(TmpDir, DataDir)
	_, err = os.Stat(DataDir)
	if err != nil {
		return fmt.Errorf("rename failed: " + TmpDir + " -> " + DataDir)
	}
	return nil
}
