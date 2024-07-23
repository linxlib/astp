package internal

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

func GetKey(pkg string, name string) string {
	return pkg + "." + name
}
func GetKeyHash(pkg string, name string) string {
	return Md5(pkg + "." + name)
}
func Md5(origin string) string {
	w := md5.New()
	io.WriteString(w, origin)
	return fmt.Sprintf("%x", w.Sum(nil))
}

func Abs(origin string) string {
	f, _ := filepath.Abs(origin)
	return f
}
func WriteFile(fname string, src []byte, isClear bool) bool {
	_ = BuildDir(fname)
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !isClear {
		flag = os.O_CREATE | os.O_RDWR | os.O_APPEND
	}
	f, err := os.OpenFile(fname, flag, 0666)
	if err != nil {
		return false
	}
	_, _ = f.Write(src)
	_ = f.Close()
	return true
}
func BuildDir(absDir string) error {
	return os.MkdirAll(path.Dir(absDir), os.ModePerm) //生成多级目录
}
func FileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
func ReadFile(fname string) []byte {
	src, _ := os.ReadFile(fname)
	return src
}
