package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

func main() {
	fmt.Println("Verifying URLs..")

	readmeFile, err := ioutil.ReadFile("README.md")
	if err != nil {
		fmt.Println("Could not find README!")
		os.Exit(1)
	}

	fileContent := string(readmeFile)
	urlElementRegex := regexp.MustCompile(`(?m)\[.+?]\(((http|https)://.+?)\)`)

	httpClient := http.Client{Timeout: 10 * time.Second}

	var brokenUrls []string
	for _, urlElement := range urlElementRegex.FindAllStringSubmatch(fileContent, -1) {
		var url = urlElement[1]

		fmt.Printf("Checking %s: ", url)

		resp, err := httpClient.Get(url)
		if err != nil || resp.StatusCode != 200 {
			brokenUrls = append(brokenUrls, url)
			fmt.Println("FAILED - ", err)
		} else {
			fmt.Println("OK")
		}
	}

	if len(brokenUrls) != 0 {
		fmt.Println("Broken URLs were found:")
		for _, brokenUrl := range brokenUrls {
			fmt.Println(brokenUrl)
		}

		os.Exit(1)
	}

	fmt.Println("No broken URLs found!")
	os.Exit(0)
}
