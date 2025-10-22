package infra

import (
	"wikicrawler/internal/model"
	"wikicrawler/internal/utils/file"
)

type WikiStore struct {
	TitlsToQueryQ   chan string
	FilteredTitlesQ chan string
	FilteredPairsQ  chan model.TitlesPair
	RawDataQ        chan model.WikiLinksResponse
}

func NewWikiStore(datapath string) *WikiStore {
	w := &WikiStore{}
	w.TitlsToQueryQ = make(chan string)
	if titles, err := file.ReadTextFile(datapath); err == nil {
		for _, e := range titles {
			w.TitlsToQueryQ <- e
		}
	}

	w.FilteredTitlesQ = make(chan string)
	w.RawDataQ = make(chan model.WikiLinksResponse)
	w.FilteredPairsQ = make(chan model.TitlesPair)
	return w
}
