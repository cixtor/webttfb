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
	messages []error
	servers  map[string]string
	results  []Result
	sync.Mutex
}

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

type ByFilter []Result

func (a ByFilter) Len() int               { return len(a) }
func (a ByFilter) Swap(i int, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFilter) Less(i int, j int) bool { return a[i].Filter < a[j].Filter }

func NewTTFB() (*TTFB, error) {
	var tester TTFB

	tester.servers = make(map[string]string, 0)

	if err := tester.loadServers(); err != nil {
		return nil, err
	}

	return &tester, nil
}

func (t *TTFB) loadServers() error {
	file, err := os.Open(config)

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
		if _, ok := t.servers[unique]; !ok {
			t.servers[unique] = name
		}
	}

	if len(t.servers) == 0 {
		return errors.New("Testing server list is empty")
	}

	return nil
}

func (t *TTFB) data(unique string) string {
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

func (t *TTFB) parseResponse(res io.Reader) (Result, error) {
	var data Result

	if err := json.NewDecoder(res).Decode(&data); err != nil {
		return Result{}, err
	}

	return data, nil
}

func (t *TTFB) serverCheck(wg *sync.WaitGroup, unique string) error {
	defer wg.Done()

	client := &http.Client{}
	body := bytes.NewBufferString(t.data(unique))
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

	data, err := t.parseResponse(&buf)

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

	t.results = append(t.results, data)

	t.Unlock()

	return nil
}

func (t *TTFB) Report(sorting string) []Result {
	var oldval string

	for idx, data := range t.results {
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
			t.messages = append(t.messages, err)
			continue
		}

		t.results[idx].Filter = num
	}

	sort.Sort(ByFilter(t.results))

	return t.results
}

func (t *TTFB) Messages() []error {
	return t.messages
}

func (t *TTFB) Analyze(domain string, private bool) {
	if domain == "" {
		t.messages = append(t.messages, errors.New("Domain is invalid"))
		return
	}

	var wg sync.WaitGroup

	t.Domain = domain   /* track domain name */
	t.Private = private /* hide results from public */

	wg.Add(len(t.servers))

	for unique := range t.servers {
		go func(wg *sync.WaitGroup, unique string) {
			if err := t.serverCheck(wg, unique); err != nil {
				t.messages = append(t.messages, err)
			}
		}(&wg, unique)
	}

	wg.Wait()
}
