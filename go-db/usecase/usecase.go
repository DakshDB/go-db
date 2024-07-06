package usecase

import (
	"encoding/json"
	"errors"
	"go-db/config"
	"go-db/domain"
	"os"
	"strconv"
	"strings"
	"time"
)

type godbUsecase struct {
}

func NewGoDBUsecase() domain.GoDBUsecase {
	return &godbUsecase{}
}

func (h *godbUsecase) HealthCheck() (string, error) {
	return "true", nil
}

// CreateTable Create a new table in the database
func (h *godbUsecase) CreateTable(databaseName string, tableName string, columns []string) error {
	// Since the table does not exist, create a new table and assign it directly in the map
	config.DBMap[databaseName].Tables[tableName] = &domain.Table{
		Name:    tableName,
		Columns: columns,
		Rows:    []domain.Row{},
		NextID:  1,
	}
	return nil
}

// Insert a row into a table
func (h *godbUsecase) Insert(databaseName string, tableName string, data map[string]interface{}) error {
	table, exists := config.DBMap[databaseName].Tables[tableName]
	if !exists {
		return errors.New("table does not exist")
	}
	// Check if the data has the correct column names
	for key := range data {
		found := false
		for _, col := range table.Columns {
			if key == col {
				found = true
				break
			}
		}
		if !found {
			return errors.New("data contains unknown column")
		}
	}

	currentTime := time.Now()
	newRow := domain.Row{
		ID:        table.NextID,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Data:      data,
	}
	table.Rows = append(table.Rows, newRow)
	table.NextID++
	return nil
}

// Update a row in a table
func (h *godbUsecase) Update(databaseName string, tableName string, id int, newData map[string]interface{}) error {
	table, exists := config.DBMap[databaseName].Tables[tableName]
	if !exists {
		return errors.New("table does not exist")
	}

	// Check if the data has the correct number of columns
	if len(newData) != len(table.Columns) {
		return errors.New("data columns count mismatch")
	}

	// Check if the data has the correct column names
	for key := range newData {
		found := false
		for _, col := range table.Columns {
			if key == col {
				found = true
				break
			}
		}
		if !found {
			return errors.New("data contains unknown column")
		}
	}

	for i, row := range table.Rows {
		if row.ID == id {
			row.Data = newData
			row.UpdatedAt = time.Now()
			table.Rows[i] = row // Update the row in the slice
			return nil
		}
	}
	return errors.New("row not found")
}

// Select rows from a table
func (h *godbUsecase) Select(databaseName string, tableName string, condition map[string]interface{}) ([]domain.Row, error) {
	table, exists := config.DBMap[databaseName].Tables[tableName]
	if !exists {
		return nil, errors.New("table does not exist")
	}
	if len(condition) == 0 {
		return table.Rows, nil
	}
	var filteredRows []domain.Row
	for _, row := range table.Rows {
		match := true
		for key, value := range condition {
			if rowVal, ok := row.Data[key]; !ok || rowVal != value {
				match = false
				break
			}
		}
		if match {
			filteredRows = append(filteredRows, row)
		}
	}
	return filteredRows, nil
}

