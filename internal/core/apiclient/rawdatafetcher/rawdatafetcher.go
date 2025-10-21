package rawdatafetcher

import (
	"fmt"
	"net/http"
	"time"
)

type RawDataFetcher struct {
	httpClient *http.Client
}

func NewRawDataFetcher(Timeout, IdleConnTimeout time.Duration, MaxIdleConns, MaxIdleConnsPerHost int) *RawDataFetcher {
	r := &RawDataFetcher{}
	r.httpClient = &http.Client{
		Timeout: Timeout, // Prevent hanging requests
		Transport: &http.Transport{
			MaxIdleConns:        MaxIdleConns,        // Total number of idle connections to keep
			MaxIdleConnsPerHost: MaxIdleConnsPerHost, // Per-host idle connections
			IdleConnTimeout:     IdleConnTimeout,     // How long to keep an idle conn alive
		},
	}
	return r
}

func (r *RawDataFetcher) GetRawData(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("[RawDataFetcher] failed to create request: %w", err)
	}

	// Setting User-Agent to be respectful to Wikipedia
	req.Header.Set("User-Agent", "Entities-Relationship/1.0 (Personal Project)")

	res, err := r.httpClient.Do(req)

	return res, err
}
