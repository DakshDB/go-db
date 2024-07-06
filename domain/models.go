package domain

import "time"

// Row Define a Row type to hold data
type Row struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
	Data      map[string]interface{}
}

// Table Define a Table type to hold rows
type Table struct {
	Name    string
	Columns []string
	Rows    []Row
	NextID  int
}

// Database Define a Database type to hold tables
type Database struct {
	Tables map[string]*Table
}
