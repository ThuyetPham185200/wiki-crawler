package tables

import dbclient "wikicrawler/internal/infra/postgresclient"

// TitlesTable kế thừa BaseTable
type TitlesTable struct {
	dbclient.BaseTable
}

// NewTitlesTable khởi tạo table entities
func NewTitlesTable(client *dbclient.PostgresClient) *TitlesTable {
	return &TitlesTable{
		BaseTable: dbclient.BaseTable{
			Client:    client,
			TableName: "titles",
			Columns: map[string]string{
				"title_id":   "UUID PRIMARY KEY",
				"name":       "VARCHAR(255) NOT NULL",
				"created_at": "TIMESTAMP NOT NULL DEFAULT now()",
				"updated_at": "TIMESTAMP NOT NULL DEFAULT now()",
			},
			Constraints: []string{
				"UNIQUE (name)",
				"CREATE INDEX IF NOT EXISTS idx_titles_name ON titles (name)",
			},
		},
	}
}
