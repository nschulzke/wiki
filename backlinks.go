package main

import (
	"os"
	"strings"
)

func backlinks(filename string) ([]string, error) {
	files, err := os.ReadDir("data")
	if err != nil {
		return nil, err
	}
	links := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == filename {
			continue
		}
		bareName := strings.TrimSuffix(file.Name(), ".md")
		p, err := loadPage(bareName, false)
		if err != nil {
			return nil, err
		}
		// Look for [[filename]] in the body of the page.
		if strings.Contains(string(p.Body), "[["+filename+"]]") {
			links = append(links, bareName)
		}
	}
	return links, nil
}
