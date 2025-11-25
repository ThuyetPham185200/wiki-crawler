package infra

import (
	"fmt"
	"log"
	dbclient "wikicrawler/internal/infra/postgresclient"
	"wikicrawler/internal/infra/postgresclient/tables"
	"wikicrawler/internal/infra/redisclient"
	"wikicrawler/internal/model"
	"wikicrawler/internal/utils/file"

	"github.com/google/uuid"
)

type WikiStore struct {
	TitlsToQueryQ chan model.TitleQuery
	RawDataQ      chan model.RawDataWiki
	TitlesMap     map[string]bool
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
	w.TitlesMap = make(map[string]bool)
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

	if titles, err := file.ReadJsonArrayFile(datapath); err == nil {
		//fmt.Println(titles)
		for _, e := range titles {
			m, ok := e.(map[string]any) // ép kiểu
			if !ok {
				fmt.Println("invalid item:", e)
				continue
			}

			t, ok := m["wiki_title"].(string)
			if !ok {
				fmt.Println("missing wiki_title:", m)
				continue
			}

			id := uuid.New().String()
			if err := w.TitlesTable.Insert(map[string]interface{}{
				"title_id": id,
				"name":     t,
			}); err != nil {
				log.Printf("[RawDataHandler] Failed to insert title %s: %v", t, err)
			} else {
				if _, ok := w.TitlesMap[t]; !ok {
					w.TitlesMap[t] = true
				}

				w.TitlsToQueryQ <- model.TitleQuery{
					Title: t,
				}
				log.Printf("[RawDataHandler] Success to insert title '%s' to titles", t)
			}
		}
	} else {
		fmt.Printf("[WikiStore] Failed to read seed names from %s: %v\n", datapath, err)
	}
	fmt.Println(w.TitlesMap)
	w.RawDataQ = make(chan model.RawDataWiki, RawDataQCap)
	return w
}
