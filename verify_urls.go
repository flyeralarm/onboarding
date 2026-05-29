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

// URLs that return a non-200 status code by design (e.g. bot protection).
// Maps URL -> expected status code. Loaded from url_exceptions.txt (format: "status_code url").
func loadUrlExceptions(filename string) map[string]int {
	exceptions := map[string]int{}
	if data, err := os.ReadFile(filename); err == nil {
		for _, m := range regexp.MustCompile(`(?m)^(\d+)\s(.+)$`).FindAllStringSubmatch(string(data), -1) {
			var code int
			fmt.Sscanf(m[1], "%d", &code)
			exceptions[m[2]] = code
		}
	}
	return exceptions
}

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
    urlExceptions := loadUrlExceptions(".github/url_exceptions.txt")

    var brokenUrls []string
    for _, urlElement := range urlElementRegex.FindAllStringSubmatch(fileContent, -1) {
        var url = urlElement[1]

        expectedCode := 200
		// look for http codes from url_exceptions.txt, fallback to code 200
		code, isException := urlExceptions[url]
		if isException {
			expectedCode = code
		}

		// code 0 means skip this URL entirely (e.g. known timeouts or connection issues)
		if expectedCode == 0 {
			fmt.Printf("Checking %s: SKIPPED\n", url)
			continue
		}

        fmt.Printf("Checking %s: ", url)

        req, err := http.NewRequest("GET", url, nil)
        req.Header.Add("User-Agent", "URL status code verification for the Flyeralarm onboarding resources; https://github.com/flyeralarm/onboarding")
        resp, err := httpClient.Do(req)

        errormessage := err
        if errormessage == nil {
            errormessage = errors.New(http.StatusText(resp.StatusCode))
            resp.Body.Close()
        }
        // accept the expected status code or a 200 
        if err != nil || (resp.StatusCode != expectedCode && !(isException && resp.StatusCode == 200)) {
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
