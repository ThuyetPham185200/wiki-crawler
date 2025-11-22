package app

import (
	"fmt"
	"time"
	"wikicrawler/internal/core/apiclient"
	"wikicrawler/internal/core/cdcmanager"
	"wikicrawler/internal/core/cdcmanager/cdc"
	"wikicrawler/internal/core/kafkaclient"
	"wikicrawler/internal/core/rawdatahandler"
	"wikicrawler/internal/infra"
	dbclient "wikicrawler/internal/infra/postgresclient"
	"wikicrawler/internal/infra/redisclient"
)

type App struct {
	apiclient   *apiclient.APIClient
	cdc         *cdcmanager.CDCManager
	datahandler *rawdatahandler.RawDataHandler
}

func NewWikiCrawlerApp() *App {
	var app = &App{}
	app.init()
	return app
}

func (a *App) Start() {
	if err := a.cdc.Open(); err != nil {
		fmt.Printf("[WikiCrawlerApp] Failed to open CDC: %v\n", err)
		return
	} else {
		fmt.Printf("[WikiCrawlerApp] CDC opened successfully\n")
	}
	time.Sleep(2 * time.Second)

	if err := a.apiclient.Start(); err != nil {
		fmt.Printf("[WikiCrawlerApp] Failed to start apiclient: %v\n", err)
		return
	} else {
		fmt.Printf("[WikiCrawlerApp] apiclient started successfully\n")
	}

	time.Sleep(2 * time.Second)
	if err := a.datahandler.Start(); err != nil {
		fmt.Printf("[WikiCrawlerApp] Failed to start rawdatahandler: %v\n", err)
		return
	} else {
		fmt.Printf("[WikiCrawlerApp] rawdatahandler started successfully\n")
	}

}

func (a *App) Stop() {
	if err := a.apiclient.Stop(); err != nil {
		fmt.Printf("[WikiCrawlerApp] Failed to stop apiclient: %v\n", err)
	}
	if err := a.cdc.Close(); err != nil {
		fmt.Printf("[WikiCrawlerApp] Failed to close CDC: %v\n", err)
	}
	if err := a.datahandler.Stop(); err != nil {
		fmt.Printf("[WikiCrawlerApp] Failed to stop rawdatahandler: %v\n", err)
	}
}

// ///////////////////////////////////////////////////////////////////////////////////////
func (a *App) init() {
	datapath := "./data/seed_names.txt"
	cfg := dbclient.PostGresConfig{
		Host:     "localhost", // IP
		Port:     "5432",      // Port
		User:     "erduser",   // user_name
		Password: "erdp@ss",   // password
		DBname:   "wikidb",    // db
	}
	rcfg := redisclient.RedisConfig{
		Addr:     "localhost:6379", // redis server host
		Password: "",
		DB:       2, // Logical database index (Redis has 16 by default: 0–15).
	}

	RawDataQCap := 1000
	TitlsToQueryQCap := 1000
	store := infra.NewWikiStore(datapath, cfg, rcfg, RawDataQCap, TitlsToQueryQCap)
	a.apiclient = apiclient.NewAPIClient(store, 30*time.Second, 90*time.Second, 10000, 10)

	cfgcd := &cdc.CDCConfig{
		Replicator:       "replicator",
		Psw:              "erdp@ss",
		Address:          "localhost:5432",
		DB:               "wikidb",
		Replication_slot: "wikidb_slot",
		OutputPlugin:     "pgoutput",
		Publication:      "wikidb_pub",
	}
	cfgrelpro := kafkaclient.KafkaConfig{
		BootstrapServers: "localhost:19092",
		Topic:            store.PairsTable.TableName,
	}
	cfgentpro := kafkaclient.KafkaConfig{
		BootstrapServers: "localhost:19092",
		Topic:            store.TitlesTable.TableName,
	}
	relTBname := store.PairsTable.TableName
	entTBname := store.TitlesTable.TableName
	a.cdc = cdcmanager.NewCDCManager(cfgcd, cfgrelpro, cfgentpro, relTBname, entTBname)

	a.datahandler = rawdatahandler.NewRawDataHandler(store, 10, 100)

	fmt.Printf("[WikiCrawlerApp] done to init all components!\n")
}
