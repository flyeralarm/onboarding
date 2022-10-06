package main

import (
    "crypto/tls"
    "errors"
    "fmt"
    "net/http"
    "os"
    "regexp"
    "time"
)

func main() {
    fmt.Println("Verifying URLs..")

    readmeFile, err := os.ReadFile("README.md")
    if err != nil {
        fmt.Println("Could not find README!")
        os.Exit(1)
    }

    fileContent := string(readmeFile)
    urlElementRegex := regexp.MustCompile(`(?m)\[.+?]\(((http|https)://.+?)\)`)

    httpClient := http.Client{
        Timeout: 20 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{},
        },
    }

    var brokenUrls []string
    for _, urlElement := range urlElementRegex.FindAllStringSubmatch(fileContent, -1) {
        var url = urlElement[1]

        fmt.Printf("Checking %s: ", url)

        req, err := http.NewRequest("GET", url, nil)
        req.Header.Add("User-Agent", "URL status code verification for the Flyeralarm onboarding resources; https://github.com/flyeralarm/onboarding")
        resp, err := httpClient.Do(req)

        errormessage := err
        if errormessage == nil {
            errormessage = errors.New(http.StatusText(resp.StatusCode))
        }

        if err != nil || resp.StatusCode != 200 {
            brokenUrls = append(brokenUrls, url)
            fmt.Println("FAILED - ", errormessage)
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
