package internal

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
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

func FileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
