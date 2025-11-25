package api

import (
	"fmt"
	"net/url"
	"strings"
)

type WiKiAPI struct {
	baseURL string
}

func NewWikiAPI() *WiKiAPI {
	return &WiKiAPI{
		baseURL: "https://vi.wikipedia.org/w/api.php?action=query&prop=links&format=json&plnamespace=0&pllimit=max&titles=%s",
	}
}

// URLwithTitle returns a properly encoded URL for a given title
func (w *WiKiAPI) URLwithTitle(title string) string {
	// Replace spaces with '+' instead of using QueryEscape
	safeTitle := strings.ReplaceAll(title, " ", "+")
	requestURL := fmt.Sprintf(w.baseURL, safeTitle)
	return requestURL
}

// URLwithTitleAndPLcontinue returns a properly encoded URL with pagination
func (w *WiKiAPI) URLwithTitleAndPLcontinue(title string, plcontinue string) string {
	safeTitle := strings.ReplaceAll(title, " ", "+")
	requestURL := fmt.Sprintf(w.baseURL, safeTitle)
	requestURL += "&plcontinue=" + url.QueryEscape(plcontinue) // plcontinue may contain special characters
	return requestURL
}
