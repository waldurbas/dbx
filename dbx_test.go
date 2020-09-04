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
	"fmt"
	"os"

	"github.com/waldurbas/dbx"
	"github.com/waldurbas/dbx/dbt/fdb"
	"github.com/waldurbas/dbx/dbt/myd"
	"github.com/waldurbas/dbx/script"

	"testing"
)

func TestFDB(t *testing.T) {
	conStr := os.Getenv("FDB_CON")
	if conStr == "" {
		t.Errorf("env.variable FDB_CON not defined")
		return
	}

	db := fdb.NewDatabase(conStr)
	if !db.Connect() {
		t.Errorf("db.Connect fail, err: %v", db.Err)
		return
	}
	defer db.Close()

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

	type dbu struct {
		db       *dbx.DB
		ExecCmd  func(a int, ix int, cmd string) (bool, error)
		SaveVers func(v int) error
	}

	xdb := &dbu{db: db,
		ExecCmd: func(a int, ix int, cmd string) (bool, error) {
			if cmd == "$exit" {
				return true, nil
			}

			_, err := db.Exec(cmd)

			return false, err
		},
		SaveVers: func(v int) error {
			return nil
		},
	}

	dbs := script.NewScript()
	dbs.ExecCmd = xdb.ExecCmd
	dbs.ExistFunc = db.ExistFunc
	dbs.ExistIndex = db.ExistIndex
	dbs.ExistProc = db.ExistProc
	dbs.ExistTrigger = db.ExistTrigger
	dbs.ExistTable = db.ExistTable
	dbs.ExistTableCol = db.ExistTableCol
	dbs.SaveVers = xdb.SaveVers

	// select for Version
	sver := os.Getenv("FDB_VER")
	if sver == "" {
		t.Errorf("env.variable FDB_VER not defined")
		return
	}
	q := db.ExecQ(sver)
	if q.Fetch() {
		dbs.Vinfo.Dbu = q.AsInteger(0)
		dbs.Vinfo.App = q.AsString(1)
		dbs.Vinfo.Chg = q.AsString(2)
	}
	q.Close()

	// scriptname
	scr := os.Getenv("FDB_SCR")
	if scr == "" {
		t.Errorf("env.variable FDB_SCR not defined")
		return
	}

	px := script.NewParser()
	err := px.LoadFile(scr)
	if err != nil {
		fmt.Printf("LoadScript, err: %v", err)
		return
	}

	a, err := dbs.Execute(px)
	if err != nil {
		fmt.Printf("Execute.Script, a:%d, err: %v", a, err)
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

	if !db.Connect() {
		t.Errorf("myd.Connect fail, err: %v", db.Err)
		return
	}
	defer func() {
		db.Close()
	}()

	n := db.ExecI("select count(*) from INFORMATION_SCHEMA.TABLES")
	if db.Err != nil || n < 0 {
		t.Errorf("select fail, err: %v", db.Err)
		return
	}

	type dbu struct {
		db       *dbx.DB
		ExecCmd  func(a int, ix int, cmd string) (bool, error)
		SaveVers func(v int) error
	}

	xdb := &dbu{db: db,
		ExecCmd: func(a int, ix int, cmd string) (bool, error) {
			if cmd == "$exit" {
				return true, nil
			}

			_, err := db.Exec(cmd)

			return false, err
		},
		SaveVers: func(v int) error {
			return nil
		},
	}

	dbs := script.NewScript()
	dbs.ExecCmd = xdb.ExecCmd
	dbs.ExistFunc = db.ExistFunc
	dbs.ExistIndex = db.ExistIndex
	dbs.ExistProc = db.ExistProc
	dbs.ExistTable = db.ExistTable
	dbs.ExistTableCol = db.ExistTableCol
	dbs.SaveVers = xdb.SaveVers

	sver := os.Getenv("MYD_VER")
	if sver == "" {
		t.Errorf("env.variable MYD_VER not defined")
		return
	}
	q := db.ExecQ(sver)
	if q.Fetch() {
		dbs.Vinfo.Dbu = q.AsInteger(0)
		dbs.Vinfo.App = q.AsString(1)
		dbs.Vinfo.Chg = q.AsString(2)
	}
	q.Close()

	scr := os.Getenv("MYD_SCR")
	if scr == "" {
		t.Errorf("env.variable MYD_SCR not defined")
		return
	}

	px := script.NewParser()
	err := px.LoadFile(scr)
	if err != nil {
		fmt.Printf("LoadScript, err: %v", err)
		return
	}

	a, err := dbs.Execute(px)
	if err != nil {
		fmt.Printf("Execute.Script, a:%d, err: %v", a, err)
		return
	}
}
