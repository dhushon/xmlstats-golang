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
import "io"
import "io/ioutil"
import "encoding/json"
import "os"
import "time"
import "compress/gzip"
import "errors"

var extractTime = time.Now()

const extractSrc = "xmlstats.com"

// Configuration variables including URL, BEARERTOKEN and USERAGENT are pre-requisites and should be set as
// Environment Variables
//	XMLSTATS_URL =  "https://erikberg.com/"
//	XMLSTATS_BEARERTOKEN =  "xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxx"
//	XMLSTATS_USERAGENT =  "Golang_XMLStatsRobot/0.0 (someone@example.com)"

// XmlstatsTime is a custom Time parser
type XmlstatsTime time.Time

// UnmarshalJSON ... Custom unxmarshall side effect of time.Time not parsing RFC3339
func (xmlt *XmlstatsTime) UnmarshalJSON(bs []byte) error {
	var s string

	if err := json.Unmarshal(bs, &s); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*xmlt = XmlstatsTime(t)
	return nil
}

// Events ... set of events provided by xmlstats
type Events struct {
	// unmarshall side effect of time.Time not parsing RFC3339
	Extracted     time.Time `json:"extract_time"`
	ExtractedSrc  string    `json:"extract_src"`
	EventsDate    time.Time `json:"events_date"`
	Count         int       `json:"count"`
	Event         []Event   `json:"event" binding:"required"`
	CollectedData time.Time `json:"last_updated"`
}

// Event ... specific event
type Event struct {
	Extracted        time.Time `json:"extract_time"`
	ExtractedSrc     string    `json:"extract_src"`
	EventID          string    `json:"event_id" binding:"required"`
	EventStatus      string    `json:"event_status"`
	Sport            string    `json:"sport"`
	SeasonType       string    `json:"season_type"`
	AwayTeam         Team      `json:"away_team"`
	HomeTeam         Team      `json:"home_team"`
	SiteInfo         Site      `json:"site"`
	AwayPeriodScores []int     `json:"away_period_scores,omitempty"`
	HomePeriodScores []int     `json:"home_period_scores,omitempty"`
	AwayScore        int       `json:"away_points_scored"`
	HomeScore        int       `json:"home_points_scored"`
}

// Team ... specific team
//"team_id":"memphis-grizzlies","abbreviation":"MEM","active":true,"first_name":"Memphis","last_name":"Grizzlies",
//"conference":"West","division":"Southwest","site_name":"FedExForum","city":"Memphis","state":"Tennessee",
//"full_name":"Memphis Grizzlies"
type Team struct {
	Extracted    time.Time `json:"extract_time"`
	ExtractedSrc string    `json:"extract_src"`
	TeamID       string    `json:"team_id" binding:"required"`
	Abbreviation string    `json:"abbreviation,omitempty"`
	Active       bool      `json:"active,omitempty"`
	FName        string    `json:"first_name,omitempty"`
	LName        string    `json:"last_name,omitempty"`
	Conference   string    `json:"conference,omitempty"`
	Division     string    `json:"division,omitempty"`
	SiteName     string    `json:"site_name,omitempty"`
	City         string    `json:"city,omitempty"`
	State        string    `json:"state,omitempty"`
	FullName     string    `json:"full_name,omitempty"`
}

// Site .. details of site where game was played
//"site":{"capacity":19599,"surface":"Hardwood","name":"Chesapeake Energy Arena","city":"Oklahoma City",
//"state":"Oklahoma"}
type Site struct {
	Extracted    time.Time `json:"extract_time"`
	ExtractedSrc string    `json:"extract_src"`
	SiteID       string    `json:"site_id"`
	Capacity     int       `json:"capacity,omitempty"`
	Surface      string    `json:"surface,omitempty"`
	Name         string    `json:"name,omitempty"`
	City         string    `json:"city,omitempty"`
	State        string    `json:"state,omitempty"`
}

// Roster ... current roster of team
type Roster struct {
	Extracted    time.Time `json:"extract_time"`
	ExtractedSrc string    `json:"extract_src"`
	Team         Team      `json:"team"`
	Member       []Player  `json:"players"`
}