// Execute Simple query parser
func (h *godbUsecase) Execute(databaseName string, query string) (interface{}, error) {
	parts := strings.Fields(query)
	if len(parts) == 0 {
		return nil, errors.New("empty query")
	}
	switch parts[0] {
	case "CREATE":
		// Splitting the query to extract table name and column names
		createParts := strings.SplitN(query, " ", 3)
		if len(createParts) < 3 {
			return nil, errors.New("invalid CREATE TABLE syntax")
		}
		tableNameAndColumns := strings.TrimSpace(createParts[2])
		openParenIndex := strings.Index(tableNameAndColumns, "(")
		closeParenIndex := strings.LastIndex(tableNameAndColumns, ")")
		if openParenIndex == -1 || closeParenIndex == -1 || closeParenIndex < openParenIndex {
			return nil, errors.New("invalid CREATE TABLE syntax, missing or incorrect parentheses")
		}
		tableName := tableNameAndColumns[:openParenIndex]
		tableName = strings.TrimSpace(tableName)
		columnsString := tableNameAndColumns[openParenIndex+1 : closeParenIndex]
		columns := strings.Split(columnsString, ",")
		for i, col := range columns {
			columns[i] = strings.TrimSpace(col)
		}
		// Assuming CreateTable now accepts a slice of column names
		return nil, h.CreateTable(databaseName, tableName, columns)
	case "ADD":
		if len(parts) < 4 || parts[1] != "INTO" {
			return nil, errors.New("invalid ADD syntax")
		}
		tableName := parts[2]
		// Find the indices of the parentheses to extract column names and values
		openParenIndex := strings.Index(query, "(")
		closeParenIndex := strings.Index(query, ")")
		if openParenIndex == -1 || closeParenIndex == -1 || closeParenIndex < openParenIndex {
			return nil, errors.New("invalid ADD syntax, missing parentheses")
		}
		columnNames := strings.Split(query[openParenIndex+1:closeParenIndex], ",")
		// Extract values, assuming they are immediately after column names
		valuesPart := query[closeParenIndex+2:]
		openValuesIndex := strings.Index(valuesPart, "(")
		closeValuesIndex := strings.Index(valuesPart, ")")
		if openValuesIndex == -1 || closeValuesIndex == -1 || closeValuesIndex < openValuesIndex {
			return nil, errors.New("invalid ADD syntax, missing value parentheses")
		}
		values := strings.Split(valuesPart[openValuesIndex+1:closeValuesIndex], ",")
		if len(columnNames) != len(values) {
			return nil, errors.New("column names and values count mismatch")
		}
		// Trim spaces and quotes from column names and values
		for i := range columnNames {
			columnNames[i] = strings.TrimSpace(columnNames[i])
			values[i] = strings.Trim(strings.TrimSpace(values[i]), "\"")
		}
		// Prepare data map
		data := make(map[string]interface{})
		for i, columnName := range columnNames {
			data[columnName] = values[i]
		}
		return nil, h.Insert(databaseName, tableName, data)
	case "UPDATE":
		// Validate the basic structure of the query
		if !strings.HasPrefix(query, "UPDATE") || !strings.Contains(query, "WHERE id=") {
			return nil, errors.New("invalid UPDATE syntax")
		}

		// Extract table name
		parts := strings.Fields(query)
		tableName := parts[1]

		// Find the indices for column names and values
		openParenIndex := strings.Index(query, "(")
		closeParenIndex := strings.Index(query, ")")
		if openParenIndex == -1 || closeParenIndex == -1 || closeParenIndex < openParenIndex {
			return nil, errors.New("invalid UPDATE syntax, missing parentheses for columns")
		}
		columnNames := strings.Split(query[openParenIndex+1:closeParenIndex], ",")

		// Extract values, assuming they are immediately after column names
		valuesPart := query[closeParenIndex+2:]
		openValuesIndex := strings.Index(valuesPart, "(")
		closeValuesIndex := strings.Index(valuesPart, ")")
		if openValuesIndex == -1 || closeValuesIndex == -1 || closeValuesIndex < openValuesIndex {
			return nil, errors.New("invalid UPDATE syntax, missing parentheses for values")
		}
		values := strings.Split(valuesPart[openValuesIndex+1:closeValuesIndex], ",")

		// Validate column names and values count match
		if len(columnNames) != len(values) {
			return nil, errors.New("column names and values count mismatch")
		}

		// Trim spaces and quotes from column names and values
		for i := range columnNames {
			columnNames[i] = strings.TrimSpace(columnNames[i])
			values[i] = strings.Trim(strings.TrimSpace(values[i]), "\"")
		}

		// Extract ID from WHERE clause
		idPart := strings.SplitAfter(query, "WHERE id=")
		if len(idPart) < 2 {
			return nil, errors.New("missing ID in WHERE clause")
		}
		id, err := strconv.Atoi(strings.TrimSpace(idPart[1]))
		if err != nil {
			return nil, errors.New("invalid ID format")
		}

		// Prepare data map
		data := make(map[string]interface{})
		for i, columnName := range columnNames {
			data[columnName] = values[i]
		}

		// Call Update method
		return nil, h.Update(databaseName, tableName, id, data)
	case "GET":
		tableName := parts[3]
		condition := make(map[string]interface{})
		if strings.Contains(query, "WHERE") {
			whereParts := strings.SplitN(query, "WHERE", 2)
			if len(whereParts) == 2 {
				conditionParts := strings.Split(whereParts[1], "=")
				if len(conditionParts) == 2 {
					conditionKey := strings.TrimSpace(conditionParts[0])
					conditionValue := strings.TrimSpace(conditionParts[1])
					conditionValue = strings.Trim(conditionValue, "\"") // Assuming the value is always a string
					condition[conditionKey] = conditionValue
				}
			}
		}
		return h.Select(databaseName, tableName, condition)
	default:
		return nil, errors.New("unknown query type")
	}
}

// Save the database state to a file
func (h *godbUsecase) Save(databaseName string) error {
	db := config.DBMap[databaseName]

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(databaseName+".json", data, 0644)
}

// Load the database state from a file
func (h *godbUsecase) Load(databaseName string) error {
	// Check if the database is already loaded
	if _, exists := config.DBMap[databaseName]; exists {
		return nil
	}

	data, err := os.ReadFile(databaseName + ".json")
	if err != nil {
		// If the file does not exist, create a new database
		if os.IsNotExist(err) {
			config.DBMap[databaseName] = &domain.Database{Tables: make(map[string]*domain.Table)}
			return nil
		}
	}

	db := &domain.Database{Tables: make(map[string]*domain.Table)}
	if err := json.Unmarshal(data, db); err != nil {
		return err
	}

	config.DBMap[databaseName] = db

	return nil
}
