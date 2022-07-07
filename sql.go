package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// SQLResult holds the result of a SQL query
type SQLResult struct {
	Columns []string
	Values  [][]interface{}
	Strings [][]string
}

// github.com/lib/pq returns the following types:
// - nil (NULL values)
// - int64
// - float64
// - string
// - time.Time
// - bool
// - []byte

func sqlConnect(connStr string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func sqlQuery(db *sqlx.DB, query string, args ...interface{}) (SQLResult, error) {
	result := SQLResult{}
	rows, err := db.Query(query, args...)
	if err != nil {
		return result, err
	}
	result.Columns, err = rows.Columns()
	if err != nil {
		return result, err
	}
	for rows.Next() {
		types, _ := rows.ColumnTypes()
		values := make([]interface{}, len(result.Columns))
		for i := range values {
			values[i] = new(interface{})
			// "_text" is the PostgreSQL database type name for []text, or array of strings.
			// See:
			// - https://stackoverflow.com/questions/71332976/text-postgres-data-type
			// - https://dba.stackexchange.com/questions/307344/what-is-the-difference-between-postgres-text-and-text-array-types
			if strings.EqualFold("_text", types[i].DatabaseTypeName()) {
				array := make(pq.StringArray, 0)
				values[i] = &array
			}
		}
		err = rows.Scan(values...)
		if err != nil {
			return result, err
		}
		strs := make([]string, len(result.Columns))
		moreArrays := false
		for i := range values {
			switch ptr := values[i].(type) {
			case *interface{}:
				values[i] = *ptr
				strs[i] = sqlString(*ptr)
			case *pq.StringArray:
				arr := *ptr
				slice := make([]string, len(arr))
				for j := range arr {
					slice[j] = string(arr[j])
				}
				values[i] = slice
				if len(arr) > 0 {
					strs[i] += sqlString(arr[0])
				}
				if len(arr) > 1 {
					moreArrays = true
				}
			}
		}
		result.Values = append(result.Values, values)
		result.Strings = append(result.Strings, strs)
		for j := 1; moreArrays; j++ {
			strs = make([]string, len(result.Columns))
			moreArrays = false
			for i := range values {
				if arr, ok := values[i].([]string); ok {
					if len(arr) > j {
						strs[i] = sqlString(arr[j])
					}
					if len(arr) > j+1 {
						moreArrays = true
					}
				}
			}
			result.Values = append(result.Values, values)
			result.Strings = append(result.Strings, strs)
		}
	}
	return result, nil
}

func sqlString(a interface{}) string {
	if t, ok := a.(time.Time); ok {
		if t.Truncate(24*time.Hour) == t {
			return t.Format("2006-01-02")
		}
		if t.Year() == 0 && t.Month() == 1 && t.Day() == 1 {
			return t.Format("15:04:05")
		}
		return t.Format("2006-01-02 15:04:05")
	}
	if s, ok := a.([]byte); ok {
		return fmt.Sprint(string(s))
	}
	if a == nil {
		return ""
	}
	return fmt.Sprint(a)
}

// sqlBind converts a query with DOLLAR bindvars ($1, $2...) into driver's bindvar type
func sqlBind(db *sqlx.DB, query string, args []string) (string, []interface{}) {
	var res string
	var resArgs []interface{}
	// First, we convert DOLLARs into QUESTIONs
	for i := strings.Index(query, "$"); i != -1; i = strings.Index(query, "$") {
		res += query[:i]
		query = query[i+1:]
		argNum := 0
		for len(query) > 0 && query[0] >= '0' && query[0] <= '9' {
			argNum *= 10
			argNum += int(query[0]) - '0'
			query = query[1:]
		}
		res += "?"
		resArgs = append(resArgs, args[argNum-1])
	}
	res += query
	return db.Rebind(res), resArgs
}
