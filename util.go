package main

import (
	"io/ioutil"
	"os"
	"strings"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func TempDir() (string, error) {
	return ioutil.TempDir("", "")
}

func Chomp(s string) string {
	return strings.TrimRight(s, "\n")
}

func StringInSlice(slice []string, s string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
