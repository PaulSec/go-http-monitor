package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	_ "fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
		URL          string  `yaml:"url"`
		StatusCode   *int    `yaml:"status_code"`
		Match        *string `yaml:"match"`
		ResponseTime *int    `yaml:"response_time"`
		TCP          string  `yaml:"tcp"`
		Port         *int    `yaml:"port"`
	} `yaml:"checks"`
}

// Config has been created
type CheckOutput struct {
	Resource string `json:"resource"`
	Status   string `json:"available"`
	Elapsed  string `json:"elapsed"`
}

type JsonOutput struct {
	Results []CheckOutput `json:"checks"`
}

func addEntry(results []CheckOutput, url string, active bool, elapsed time.Duration) []CheckOutput {
	check := &CheckOutput{
		Resource: url,
		Status:   strconv.FormatBool(!active),
		Elapsed:  elapsed.String(),
	}
	results = append(results, *check)
	return results
}

func main() {

	filenamePtr := flag.String("file", "monitor.yml", "Monitoring file")
	flag.Parse()

	hostUnreachable := false
	file, err := os.Open(*filenamePtr)
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

	results := &JsonOutput{}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: y.Insecure}
	client := http.Client{
		Timeout: time.Duration(y.TimeoutRequest) * time.Second,
	}

	for index, plugin := range y.Checks {
		_ = index
		tmpString := ""

		start := time.Now()

		t := time.Now()

		if strings.Contains(plugin.URL, "http") {
			resp, err := client.Get(plugin.URL)
			elapsed := t.Sub(start)

			// if we fail connecting to the host
			if err != nil {
				tmpString = "[NOK] " + plugin.URL + "\n"
				fmt.Printf(ErrorColor, tmpString)
				hostUnreachable = true

				results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed)
				continue
			}

			// if the status code does not correspond
			if plugin.StatusCode != nil && *plugin.StatusCode != resp.StatusCode {
				tmpString = "[NOK] " + plugin.URL + "\n"
				fmt.Printf(ErrorColor, tmpString)
				hostUnreachable = true

				results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed)
				continue
			}

			// if your search string does not appear in the response body
			content, err := ioutil.ReadAll(resp.Body)
			if plugin.Match != nil && !strings.Contains(string(content), *plugin.Match) {
				tmpString = "[NOK] " + plugin.URL + "\n"
				fmt.Printf(ErrorColor, tmpString)
				hostUnreachable = true

				results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed)
				continue
			}

			// if http response takes more time than expected
			if plugin.ResponseTime != nil {
				responseTimeDuration := time.Duration(*plugin.ResponseTime) * time.Millisecond
				if responseTimeDuration-elapsed < 0 {
					responseTime := strconv.Itoa(*plugin.ResponseTime)
					tmpString = "[NOK]  " + plugin.URL + ", Elapsed time: " + elapsed.String() + " instead of " + responseTime + "\n"
					fmt.Printf(ErrorColor, tmpString)
					hostUnreachable = true

					results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed)
					continue
				}
			}

			tmpString = "[OK] " + plugin.URL + "\n"
			fmt.Printf(NoticeColor, tmpString)
			results.Results = addEntry(results.Results, plugin.URL, true, elapsed)
		} else if plugin.TCP != "" {
			servAddr := plugin.TCP + ":" + strconv.Itoa(*plugin.Port)
			tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			_ = conn
			// fmt.Println("Foobar?")
			elapsed := t.Sub(start)
			if err != nil { // error on tcp connect
				hostUnreachable = true
				tmpString = "[NOK] TCP:" + servAddr + "\n"
				fmt.Printf(ErrorColor, tmpString)
				results.Results = addEntry(results.Results, servAddr, hostUnreachable, elapsed)
				continue
			} else if plugin.ResponseTime != nil { // error on connection
				responseTimeDuration := time.Duration(*plugin.ResponseTime) * time.Millisecond
				if responseTimeDuration-elapsed < 0 {
					responseTime := strconv.Itoa(*plugin.ResponseTime)
					tmpString = "[NOK] TCP:" + servAddr + ", Elapsed time: " + elapsed.String() + " instead of " + responseTime + "\n"
					fmt.Printf(ErrorColor, tmpString)
					hostUnreachable = true
					results.Results = addEntry(results.Results, plugin.URL, hostUnreachable, elapsed)
					continue
				}
			}
			tmpString = "[OK] TCP:" + servAddr + "\n"
			fmt.Printf(NoticeColor, tmpString)
			results.Results = addEntry(results.Results, servAddr, true, elapsed)
		}
	}

	jsonFile, _ := json.MarshalIndent(results, "", " ")
	_ = ioutil.WriteFile("output.json", jsonFile, 0644)

	// if any host is unreachable, exit(1) to fail execution
	if hostUnreachable {
		os.Exit(1)
	}
	os.Exit(0)
}
