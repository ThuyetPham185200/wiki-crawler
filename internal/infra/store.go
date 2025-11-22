package infra

import (
	"fmt"
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
	if !db.SearchTable(w.TitlesTable.TableName) {
		fmt.Printf("%s NOT EXIST - CREATION PROCESS STARTING\n", w.TitlesTable.TableName)
		w.TitlesTable.CreateTable()
	} else {
		fmt.Printf("%s EXISTED\n", w.TitlesTable.TableName)
	}

	if !db.SearchTable(w.PairsTable.TableName) {
		fmt.Printf("%s NOT EXIST - CREATION PROCESS STARTING\n", w.PairsTable.TableName)
		w.PairsTable.CreateTable()
	} else {
		fmt.Printf("%s EXISTED\n", w.PairsTable.TableName)
	}

	w.RedisClient = redisclient.InitSingleton(rcfg)
	w.TitlsToQueryQ = make(chan model.TitleQuery, TitlsToQueryQCap)

	if titles, err := file.ReadTextFile(datapath); err == nil {
		fmt.Println(titles)
		for _, e := range titles {
			w.TitlsToQueryQ <- model.TitleQuery{
				Title: e,
			}
		}
	} else {
		fmt.Printf("[WikiStore] Failed to read seed names from %s: %v\n", datapath, err)
	}

	w.RawDataQ = make(chan model.RawDataWiki, RawDataQCap)
	return w
}
