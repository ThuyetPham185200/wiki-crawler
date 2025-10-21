package api

import (
	"fmt"
	"net/url"
)

type WiKiAPI struct {
	baseURL string
}

func NewWikiAPI() *WiKiAPI {
	return &WiKiAPI{
		baseURL: "https://en.wikipedia.org/w/api.php?action=query&prop=links&format=json&plnamespace=0&pllimit=max&titles=%s",
	}
}

func (w *WiKiAPI) URLwithTitle(title string) string {
	encodedTitle := url.QueryEscape(title)
	requestURL := fmt.Sprintf(w.baseURL, encodedTitle)
	return requestURL
}

func (w *WiKiAPI) URLwithTitleAndPLcontinue(title string, plcontinue string) string {
	encodedTitle := url.QueryEscape(title)
	requestURL := fmt.Sprintf(w.baseURL, encodedTitle)
	requestURL += "&plcontinue=" + url.QueryEscape(plcontinue)
	return requestURL
}
