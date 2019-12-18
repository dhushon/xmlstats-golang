/**
MIT License

Copyright (c) 2019 Dan Hushon

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import "net/http"
import "fmt"
import "io/ioutil"
import "encoding/json"
import "os"
import "time"

// Configuration variables including URL, BEARERTOKEN and USERAGENT are pre-requisites and should be set as
// Environment Variables
//	XMLSTATS_URL =  "https://erikberg.com/"
//	XMLSTATS_BEARERTOKEN =  "xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxx"
//	XMLSTATS_USERAGENT =  "Golang_XMLStatsRobot/0.0 (someone@example.com)"

type xmlstatsTime time.Time

// Custom unmarshall side effect of time.Time not parsing RFC3339
func (xmlt *xmlstatsTime) UnmarshalJSON(bs []byte) error {
	var s string

	if err := json.Unmarshal(bs, &s); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*xmlt = xmlstatsTime(t)
	return nil
}

// Events ... set of events provided by xmlstats
type Events struct {
	// unmarshall side effect of time.Time not parsing RFC3339
	EventsDate xmlstatsTime `json:"events_date"`
	Count      int          `json:"count"`
	Event      []Event      `json:"event"`
}

// Event ... specific event
type Event struct {
}

func getRequestHeader(url string) (*http.Request, error) {
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return req, err
	}
	// setup token
	// curl -H "Authorization: Bearer f15a9af4-c6a6-4000-bc54-f24feb0d5158"
	// Create a Bearer string by appending string access token
	token, exists := os.LookupEnv("XMLSTATS_BEARERTOKEN")
	if !exists {
		fmt.Println("XMLSTATS_BEARERTOKEN not found, you should get your token from xmlstats registration ")
	}
	// add authorization header to the req
	req.Header.Add("Authorization", "Bearer "+token)

	// get
	agent, exists := os.LookupEnv("XMLSTATS_USERAGENT")
	if !exists {
		fmt.Println("XMLSTATS_USERAGENT not found, should include your website or email credential")
	}
	// per instructions set user-information to prevent robot blocking
	req.Header.Set("User-Agent", agent)

	//req.Header.Set("Host", "domain.tld")
	return req, err
}

func getEvents(body []byte) (*Events, error) {
	var ev Events
	err := json.Unmarshal(body, &ev)
	if err != nil {
		fmt.Println("parsing error", err)
	}
	return &ev, err
}

func main() {
	fmt.Println("Loading environment configuration constants")

	//construct the event header
	baseurl, exists := os.LookupEnv("XMLSTATS_URL")
	if !exists {
		fmt.Println("XMLSTATS_URL not found, should include your website or email credential")
		baseurl = "https://erikberg.com/"
	}
	req, err := getRequestHeader(baseurl + "events.json")
	if err != nil {
		fmt.Printf("The HTTP request header building failed with error %s\n", err)
		return
	}

	// Send req using http Client
	client := &http.Client{}
	fmt.Println("doing HTTP GET")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("The HTTP request header building failed with error %s\n", err)
	} else if resp.StatusCode != 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("The HTTP request header building failed with error %s : %s\n", resp.Status, data)
	} else {
		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(data))
		fmt.Println("parsing HTTP response")
		events, _ := getEvents(data)
		// unmarshall side effect of time.Time not parsing RFC3339
		//The fmt.Println invokes the Time's .String() function that returns the time in the following format:
		//"2006-01-02 15:04:05.999999999 -0700 MST"
		//Which as you see contains both the timezone offset and the timezone name.
		//In your our case there is no timezone name known for the time, so it outputs the offset twice.
		fmt.Println(time.Time(events.EventsDate))
	}
	fmt.Println("Terminating the application...")
}
