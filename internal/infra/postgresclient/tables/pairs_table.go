package tables

import (
	dbclient "wikicrawler/internal/infra/postgresclient"
)

// PairsTable kế thừa BaseTable
type PairsTable struct {
	dbclient.BaseTable
}

// NewPairsTable khởi tạo table pairs
func NewPairsTable(client *dbclient.PostgresClient) *PairsTable {
	return &PairsTable{
		BaseTable: dbclient.BaseTable{
			Client:    client,
			TableName: "pairs",
			Columns: map[string]string{
				"pair_id":    "UUID PRIMARY KEY",
				"title_src":  "UUID NOT NULL",
				"title_dst":  "UUID NOT NULL",
				"created_at": "TIMESTAMP NOT NULL DEFAULT now()",
				"updated_at": "TIMESTAMP NOT NULL DEFAULT now()",
			},
			Constraints: []string{
				"FOREIGN KEY (title_src) REFERENCES titles(title_id)",
				"FOREIGN KEY (title_dst) REFERENCES titles(title_id)",
				"CONSTRAINT uniq_src_dst UNIQUE (title_src, title_dst)",

				"CREATE INDEX IF NOT EXISTS idx_pairs_src_dst ON pairs (title_src, title_dst)",
			},
		},
	}
}
