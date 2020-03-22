package main

import (
	"crypto/tls"
	"fmt"
	_ "fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config for the colors used in the tool
const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

// Config has been created
type Config struct {
	Insecure       bool `yaml:"insecure"`
	TimeoutRequest int  `yaml:"timeout_seconds"`
	Checks         []struct {
		URL        string  `yaml:"url"`
		StatusCode *int    `yaml:"status_code"`
		Match      *string `yaml:"match"`
	} `yaml:"checks"`
}

func main() {

	hostUnreachable := false
	file, err := os.Open("monitor.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)

	y := Config{}

	err = yaml.Unmarshal([]byte(data), &y)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: y.Insecure}
	client := http.Client{
		Timeout: time.Duration(y.TimeoutRequest) * time.Second,
	}

	for index, plugin := range y.Checks {
		_ = index
		tmpString := ""

		resp, err := client.Get(plugin.URL)

		// if we fail connecting to the host
		if err != nil {
			tmpString = "[NOK] " + plugin.URL + "\n"
			fmt.Printf(ErrorColor, tmpString)
			hostUnreachable = true
			continue
		}

		// if the status code does not correspond
		if plugin.StatusCode != nil && *plugin.StatusCode != resp.StatusCode {
			tmpString = "[NOK] " + plugin.URL + "\n"
			fmt.Printf(ErrorColor, tmpString)
			hostUnreachable = true
			continue
		}

		// if your search string does not appear in the response body
		content, err := ioutil.ReadAll(resp.Body)
		if plugin.Match != nil && !strings.Contains(string(content), *plugin.Match) {
			tmpString = "[NOK] " + plugin.URL + "\n"
			fmt.Printf(ErrorColor, tmpString)
			hostUnreachable = true
			continue
		}

		tmpString = "[OK] " + plugin.URL + "\n"
		fmt.Printf(NoticeColor, tmpString)
	}

	if hostUnreachable {
		os.Exit(1)
	}
	os.Exit(0)
}
