package dbclient

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type PostGresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBname   string
}

type PostgresClient struct {
	DB *sql.DB
}

func NewPostgresClient(config PostGresConfig) *PostgresClient {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Không thể mở kết nối DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Không thể ping DB: %v", err)
	}

	log.Println("✅ Kết nối PostgreSQL thành công!")
	return &PostgresClient{DB: db}
}

func (pc *PostgresClient) Close() {
	if pc.DB != nil {
		pc.DB.Close()
	}
}

func (pc *PostgresClient) SearchTable(tb string) bool {
	if pc.DB == nil {
		log.Println("❌ Database connection is not initialized")
		return false
	}

	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`
	err := pc.DB.QueryRow(query, tb).Scan(&exists) // tb (kiểu string) sẽ được gán vào chỗ $1 trong câu SQL.
	if err != nil {
		log.Printf("❌ Error checking table existence: %v", err)
		return false
	}

	return exists
}
