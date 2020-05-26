package myd

// ----------------------------------------------------------------------------------
// myd.go for Go's dbx package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------
// HISTORY
// ----------------------------------------------------------------------------------
// 2020.05.26 meine Änderung 2020.02.22 @Wald.Urbas for strings (x.length >> 2)
//            wurde vom github.com/go-sql-driver/mysql übernommen
//            switch to "github.com/go-sql-driver/mysql"
// ----------------------------------------------------------------------------------

import (
	"github.com/waldurbas/dbx"

	// mysql #
	_ "github.com/go-sql-driver/mysql"
)

// NewDatabase #new instance
func NewDatabase(a interface{}) *dbx.DB {
	db := dbx.NewDB("mysql", a)

	db.AddOp(dbx.OpExistTable, "select count(*) from INFORMATION_SCHEMA.TABLES where TABLE_SCHEMA='"+db.Cfg.DBName+"' and TABLE_NAME='%s'")
	db.AddOp(dbx.OpExistTableCol, "select count(*) from INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA='"+db.Cfg.DBName+"' and TABLE_NAME='%s' and COLUMN_NAME='%s'")
	db.AddOp(dbx.OpExistIdx, "select count(*) from INFORMATION_SCHEMA.STATISTICS where TABLE_SCHEMA='"+db.Cfg.DBName+"' and TABLE_NAME='%s' and INDEX_NAME='%s'")

	return db
}
