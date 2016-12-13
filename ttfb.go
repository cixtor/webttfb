package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sort"
)

// TTFB holds the list of testing servers and the
type TTFB struct {
	Domain   string
	Private  bool
	Messages []error
	Servers  map[string]string
	Results  []Result
}

// Result holds the information of each test case.
type Result struct {
	Status         int     `json:"status"`
	Action         string  `json:"action"`
	Message        string  `json:"message"`
	ResetLastTest  bool    `json:"reset_last_test"`
	DataFromCache  bool    `json:"data_from_cache"`
	LastTestTime   int     `json:"last_test_time"`
	LocationsCount int     `json:"_locations_count"`
	TestedServers  int     `json:"_tested_servers"`
	IsLastTest     bool    `json:"_is_last_test"`
	Output         Info    `json:"output"`
	Filter         float64 `json:"-"`
}

// Info holds the data of each test case.
type Info struct {
	Domain          string  `json:"domain"`
	IP              string  `json:"ip"`
	ConnectTime     float64 `json:"connect_time,string"`
	FirstbyteTime   float64 `json:"firstbyte_time,string"`
	TotalTime       float64 `json:"total_time,string"`
	DomainID        string  `json:"domain_id"`
	DomainUnique    string  `json:"domain_unique"`
	ServerID        string  `json:"server_id"`
	ServerAbbr      string  `json:"server_abbr"`
	ServerTitle     string  `json:"server_title"`
	ServerFlagImage string  `json:"server_flag_image"`
	DomainAndIP     string  `json:"domain_and_ip"`
	RequestTime     int64   `json:"request_time"`
	ServerLocation  string  `json:"server_location"`
	ServerLatitude  float64 `json:"server_latitude,string"`
	ServerLongitude float64 `json:"server_longitude,string"`
}

// ByFilter implements sort.Interface to allow data sorting.
type ByFilter []Result

func (a ByFilter) Len() int               { return len(a) }
func (a ByFilter) Swap(i int, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFilter) Less(i int, j int) bool { return a[i].Filter < a[j].Filter }

// NewTTFB returns a new pointer to the TTFB interface.
func NewTTFB(domain string, private bool) (*TTFB, error) {
	var tester TTFB

	if domain == "" {
		return nil, errors.New("Domain is invalid")
	}

	tester.Domain = domain   /* track domain name */
	tester.Private = private /* hide results from public */
	tester.Servers = make(map[string]string, 0)

	if err := tester.LoadServers(); err != nil {
		return nil, err
	}

	return &tester, nil
}

// LoadServers reads and loads the content of the configuration file.
func (t *TTFB) LoadServers() error {
	file, err := os.Open(os.Getenv("HOME") + "/" + config)

	if err != nil {
		return err
	}

	defer file.Close()

	var line string
	var name string
	var unique string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line = scanner.Text()

		if len(line) < 10 {
			continue
		}

		unique = line[0:7]
		name = line[9:]

		// Skip servers without name.
		if name == "" {
			continue
		}

		// Append non-duplicated servers to the list.
		if _, ok := t.Servers[unique]; !ok {
			t.Servers[unique] = name
		}
	}

	if len(t.Servers) == 0 {
		return errors.New("Testing server list is empty")
	}

	return nil
}

// FormData builds the HTTP query object with the necessary parameters for each
// test. A basic test request requires the domain name and the unique identifier
// for the testing server that will be used to run the test in itself.
// Additionally, if the user decides to hide the results of the test from the
// public the program will append another request parameter to force the API
// service to consider the test as private.
func (t *TTFB) FormData(unique string) string {
	form := url.Values{}

	form.Add("load_time_tester", "1")
	form.Add("form_action", "test_load_time")
	form.Add("location", unique)
	form.Add("domain", t.Domain)

	if t.Private {
		form.Add("is_private", "true")
	}

	return form.Encode()
}

// ParseResponse reads and decodes the JSON-encoded data from the API service.
func (t *TTFB) ParseResponse(res io.Reader) (Result, error) {
	var data Result

	if err := json.NewDecoder(res).Decode(&data); err != nil {
		return Result{}, err
	}

	return data, nil
}

// ServerCheck sends the HTTP request to the API service.
func (t *TTFB) ServerCheck(ch chan Result, unique string) error {
	client := &http.Client{}
	body := bytes.NewBufferString(t.FormData(unique))
	req, err := http.NewRequest("POST", service, body)

	if err != nil {
		ch <- Result{}
		return err
	}

	req.Header.Set("accept-language", "en-US,en;q=0.8")
	req.Header.Set("accept-encoding", "gzip, deflate, br")
	req.Header.Set("user-agent", "Mozilla/5.0 (KHTML, like Gecko) Safari/537.36")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("referer", "https://performance.sucuri.net/")
	req.Header.Set("origin", "https://performance.sucuri.net")
	req.Header.Set("authority", "performance.sucuri.net")
	req.Header.Set("x-requested-with", "XMLHttpRequest")

	res, err := client.Do(req)

	if err != nil {
		ch <- Result{}
		return err
	}

	defer res.Body.Close()

	var buf bytes.Buffer
	(&buf).ReadFrom(res.Body)

	data, err := t.ParseResponse(&buf)

	if err != nil {
		ch <- Result{}
		return err
	}

	ch <- data
	return nil
}

