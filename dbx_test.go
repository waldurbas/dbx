package dbx_test

// ----------------------------------------------------------------------------------
// dbx_test.go for Go's dbx package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"log"
	"os"

	"github.com/waldurbas/dbx"
	"github.com/waldurbas/dbx/dbt/fdb"
	"github.com/waldurbas/dbx/dbt/myd"

	"testing"
)

func TestFDB(t *testing.T) {

	conStr := os.Getenv("FDB_CON")
	if conStr == "" {
		t.Errorf("env.variable FDB_CON not defined")
		return
	}

	db := fdb.NewDatabase(conStr)
	c := dbx.ConStr2DBCfg(conStr)

	log.Println("fdb.connect", c)
	if !db.Connect() {
		t.Errorf("db.Connect fail, err: %v", db.Err)
		return
	}
	defer func() {
		log.Println("fdb.disconnect..")
		db.Close()
	}()

	log.Println("check TableExists")
	b := db.ExistTable("RDB$DATABASE")
	if db.Err != nil || !b {
		t.Errorf("Exist-Table fail, err: %v", db.Err)
		return
	}

	n := db.ExecI("select count(*) from rdb$database")
	if db.Err != nil || n != 1 {
		t.Errorf("select fail, err: %v", db.Err)
		return
	}
}

func TestMYD(t *testing.T) {

	conStr := os.Getenv("MYD_CON")
	if conStr == "" {
		t.Errorf("env.variable MYD_CON not defined")
		return
	}

	c := dbx.ConStr2DBCfg(conStr)
	db := myd.NewDatabase(*c)

	log.Println("myd.connect", dbx.DBCfg2ConStr(*c))
	if !db.Connect() {
		t.Errorf("myd.Connect fail, err: %v", db.Err)
		return
	}
	defer func() {
		log.Println("myd.disconnect..")
		db.Close()
	}()

	log.Println("check select count(*)")
	n := db.ExecI("select count(*) from kunden")
	if db.Err != nil || n < 0 {
		t.Errorf("select fail, err: %v", db.Err)
		return
	}
}
