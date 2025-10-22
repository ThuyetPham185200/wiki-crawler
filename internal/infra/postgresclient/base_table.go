package dbclient

import (
	"fmt"
	"log"
	"strings"
)

type BaseTable struct {
	Client      *PostgresClient
	TableName   string
	Columns     map[string]string // column_name -> type (VD: "id": "SERIAL PRIMARY KEY")
	Constraints []string          // danh sách constraint ở mức table (FOREIGN KEY, UNIQUE, CHECK, ...)
}

// CreateTable tạo bảng dựa trên metadata
func (bt *BaseTable) CreateTable() {
	var cols []string
	for col, typ := range bt.Columns {
		cols = append(cols, fmt.Sprintf("%s %s", col, typ))
	}

	allDefs := cols
	if len(bt.Constraints) > 0 {
		allDefs = append(allDefs, bt.Constraints...)
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s)`,
		bt.TableName,
		strings.Join(allDefs, ", "),
	)

	_, err := bt.Client.DB.Exec(query)
	if err != nil {
		log.Fatalf("❌ Lỗi tạo bảng %s: %v", bt.TableName, err)
	}
	log.Printf("✅ Bảng %s sẵn sàng.", bt.TableName)
}

// Insert thêm dữ liệu vào bảng
func (bt *BaseTable) Insert(values map[string]interface{}) error {
	cols := []string{}
	vals := []interface{}{}
	placeholders := []string{}

	i := 1
	for col, val := range values {
		cols = append(cols, col)
		vals = append(vals, val)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`,
		bt.TableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	_, err := bt.Client.DB.Exec(query, vals...)
	if err != nil {
		log.Printf("❌ Lỗi insert vào %s: %v", bt.TableName, err)
	} else {
		log.Printf("✅ Insert thành công vào %s", bt.TableName)
	}
	return err
}

// GetAll lấy tất cả dữ liệu trong table và trả về []map[string]interface{}
func (bt *BaseTable) GetAll() ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`SELECT * FROM %s`, bt.TableName)
	rows, err := bt.Client.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("❌ lỗi query %s: %w", bt.TableName, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Chuẩn bị mảng giá trị
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}

		// Scan vào valuePtrs
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Đưa vào map
		rowData := make(map[string]interface{})
		for i, col := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowData[col] = string(b)
			} else {
				rowData[col] = val
			}
		}
		results = append(results, rowData)
	}

	return results, nil
}
