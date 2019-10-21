package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	lru "github.com/hashicorp/golang-lru"
)

const jsonAcceptType = "application/json"

const datadogOutputTimeFormat = "2006-01-02T15:04:05.000Z"
const datadogInputTimeFormat = "2006-01-02 15:04:05"

// Stores recent log messages. Datadog doesn't have any methods for preventing duplicates or overlaps, so we have to
// filter them out ourselves.
var msgCache, _ = lru.New(1024)

// Simple structure to hold a single log message.
type logMessage struct {
	id        string
	timestamp time.Time
	fields    map[string]string
}

// Fetch all messages that match the settings in the options.
func fetchMessages(opts *options, startingId string) (result []logMessage, nextId string) {
	api := messageAPIURI(opts)
	jsonBytes := callDatadog(opts, api, jsonAcceptType)
	messages := getJSONArray(jsonBytes, "logs")
	nextId = getJSONString(jsonBytes, "nextLogId")
	_, _ = jsonparser.ArrayEach(messages, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		id := getJSONString(value, "id")
		msg := getJSONSimpleMap(value, "content")
		tsStr := msg[timestampField]
		// 2019-10-03T13:22:52.882Z

		ts, err := time.Parse(datadogOutputTimeFormat, tsStr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid json timestamp: %s - %s\n", tsStr, err.Error())
		}
		if err == nil {
			msgObj := logMessage{
				id:        id,
				timestamp: ts,
				fields:    msg,
			}
			result = append(result, msgObj)
		}
	})

	return result, nextId
}

// Compute the API Uri to call. Determined by examining the command-line options.
func messageAPIURI(opts *options) (uri string) {
	api := "{\"query\": \"%QUERY%\",\"time\": {\"from\": \"%START%\", \"to\": \"%END%\"}, \"sort\": \"desc\", \"limit\": %LIMIT%}"
	if opts.startDate == nil || opts.endDate == nil {
		// uri = fmt.Sprintf(relativeSearch, strconv.Itoa(opts.timeRange))
		strings.Replace(api, "%START%", "now - "+strconv.Itoa(opts.timeRange)+"s", 1)
		strings.Replace(api, "%END%", "now", 1)
	} else {
		strings.Replace(api, "%START%", (*opts.startDate).Format(datadogInputTimeFormat), 1)
		strings.Replace(api, "%END%", (*opts.endDate).Format(datadogInputTimeFormat), 1)
	}
	if opts.limit > 0 {
		strings.Replace(api, "%LIMIT%", strconv.Itoa(opts.limit), 1)
	} else {
		strings.Replace(api, "%LIMIT%", "50", 1)
	}
	if len(opts.query) > 0 {
		strings.Replace(api, "%QUERY%", opts.query, 1)
	} else {
		strings.Replace(api, "%QUERY%", "*", 1)
	}

	return uri
}

// Common entry-point for calls to Datadog.
func callDatadog(opts *options, api string, acceptType string) []byte {
	cfg := opts.serverConfig

	apiKey := cfg.ApiKey()
	applicationKey := cfg.ApplicationKey()

	if acceptType == jsonAcceptType {
		uri := fmt.Sprintf("https://api.datadoghq.com/api/v1/logs-queries/list?api_key=%s&application_key=%s", apiKey, applicationKey)
		return readBytes(uri, api)
	}

	return nil
}

// Return the raw bytes sent by Datadog.
func readBytes(uri string, body string) []byte {
	return fetch(uri, body, jsonAcceptType)
}

// Low-level HTTP call to Datadog.
func fetch(uri string, api string, acceptType string) []byte {
	var client *http.Client
	client = &http.Client{}

	req, err := http.NewRequest("POST", uri, strings.NewReader(api))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Request is malformed: %s\n", err.Error())
		os.Exit(1)
	}
	req.Header.Add("Accept", acceptType)
	resp, err := client.Do(req)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to Datadog: %s\n", err.Error())
		os.Exit(1)
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to read content from Datadog: %s\n", err.Error())
		os.Exit(1)
	}

	return body
}
