package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"strings"
)

const LockPrefix = "locked_"
const LockSuffix = ".locked"

func empty(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

func Lock(fp string) error {
	defer func() {
		<-c
	}()

	parts := strings.Split(fp, "\\")
	filename := parts[len(parts)-1]

	if strings.HasPrefix(filename, LockPrefix) {
		return nil
	}

	f, err := os.OpenFile(fp, os.O_RDWR, 5)
	if err != nil {
		return err
	}

	src := make([]byte, encryptNum)

	// 需要记录从源文件读取的字节数， 不然小文件会有多余的 null
	n, err := f.Read(src)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(LockPrefix)
	buf.Write(src[:n])

	dst, err := AesEncrypt([]byte(password), buf.Bytes())
	if err != nil {
		return err
	}

	empty(src)
	_, err = f.WriteAt(src[:n], 0)
	f.Close()
	if err != nil {
		return err
	}

	return update(fp, filename, string(dst))
}

func Unlock(fp string) error {
	defer func() {
		<-c
	}()
	parts := strings.Split(fp, "\\")
	filename := parts[len(parts)-1]
	var src []byte

	if !strings.HasPrefix(filename, LockPrefix) {
		return nil
	}

	f, err := os.OpenFile(fp, os.O_WRONLY, 5)
	if err != nil {
		return err
	}

	lf, err := os.OpenFile(fp+LockSuffix, os.O_RDONLY, 5)
	if err != nil {
		return err
	}

	src, err = io.ReadAll(lf)
	lf.Close()
	if err != nil {
		return err
	}

	origData, err := AesDecrypt([]byte(password), src)
	if err != nil || !bytes.HasPrefix(origData, []byte(LockPrefix)) {
		panic("\n解密失败：密码错误或数据损坏")
		os.Exit(0)
	}

	_, err = f.WriteAt(origData[len(LockPrefix):], 0)
	f.Close()
	if err != nil {
		return err
	}

	return update(fp, filename, "")
}

func update(filepath, filename string, encData string) error {
	var newname string

	if encData == "" {
		if !strings.HasPrefix(filename, LockPrefix) {
			return nil
		}

		res, err := base64.URLEncoding.DecodeString(filename[len(LockPrefix):])
		if err != nil {
			return err
		}

		newname = string(res)

		if err = os.Remove(filepath + LockSuffix); err != nil {
			return err
		}

	} else {
		if strings.HasPrefix(filename, LockPrefix) {
			return nil
		}

		newname = LockPrefix + base64.URLEncoding.EncodeToString([]byte(filename))

		//创建lock file
		lockfile, err := os.Create(strings.Replace(filepath, filename, newname+LockSuffix, 1))
		if err != nil {
			return err
		}

		defer lockfile.Close()
		if _, err = lockfile.Write([]byte(encData)); err != nil {
			return err
		}
	}

	err := os.Rename(filepath, strings.Replace(filepath, filename, newname, 1))
	if err != nil {
		return err
	}

	return nil
}

func base64Encode(str string) (string, error) {
	res := base64.URLEncoding.EncodeToString([]byte(str))
	return res, nil
}

func base64Decode(str string) (string, error) {
	res, err := base64.URLEncoding.DecodeString(str)
	return string(res), err
}

func dirRename(dirpath string, isUnlock bool) error {
	var newDirName string
	var err error
	parts := strings.Split(dirpath, "\\")
	dirName := parts[len(parts)-1]

	// 处理过的文件夹
	if isUnlock && !strings.HasPrefix(dirName, LockPrefix) || !isUnlock && strings.HasPrefix(dirName, LockPrefix) {
		return nil
	}

	if isUnlock {
		newDirName, err = base64Decode(dirName[len(LockPrefix):])
	} else {
		newDirName, _ = base64Encode(dirName)
		newDirName = LockPrefix + newDirName
	}

	if err != nil {
		return err
	}

	newDirPath := strings.Join(parts[:len(parts)-1], "\\") + "\\" + newDirName
	return os.Rename(dirpath, newDirPath)
}