// Report takes the data generated after the execution of all the HTTP requests
// and sorts all the values by a specific field in the JSON-encoded object.
// Currently the program allows sorting by the status of the test, failed tests
// are listed at the end of the report. The program also allows to sort by the
// connection time, the time to first byte and the total time, these values are
// returned as strings and the program parses and converts them to floating
// points for accessibility.
func (t *TTFB) Report(sorting string) []Result {
	var oldval float64

	for idx, data := range t.Results {
		switch sorting {
		case connectionTime:
			oldval = data.Output.ConnectTime
		case timeToFirstByte:
			oldval = data.Output.FirstbyteTime
		case totalTime:
			oldval = data.Output.TotalTime
		default:
			// If the HTTP request status is equal to the integer one we
			// consider it a successful operation and a failure otherwise. Since
			// the sort interface works with a less-than comparison by default
			// we have to invert the values and make the non-positive status
			// greater than the expected number so the entries associated with a
			// failure are displayed at the end of the table.
			oldval = 2.0
			if data.Status == 1 {
				oldval = 1.0
			}
		}

		// Increase filter if the HTTP request failed.
		if data.Status == 0 {
			oldval = 360 + oldval
		}

		t.Results[idx].Filter = oldval
	}

	sort.Sort(ByFilter(t.Results))

	return t.Results
}

// ErrorMessages returns an array of errors for any failure occurred during the
// execution of the HTTP requests. No error message will be reported when the
// goroutines are locked, they will all be merged into one big pile of data and
// then printed at the end of all the operations (reading, parsing, sorting,
// etc).
func (t *TTFB) ErrorMessages() []error {
	return t.Messages
}

// Analyze sends a HTTP GET request to the external API service for each testing
// server found in the configuration file. Each testing server is supposed to
// return a JSON-encoded object with information that describes the speed of the
// website from different locations in the world.
func (t *TTFB) Analyze() {
	var done int
	total := len(t.Servers)
	ch := make(chan Result, total)

	for unique := range t.Servers {
		go func(ch chan Result, unique string) {
			if err := t.ServerCheck(ch, unique); err != nil {
				t.Messages = append(t.Messages, err)
			}
		}(ch, unique)
	}

	for idx := 0; idx < total; idx++ {
		done++
		data := <-ch
		// Print a loading message until finished.
		fmt.Printf("\rTesting %02d/%d ...", done, total)
		t.Results = append(t.Results, data)
	}

	fmt.Print("\r") /* reset previous line */
}

// LocalTest leverages the power of CURL to execute a simple speed test against
// the specified domain name, the test consists of a single HTTP GET request
// from the current internet connection and reports the connection time, the
// time to the first byte, the total transmission time among other things.
func (t *TTFB) LocalTest() ([]byte, error) {
	var stats string

	stats += "  Connection:  %{time_connect} secs\n"
	stats += "  FirstByte:   %{time_starttransfer} secs\n"
	stats += "  TotalTime:   %{time_total} secs\n"
	stats += "  NameLookup:  %{time_namelookup} secs\n"
	stats += "  Redirection: %{time_redirect} secs\n"

	out, err := exec.Command("/usr/bin/env",
		"curl", "-L", "-s", "-o", "/dev/null",
		"-w", stats, *domain).CombinedOutput()

	if err != nil {
		return nil, err
	}

	return out, nil
}

// Average measures the average responsiveness of each test case ignoring the
// highest and lowest value to increase the accuracy of the total number. Notice
// that if the number of successful HTTP requests is lower than 3 it means we
// cannot use any value because after the removal of the highest and lowest we
// will be left with nothing so we return zero.
func (t *TTFB) Average(group string) float64 {
	var total float64
	var values []float64

	for _, data := range t.Results {
		if group == connectionTime {
			values = append(values, data.Output.ConnectTime)
			continue
		}

		if group == timeToFirstByte {
			values = append(values, data.Output.FirstbyteTime)
			continue
		}

		if group == totalTime {
			values = append(values, data.Output.TotalTime)
			continue
		}
	}

	// There is no enough data to average.
	if len(values) < 3 {
		return 0.0
	}

	sort.Float64s(values)

	// Drop first and last values for accuracy.
	for i := 1; i < len(values)-1; i++ {
		total += values[i]
	}

	return total / float64(len(values)-2)
}
