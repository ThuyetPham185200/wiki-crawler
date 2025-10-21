package apiclient

import "wikicrawler/internal/utils/processor"

type APIClient struct {
	processor.BaseProcessor
}

func NewAPIClient() *APIClient {
	return &APIClient{}
}

func (a *APIClient) RunningTask() {

}
