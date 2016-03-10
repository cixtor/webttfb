package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"sync"
)

const zeroos string = "0.000"
const config string = "servers.cfg"
const service string = "https://performance.sucuri.net/index.php?ajaxcall"

// TTFB holds the list of testing servers and the
type TTFB struct {
	Domain   string
	Private  bool
	Messages []error
	Servers  map[string]string
	Results  []Result
	sync.Mutex
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
	Domain          string `json:"domain"`
	IP              string `json:"ip"`
	ConnectTime     string `json:"connect_time"`
	FirstbyteTime   string `json:"firstbyte_time"`
	TotalTime       string `json:"total_time"`
	DomainID        string `json:"domain_id"`
	DomainUnique    string `json:"domain_unique"`
	ServerID        string `json:"server_id"`
	ServerAbbr      string `json:"server_abbr"`
	ServerTitle     string `json:"server_title"`
	ServerFlagImage string `json:"server_flag_image"`
	DomainAndIP     string `json:"domain_and_ip"`
	RequestTime     int64  `json:"request_time"`
	ServerLocation  string `json:"server_location"`
	ServerLatitude  string `json:"server_latitude"`
	ServerLongitude string `json:"server_longitude"`
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
func (t *TTFB) ServerCheck(wg *sync.WaitGroup, unique string) error {
	defer wg.Done()

	client := &http.Client{}
	body := bytes.NewBufferString(t.FormData(unique))
	req, err := http.NewRequest("POST", service, body)

	if err != nil {
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
		return err
	}

	defer res.Body.Close()

	var buf bytes.Buffer
	(&buf).ReadFrom(res.Body)

	data, err := t.ParseResponse(&buf)

	if err != nil {
		return err
	}

	t.Lock()

	if data.Output.ConnectTime == "" {
		data.Output.ConnectTime = zeroos
	}

	if data.Output.FirstbyteTime == "" {
		data.Output.FirstbyteTime = zeroos
	}

	if data.Output.TotalTime == "" {
		data.Output.TotalTime = zeroos
	}

	t.Results = append(t.Results, data)

	t.Unlock()

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
	var oldval string

	for idx, data := range t.Results {
		switch sorting {
		case "conn":
			oldval = data.Output.ConnectTime
		case "ttfb":
			oldval = data.Output.FirstbyteTime
		case "ttl":
			oldval = data.Output.TotalTime
		default:
			oldval = "2.0"
			if data.Status == 1 {
				oldval = "1.0"
			}
		}

		// Increase filter if the HTTP request failed.
		if data.Status == 0 {
			oldval = "360" + oldval
		}

		num, err := strconv.ParseFloat(oldval, 64)

		if err != nil {
			t.Messages = append(t.Messages, err)
			continue
		}

		t.Results[idx].Filter = num
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
	var wg sync.WaitGroup

	wg.Add(len(t.Servers))

	for unique := range t.Servers {
		go func(wg *sync.WaitGroup, unique string) {
			if err := t.ServerCheck(wg, unique); err != nil {
				t.Messages = append(t.Messages, err)
			}
		}(&wg, unique)
	}

	wg.Wait()
}
