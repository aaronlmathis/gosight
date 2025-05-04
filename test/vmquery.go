// cmd/queryvm/main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type VMResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func main() {
	metric := flag.String("metric", "system.cpu.usage_percent", "Metric name to query")
	host := flag.String("host", "http://localhost:8428", "VictoriaMetrics host URL")
	step := flag.Int("step", 15, "Query step in seconds")
	flag.Parse()

	end := time.Now().Unix()
	start := end - 300 // last 5 minutes

	u, _ := url.Parse(*host + "/api/v1/query_range")
	q := u.Query()
	q.Set("query", *metric)
	q.Set("start", fmt.Sprintf("%d", start))
	q.Set("end", fmt.Sprintf("%d", end))
	q.Set("step", fmt.Sprintf("%d", *step))
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "VictoriaMetrics error: %s\n", body)
		os.Exit(1)
	}

	var parsed VMResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(parsed, "", "  ")
	fmt.Println(string(out))
}
