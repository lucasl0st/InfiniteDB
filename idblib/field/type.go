/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

type DatabaseType int

const (
	BOOL   DatabaseType = 1
	NUMBER DatabaseType = 14
	TEXT   DatabaseType = 24
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
