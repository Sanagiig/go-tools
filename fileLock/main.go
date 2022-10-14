package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"processBar"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

type signal struct{}

var c = make(chan signal, runtime.NumCPU())
var fileList = make([]string, 0)
var dirList = make([]string, 0)
var encryptNum int
var password string
var pb *processBar.ProcessBar

func reverse[T any](list []T) {
	rv := reflect.ValueOf(list)
	if rv.Kind() == reflect.Slice {
		ll := len(list)
		for i := 0; i < ll/2; i++ {
			j := ll - i - 1
			list[i], list[j] = list[j], list[i]
		}
	} else {
		fmt.Println("reverse arg not a Slice")
	}
}

func walk(dir string) error {
	if strings.HasPrefix(dir, ".") {
		pw, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = path.Join(pw, dir)
	}

	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		if info.IsDir() {
			if path != dir {
				dirList = append(dirList, path)
			}
		} else if !strings.HasSuffix(info.Name(), LockSuffix) {
			fileList = append(fileList, path)
		}
		return nil
	})
}

func run(isUnlock bool) {
	var fn func(path string) error
	wg := sync.WaitGroup{}
	idx := 0

	if isUnlock {
		fn = Unlock
	} else {
		fn = Lock
	}

	for idx < len(fileList) {
		select {
		case c <- signal{}:
			wg.Add(1)
			go func(i int) {
				if err := fn(fileList[i]); err != nil {
					fmt.Println(err.Error())
				}
				pb.Process(1)
				wg.Done()
			}(idx)
			idx++
		}
	}
	wg.Wait()
	reverse(dirList)
	for _, dir := range dirList {
		if err := dirRename(dir, isUnlock); err != nil {
			fmt.Println(err.Error())
		}
		pb.Process(1)
	}
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	lastSplitIdx := strings.LastIndex(wd, "\\")
	if lastSplitIdx != -1 {
		wd = wd[:lastSplitIdx]
	}

	dir := flag.String("d", "", "需要操作的目录")
	u := flag.Bool("u", false, "目录解锁")
	flag.StringVar(&password, "p", "1234567812345678", "密钥")
	flag.IntVar(&encryptNum, "s", 256, "加密数据大小( Byte )")
	flag.Parse()

	if *dir == "" {
		fmt.Println("请输入正确目录")
		return
	}

	if err = walk(*dir); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("总共要加密 %v 个文件夹，%v 个文件\n", len(dirList), len(fileList))
	pb = processBar.NewBar(float64(len(dirList) + len(fileList)))
	pb.Start()
	run(*u)
}
