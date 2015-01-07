package util

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// var DEFAULT_MODE os.FileMode = os.ModePerm

func Fexists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FTouch(path string) error {
	return FTouch2(path, os.ModePerm)
}
func FTouch2(path string, fm os.FileMode) error {
	f, err := os.Open(path)
	if err != nil {
		p := filepath.Dir(path)
		if !Fexists(p) {
			err := os.MkdirAll(p, fm)
			if err != nil {
				return err
			}
		}
		f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fm)
		if f != nil {
			defer f.Close()
		}
		return err
	}
	defer f.Close()
	fi, _ := f.Stat()
	if fi.IsDir() {
		return errors.New("can't touch path")
	}
	return nil
}
func FWrite(path, data string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = f.WriteString(data)
	return err
}
func FAppend(path, data string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = f.WriteString(data)
	return err
}
func FCopy(src string, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	return err
}
func ReadLine(r *bufio.Reader, limit int, end bool) ([]byte, error) {
	var isPrefix bool = true
	var bys []byte
	var tmp []byte
	var err error
	for isPrefix {
		tmp, isPrefix, err = r.ReadLine()
		if err != nil {
			return nil, err
		}
		bys = append(bys, tmp...)
	}
	if end {
		bys = append(bys, '\n')
	}
	return bys, nil
}
func ReadW(r *bufio.Reader, p []byte, last *int64) error {
	l := len(p)
	all := 0
	buf := p
	for {
		l_, err := r.Read(buf)
		if err != nil {
			return err
		}
		*last = Now()
		all += l_
		if all < l {
			buf = p[all:]
			continue
		} else {
			break
		}
	}
	return nil
}

func Timestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}
func Time(timestamp int64) time.Time {
	return time.Unix(0, timestamp*1e6)
}
func Now() int64 {
	return Timestamp(time.Now())
}
func AryExist(ary interface{}, obj interface{}) bool {
	switch reflect.TypeOf(ary).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(ary)
		for i := 0; i < s.Len(); i++ {
			if obj == s.Index(i).Interface() {
				return true
			}
		}
		return false
	default:
		return false
	}
}

var C_SH string = "/bin/bash"

func Exec(args ...string) (string, error) {
	bys, err := exec.Command(C_SH, "-c", strings.Join(args, " ")).Output()
	return string(bys), err
}

func IsType(v interface{}, t string) bool {
	t = strings.Trim(t, " \t")
	if v == nil || len(t) < 1 {
		return false
	}
	return reflect.Indirect(reflect.ValueOf(v)).Type().Name() == t
}

func Append(ary []interface{}, args ...interface{}) []interface{} {
	for _, arg := range args {
		ary = append(ary, arg)
	}
	return ary
}

func List(root string, reg string) []string {
	return ListFunc(root, reg, func(t string) string {
		return t
	})
}
func ListFunc(root string, reg string, f func(t string) string) []string {
	pathes := []string{}
	regx := regexp.MustCompile(reg)
	filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if regx.MatchString(path) {
			pathes = append(pathes, f(path))
		}
		return nil
	})
	return pathes
}

func FileProtocolPath(t string) (string, error) {
	t = strings.Trim(t, " \t")
	if strings.HasPrefix(t, "file://") {
		return t, nil
	}
	t, _ = filepath.Abs(t)
	t = strings.Replace(t, "\\", "/", -1)
	return "file://" + t, nil
}

func Str2Int(s string) ([]int64, error) {
	vals := []int64{}
	ss := strings.Split(s, ",")
	for _, str := range ss {
		str = strings.Trim(str, "\t ")
		if len(str) < 1 {
			continue
		}
		v, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			vals = append(vals, v)
		} else {
			return nil, err
		}
	}
	return vals, nil
}

func Int2Str(vals []int64) string {
	str := ""
	for _, v := range vals {
		str = fmt.Sprintf("%s%d,", str, v)
	}
	return strings.Trim(str, ",")
}
func Is2Ss(vals []int64) []string {
	str := []string{}
	for _, v := range vals {
		str = append(str, fmt.Sprintf("%s%d,", str, v))
	}
	return str
}
