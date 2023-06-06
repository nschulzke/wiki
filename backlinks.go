package main

import (
	"os"
	"path/filepath"
	"strings"
)

func backlinks(filename string) ([]string, error) {
	var links []string
	err := filepath.WalkDir("data", func(path string, file os.DirEntry, err error) error {
		if file.IsDir() {
			return nil
		}
		if file.Name() == filename+".md" {
			return nil
		}
		bareName := strings.TrimSuffix(path, ".md")
		bareName = strings.TrimPrefix(bareName, "data/")
		p, err := loadPage(bareName, false)
		if err != nil {
			return err
		}
		if strings.Contains(string(p.Body), "[["+filename+"]]") {
			links = append(links, bareName)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return links, nil
}
