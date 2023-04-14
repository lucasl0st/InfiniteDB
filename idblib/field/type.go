/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

type DatabaseType string

const (
	BOOL   DatabaseType = "bool"
	NUMBER DatabaseType = "number"
	TEXT   DatabaseType = "text"
)

var databaseTypesMap = map[string]DatabaseType{
	"boolean": BOOL,
	"number":  NUMBER,
	"text":    TEXT,
}

func ParseDatabaseType(s string) *DatabaseType {
	t, ok := databaseTypesMap[s]

	if !ok {
		return nil
	}

	return &t
}