// Player ... Player, part of roster and nbadraft
type Player struct {
	Extracted       time.Time `json:"extract_time"`
	ExtractedSrc    string    `json:"extract_src"`
	UPID            string    `json:"universal_player_id"`
	LName           string    `json:"last_name"`
	FNAME           string    `json:"first_name"`
	DisplayName     string    `json:"display_name,omitempty"`
	BirthDate       string    `json:"birthdate,omitempty"`
	Age             int       `json:"age,omitempty"`
	BirthPlace      string    `json:"birthplace,omitempty"`
	HeightIN        int       `json:"height_in,omitempty"`
	HeightCM        float32   `json:"height_cm,omitempty"`
	HeightM         float32   `json:"height_m,omitempty"`
	HeightFormatted string    `json:"height_formatted,omitempty"`
	WeightLB        int       `json:"weight_lb,omitempty"`
	WeightKG        float32   `json:"weight_kg,omitempty"`
	Position        string    `json:"position,omitempty"`
	UniNumber       int       `json:"uniform_number,omitempty"`
	Bats            string    `json:"bats,omitempty"`
	Throws          string    `json:"throws,omitempty"`
	RosterStatus    string    `json:"roster_status,omitempty"`
}

func getRequest(url string) (*http.Request, error) {
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return req, err
	}
	req.Header.Add("Accept-Encoding", "gzip")
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

func doGet(baseurl string, query string) (io.Reader, error) {
	req, err := getRequest(baseurl + query)
	if err != nil {
		fmt.Printf("The HTTP request header building failed with error %s\n", err)
		return nil, err
	}

	// Send req using http Client
	client := &http.Client{}
	fmt.Println("doing HTTP GET")
	resp, err := client.Do(req)

	// Ensure we close the response body in the event of a non-nil resp
	if resp != nil {
		defer resp.Body.Close()
	}
	// received an error on the HTTP request
	if err != nil {
		defer resp.Body.Close()
		fmt.Printf("The HTTP request header building failed with error %s\n", err)
		return nil, err
	} else if resp.StatusCode != 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("The HTTP request header building failed with error %s : %s\n", resp.Status, data)
		return nil, errors.New(string(data))
	} else if resp.Header.Get("Content-Encoding") == "gzip" {
		fmt.Println("parsing HTTP GZIP-response")
		resp.Header.Del("Content-Length")
		zr, err := gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Printf("Error in gzip response decoding %s\n", err)
			return nil, err
		}
		resp.Body = gzreadCloser{zr, resp.Body}
		return resp.Body, nil
	}
	return resp.Body, nil

}

//decodeEvents .... decode into standard Event structure
//unmarshall side effect of time.Time not parsing RFC3339
//The fmt.Println invokes the Time's .String() function that returns the time in the following format:
//"2006-01-02 15:04:05.999999999 -0700 MST"
//Which as you see contains both the timezone offset and the timezone name.
//In your our case there is no timezone name known for the time, so it outputs the offset twice.
// if we are using XmlstatsTime, we need to convert to stringify fmt.Println(time.Time(events.EventsDate))
func decodeEvents(body io.Reader) (*Events, error) {
	var ev Events
	// Decode the response into our Events struct
	if err := json.NewDecoder(body).Decode(&ev); err != nil {
		return nil, err
	}
	//setup provenance
	ev.Extracted = extractTime
	ev.ExtractedSrc = extractSrc
	//TODO: dig deep into nested structure to set time/src
	return &ev, nil
}

func decodeRoster(body io.Reader) (*Roster, error) {
	var roster Roster
	// Decode the response into our Events struct
	if err := json.NewDecoder(body).Decode(&roster); err != nil {
		return nil, err
	}
	//setup provenance
	roster.Extracted = extractTime
	roster.ExtractedSrc = extractSrc
	//TODO: dig deep into nested structure to set time/src
	//TODO: define UPID - universal playerID logic
	return &roster, nil

}

type gzreadCloser struct {
	*gzip.Reader
	io.Closer
}

func (gz gzreadCloser) Close() error {
	return gz.Closer.Close()
}

func main() {
	fmt.Println("Loading environment configuration constants")

	//construct the event header
	baseurl, exists := os.LookupEnv("XMLSTATS_URL")
	if !exists {
		fmt.Println("XMLSTATS_URL not found, should include your website or email credential")
		baseurl = "https://erikberg.com/"
	}
	// test fetching BoxScores
	body, err := doGet(baseurl, "events.json?date=20130131&sport=nba")
	if (err) != nil {
		fmt.Printf("error caught: %s", err)
		return
	}
	events, _ := decodeEvents(body)
	fmt.Printf(fmt.Sprintf("Events: %#v\n", events))

	// test fetching roster
	rbody, rerr := doGet(baseurl, "nba/roster/oklahoma-city-thunder.json")
	if (rerr) != nil {
		fmt.Printf("error caught: %s", rerr)
		return
	}
	roster, _ := decodeRoster(rbody)
	fmt.Printf(fmt.Sprintf("Roster: %#v\n", roster))
	//
	fmt.Println("Terminating the application normally...")
}
