package main

import (
	"fmt"
	"log"
	dbclient "wikicrawler/internal/infra/postgresclient"
	"wikicrawler/internal/infra/postgresclient/tables"
)

func main() {
	client := dbclient.NewPostgresClient(dbclient.PostGresConfig{
		Host:     "localhost", // IP
		Port:     "5432",      // Port
		User:     "taopq",     // user_name
		Password: "123456a@",  // password
		DBname:   "mydb",      // db
	})
	defer client.Close()
	////////////////////////////////////////////////////////////////////////////////////
	// Tạo bảng Titles
	titlesTable := tables.NewTitlesTable(client)

	if !client.SearchTable(titlesTable.TableName) {
		fmt.Printf("%s NOT EXIST - CREATION PROCESS STARTING\n", titlesTable.TableName)
		titlesTable.CreateTable()
	} else {
		fmt.Printf("%s EXISTED\n", titlesTable.TableName)
	}

	// Lấy tất cả Domains
	rows, err := titlesTable.GetAll()
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range rows {
		fmt.Println(row)
	}
	////////////////////////////////////////////////////////////////////////////////////
	// Tạo bảng Pair Type
	pairsTable := tables.NewPairsTable(client)

	if !client.SearchTable(pairsTable.TableName) {
		fmt.Printf("%s NOT EXIST - CREATION PROCESS STARTING\n", pairsTable.TableName)
		pairsTable.CreateTable()
	} else {
		fmt.Printf("%s EXISTED\n", pairsTable.TableName)
	}

	// Lấy tất cả Entity Type
	rows, err = pairsTable.GetAll()
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range rows {
		fmt.Println(row)
	}
}
