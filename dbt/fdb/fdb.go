package fdb

// ----------------------------------------------------------------------------------
// fdb.go for Go's dbx package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"github.com/waldurbas/dbx"

	// firebirdsql #
	_ "github.com/waldurbas/firebirdsql"
)

// NewDatabase #new instance
func NewDatabase(a interface{}) *dbx.DB {
	db := dbx.NewDB("firebirdsql", a)

	db.AddOp(dbx.OpExistTable, `select count(*) from RDB$RELATIONS where RDB$RELATION_NAME='%s' and RDB$VIEW_BLR is NULL`)
	db.AddOp(dbx.OpExistTableCol, `select count(*) from RDB$RELATION_FIELDS b where b.RDB$RELATION_NAME='%s' and b.RDB$FIELD_NAME='%s'`)
	db.AddOp(dbx.OpExistProc, `select count(*) from RDB$PROCEDURES where RDB$PROCEDURE_NAME='%s'`)
	db.AddOp(dbx.OpExistFunc, `select count(*) from RDB$FUNCTIONS where RDB$FUNCTION_NAME='%s'`)
	db.AddOp(dbx.OpExistTrg, `select count(*) from RDB$TRIGGERS where RDB$RELATION_NAME='%s' and RDB$TRIGGER_NAME='%s'`)

	return db
}
