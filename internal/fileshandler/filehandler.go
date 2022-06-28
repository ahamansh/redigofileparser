package filehandler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/mytest/api/internal/constants"
)

type FileHanderInterface interface {
	ProcessFile(fileID string) ([]string, error)
}

type LocalFileHandler struct {
	regexp *regexp.Regexp
}

func GetLocalFileHandler() FileHanderInterface {
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	return &LocalFileHandler{regexp: re}
}

// func (l *LocalFileHandler) Loadfile(fileID string) error {

// 	fileNameWithPath := fmt.Sprintf("%s%s", "/Users/anshulsh/practice/crowdstrike/Archive/input/", fileID)
// 	data, err := ioutil.ReadFile(fileNameWithPath)

// 	l.fileData = data
// 	return nil
// }

func (l *LocalFileHandler) ProcessFile(fileID string) ([]string, error) {

	fileNameWithPath := fmt.Sprintf("%s%s", constants.DEFAULT_LOCAL_FILE_FOLDER, fileID)
	data, err := ioutil.ReadFile(fileNameWithPath)
	if err != nil {
		return []string{}, errors.New("Unable to process the file")
	}

	retArr := []string{}

	submatchall := l.regexp.FindAllString(string(data), -1)
	for _, element := range submatchall {
		retArr = append(retArr, element)
		fmt.Println(element)
	}

	return retArr, nil
}
