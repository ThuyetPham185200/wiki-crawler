package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wikicrawler/internal/core/apiclient/api"
	"wikicrawler/internal/core/apiclient/rawdatafetcher"
	"wikicrawler/internal/infra"
	"wikicrawler/internal/model"
	"wikicrawler/internal/utils/processor"
)

type APIClient struct {
	processor.BaseProcessor
	store   *infra.WikiStore
	api     *api.WiKiAPI
	fetcher *rawdatafetcher.RawDataFetcher
}

func NewAPIClient(store *infra.WikiStore, Timeout, IdleConnTimeout time.Duration, MaxIdleConns, MaxIdleConnsPerHost int) *APIClient {
	return &APIClient{
		store:   store,
		api:     api.NewWikiAPI(),
		fetcher: rawdatafetcher.NewRawDataFetcher(Timeout, IdleConnTimeout, MaxIdleConns, MaxIdleConnsPerHost),
	}
}

func (a *APIClient) RunningTask() {
	title := <-a.store.TitlsToQueryQ
	var plcontinue string

	for {
		var requestURL string

		if plcontinue != "" {
			requestURL = a.api.URLwithTitleAndPLcontinue(url.QueryEscape(title.Title), url.QueryEscape(plcontinue))
		} else {
			requestURL = a.api.URLwithTitle(url.QueryEscape(title.Title))
		}

		res, err := a.makeRequestWithRetry(requestURL)
		if err != nil {
			fmt.Printf("[APIClient] Request error for %s: %v\n", title, err)
			return
		}
		defer res.Body.Close()

		if !strings.Contains(res.Header.Get("Content-Type"), "application/json") {
			fmt.Printf("[APIClient] Unexpected content type for %s: %s\n", title, res.Header.Get("Content-Type"))
			return
		}

		var result model.WikiLinksResponse
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			fmt.Printf("[APIClient] Failed to decode JSON for %s: %v\n", title, err)
			return
		}

		a.store.RawDataQ <- model.RawDataWiki{
			TitleQ:   title,
			LinksRes: result,
		}

		if result.Continue.Plcontinue == "" {
			break
		}
		plcontinue = result.Continue.Plcontinue
	}
}

func (a *APIClient) makeRequestWithRetry(url string) (*http.Response, error) {
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		res, err := a.fetcher.GetRawData(url)

		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("[APIClient] request failed after %d attempts: %w", maxRetries, err)
			}
			// Wait before retrying (exponential backoff)
			delay := baseDelay * time.Duration(1<<attempt) // 1s, 2s, 4s
			time.Sleep(delay)
			continue
		}

		// Check for specific HTTP errors that warrant retry
		if res.StatusCode == http.StatusTooManyRequests || // 429 Rate Limited
			res.StatusCode >= 500 { // 5xx Server Errors
			res.Body.Close() // Close before retry

			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("[APIClient] request failed with status %d after %d attempts", res.StatusCode, maxRetries)
			}

			// For rate limiting, wait longer
			delay := baseDelay * time.Duration(1<<attempt)
			if res.StatusCode == http.StatusTooManyRequests {
				delay *= 2 // Double delay for rate limits
			}
			time.Sleep(delay)
			continue
		}

		// Check for non-200 status codes that shouldn't be retried
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			return nil, fmt.Errorf("[APIClient] unexpected status code: %d", res.StatusCode)
		}

		return res, nil
	}

	return nil, fmt.Errorf("[APIClient] unexpected: reached end of retry loop")
}
