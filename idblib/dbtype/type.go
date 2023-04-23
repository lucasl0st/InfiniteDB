/*
 * Copyright (c) 2023 Lucas Pape
 */

package dbtype

type DatabaseType string

const (
	BOOL   DatabaseType = "bool"
	NUMBER DatabaseType = "number"
	TEXT   DatabaseType = "text"
)

var databaseTypesMap = map[string]DatabaseType{
	"bool":   BOOL,
	"number": NUMBER,
	"text":   TEXT,
}

var stringToDatabaseTypesMap = map[DatabaseType]string{
	BOOL:   "bool",
	NUMBER: "number",
	TEXT:   "text",
}

func ParseDatabaseType(s string) *DatabaseType {
	t, ok := databaseTypesMap[s]

	if !ok {
		return nil
	}

	return &t
}

func DatabaseTypeToString(t DatabaseType) string {
	return stringToDatabaseTypesMap[t]
}
