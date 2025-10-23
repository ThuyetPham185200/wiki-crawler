package infra

import (
	dbclient "wikicrawler/internal/infra/postgresclient"
	"wikicrawler/internal/infra/postgresclient/tables"
	"wikicrawler/internal/infra/redisclient"
	"wikicrawler/internal/model"
	"wikicrawler/internal/utils/file"
)

type WikiStore struct {
	TitlsToQueryQ chan model.TitleQuery
	RawDataQ      chan model.RawDataWiki
	DBclient      *dbclient.PostgresClient
	RedisClient   *redisclient.RedisClient
	TitlesTable   *tables.TitlesTable
	PairsTable    *tables.PairsTable
}

func NewWikiStore(datapath string, cfg dbclient.PostGresConfig, rcfg redisclient.RedisConfig, RawDataQCap, TitlsToQueryQCap int) *WikiStore {
	w := &WikiStore{}
	db := dbclient.NewPostgresClient(cfg)
	w.DBclient = db
	w.PairsTable = tables.NewPairsTable(db)
	w.TitlesTable = tables.NewTitlesTable(db)
	w.RedisClient = redisclient.InitSingleton(rcfg)
	w.TitlsToQueryQ = make(chan model.TitleQuery, TitlsToQueryQCap)

	if titles, err := file.ReadTextFile(datapath); err == nil {
		for _, e := range titles {
			w.TitlsToQueryQ <- model.TitleQuery{
				Title: e,
			}
		}
	}

	w.RawDataQ = make(chan model.RawDataWiki, RawDataQCap)
	return w
}
