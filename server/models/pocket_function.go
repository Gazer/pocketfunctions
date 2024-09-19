package models

import (
	"bufio"
	"fmt"
	"os"
)

type PocketFunction struct {
	Uri  string
	Code string
	Id   int
}

func (f *PocketFunction) VendorPath() string {
	return fmt.Sprintf("../dist/executors/%d/vendor", f.Id)
}

func (f *PocketFunction) CodePath() string {
	return fmt.Sprintf("../dist/executors/%d/vendor/%s", f.Id, f.Code)
}

func (f *PocketFunction) PubspecPath() string {
	return fmt.Sprintf("../dist/executors/%d/pubspec.yaml", f.Id)
}

func (f *PocketFunction) BasePath() string {
	return fmt.Sprintf("../dist/executors/%d", f.Id)
}

func (f *PocketFunction) ReadPubspec() ([]string, error) {
	inFile, err := os.Open(f.PubspecPath())
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	var lines []string

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}
