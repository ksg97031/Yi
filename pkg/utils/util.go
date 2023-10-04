package utils

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/thoas/go-funk"
)

/**
  @author: yhy
  @since: 2022/10/13
  @desc: //TODO
**/

func getDir(path string) string {
	return subString(path, 0, strings.LastIndex(path, "/"))
}

func subString(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < start || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}

func GetName(path string) string {
	if strings.HasSuffix(path, "/") {
		path = strings.TrimRight(path, "/")
	}

	ss := strings.Split(path, "/")

	return ss[len(ss)-1]
}

// WriteCounter Download progress bar
type WriteCounter struct {
	Total    uint64
	FileName string
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading [%s] ... %s complete", wc.FileName, humanize.Bytes(wc.Total))
}

func RandStr() string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ%*.!@"
	bytes := []byte(str)
	result := []byte{}
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(100)))
	for i := 0; i < 10; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}

// CodeqlDb Get the folder codeql-database.yml Update
func CodeqlDb(dir string) string {
	var filePath string
	var walkFunc = func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.Contains(path, "codeql-database.yml") {
				filePath = path
				return nil
			}
		}
		return nil
	}
	filepath.Walk(dir, walkFunc)
	return filepath.Dir(filePath)
}

func StringInSlice(s string, slice []string) bool {
	if slice == nil {
		return false
	}
	sort.Strings(slice)
	index := sort.SearchStrings(slice, s)
	if index < len(slice) && strings.ToLower(slice[index]) == strings.ToLower(s) {
		return true
	}
	return false
}

// isSupportedProtocol checks given protocols are supported
func isSupportedProtocol(value string) bool {
	return value == "http" || value == "https" || value == "socks5"
}

// Difference Find out the rules of the change. If it is added, it runs this rule.
func Difference(old, new []string) []string {
	_, s2 := funk.Difference(old, new)

	// {"1", "2", "5"} {"3", "5"} result [1 2] [3]
	// {"1", "2", "3", "4"} {"1", "2", "3"} result [4] []

	return s2.([]string)
}

// RunGitCommand Execute the encapsulation of any git command
func RunGitCommand(path, name string, arg ...string) (string, error) {
	gitpath := path // Get the current Git warehouse from the configuration file

	cmd := exec.Command(name, arg...)
	cmd.Dir = gitpath                // Specify the working directory as the GIT warehouse directory
	msg, err := cmd.CombinedOutput() // Hybrid output Stdout+stderr
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// When an error is reported exit status 1
	return string(msg), err
}
