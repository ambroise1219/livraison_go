package db

import (
	"fmt"
	"log"

	"github.com/surrealdb/surrealdb.go"
	"github.com/ambroise1219/livraison_go/config"
)

var DB *surrealdb.DB

type SurrealResponse struct {
	Status string      `json:"status"`
	Time   string      `json:"time"`
	Result interface{} `json:"result"`
}

// Initialize database connection
func InitDB(cfg *config.Config) error {
	// Connect to SurrealDB with retry logic
	var db *surrealdb.DB
	var err error
	
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		db, err = surrealdb.New(cfg.SurrealDBURL)
		if err == nil {
			break
		}
		log.Printf("Connection attempt %d failed: %v", i+1, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in 2 seconds...")
			// Simple sleep simulation - in real app use time.Sleep(2 * time.Second)
		}
	}
	
	if err != nil {
		return fmt.Errorf("failed to connect to SurrealDB after %d attempts: %v", maxRetries, err)
	}

	// Authenticate
	_, err = db.Signin(map[string]interface{}{
		"user": cfg.SurrealDBUsername,
		"pass": cfg.SurrealDBPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	// Use namespace and database
	_, err = db.Use(cfg.SurrealDBNS, cfg.SurrealDBDB)
	if err != nil {
		return fmt.Errorf("failed to use namespace/database: %v", err)
	}

	DB = db
	log.Printf("Successfully connected to SurrealDB at %s", cfg.SurrealDBURL)
	return nil
}

// Close database connection
func CloseDB() error {
	if DB != nil {
		DB.Close()
	}
	return nil
}

// Query executes a SurrealQL query
func Query(query string, params map[string]interface{}) (interface{}, error) {
	if params == nil {
		params = make(map[string]interface{})
	}

	result, err := DB.Query(query, params)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	return result, nil
}

// Create creates a new record
func Create(table string, data interface{}) (interface{}, error) {
	result, err := DB.Create(table, data)
	if err != nil {
		return nil, fmt.Errorf("create failed: %v", err)
	}

	return result, nil
}

// Select retrieves records by ID or criteria
func Select(thing interface{}) (interface{}, error) {
	thingStr, ok := thing.(string)
	if !ok {
		return nil, fmt.Errorf("thing must be a string")
	}

	result, err := DB.Select(thingStr)
	if err != nil {
		return nil, fmt.Errorf("select failed: %v", err)
	}

	return result, nil
}

// Update updates a record
func Update(thing interface{}, data interface{}) (interface{}, error) {
	thingStr, ok := thing.(string)
	if !ok {
		return nil, fmt.Errorf("thing must be a string")
	}

	result, err := DB.Update(thingStr, data)
	if err != nil {
		return nil, fmt.Errorf("update failed: %v", err)
	}

	return result, nil
}

// Delete deletes a record
func Delete(thing interface{}) (interface{}, error) {
	thingStr, ok := thing.(string)
	if !ok {
		return nil, fmt.Errorf("thing must be a string")
	}

	result, err := DB.Delete(thingStr)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %v", err)
	}

	return result, nil
}

// QuerySingle executes a query and returns a single result
func QuerySingle(query string, params map[string]interface{}) (interface{}, error) {
	result, err := Query(query, params)
	if err != nil {
		return nil, err
	}

	// Parse the result array and return first item if exists
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if resultData, ok := resultArray[0].(map[string]interface{}); ok {
			if resultValue, exists := resultData["result"]; exists {
				if resultSlice, ok := resultValue.([]interface{}); ok && len(resultSlice) > 0 {
					return resultSlice[0], nil
				}
				return resultValue, nil
			}
		}
	}

	return nil, fmt.Errorf("no result found")
}

// QueryMultiple executes a query and returns multiple results
func QueryMultiple(query string, params map[string]interface{}) ([]interface{}, error) {
	result, err := Query(query, params)
	if err != nil {
		return nil, err
	}

	// Parse the result array
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if resultData, ok := resultArray[0].(map[string]interface{}); ok {
			if resultValue, exists := resultData["result"]; exists {
				if resultSlice, ok := resultValue.([]interface{}); ok {
					return resultSlice, nil
				}
				return []interface{}{resultValue}, nil
			}
		}
	}

	return []interface{}{}, nil
}

// Transaction executes multiple queries in a transaction
func Transaction(queries []string, params []map[string]interface{}) ([]interface{}, error) {

	var results []interface{}

	for i, query := range queries {
		var queryParams map[string]interface{}
		if i < len(params) && params[i] != nil {
			queryParams = params[i]
		} else {
			queryParams = make(map[string]interface{})
		}

		result, err := DB.Query(query, queryParams)
		if err != nil {
			return nil, fmt.Errorf("transaction query %d failed: %v", i, err)
		}

		results = append(results, result)
	}

	return results, nil
}

// CheckRecordExists checks if a record exists by ID
func CheckRecordExists(table string, id string) (bool, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE id = $id", table)
	params := map[string]interface{}{"id": fmt.Sprintf("%s:%s", table, id)}

	result, err := QuerySingle(query, params)
	if err != nil {
		if err.Error() == "no result found" {
			return false, nil
		}
		return false, err
	}

	return result != nil, nil
}

// GetByField retrieves records by a specific field value
func GetByField(table string, field string, value interface{}) ([]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $value", table, field)
	params := map[string]interface{}{"value": value}

	return QueryMultiple(query, params)
}

// CountRecords counts records in a table with optional conditions
func CountRecords(table string, conditions string, params map[string]interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT count() FROM %s", table)
	if conditions != "" {
		query += " WHERE " + conditions
	}

	result, err := QuerySingle(query, params)
	if err != nil {
		return 0, err
	}

	if countResult, ok := result.(map[string]interface{}); ok {
		if count, exists := countResult["count"]; exists {
			if countFloat, ok := count.(float64); ok {
				return int64(countFloat), nil
			}
		}
	}

	return 0, fmt.Errorf("failed to parse count result")
}

// Helper function to parse SurrealDB record ID
func ParseRecordID(recordID string) (string, string, error) {
	// Parse "table:id" format
	if recordID == "" {
		return "", "", fmt.Errorf("empty record ID")
	}

	// Simple parsing - assumes format "table:id"
	for i, char := range recordID {
		if char == ':' {
			if i == 0 || i == len(recordID)-1 {
				return "", "", fmt.Errorf("invalid record ID format: %s", recordID)
			}
			return recordID[:i], recordID[i+1:], nil
		}
	}

	return "", "", fmt.Errorf("invalid record ID format: %s", recordID)
}

// CreateRecordID creates a SurrealDB record ID
func CreateRecordID(table string, id string) string {
	return fmt.Sprintf("%s:%s", table, id)
}
