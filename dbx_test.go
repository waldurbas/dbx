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
	"strings"

	"github.com/waldurbas/dbx"
	"github.com/waldurbas/dbx/dbt/fdb"
	"github.com/waldurbas/dbx/dbt/myd"
	"github.com/waldurbas/dbx/script"

	"testing"
)

func TestSplit(*testing.T) {
	s := "create exception DEVI_MISSING 'entry in table DEVI is missing'"
	ss := script.SplitLine([]rune(s))
	for i, s := range ss {
		fmt.Printf("%d: [%s]\n", i, s)
	}
}

func TestFDB(t *testing.T) {
	conStr := os.Getenv("FDB_CON")
	if conStr == "" {
		//		t.Errorf("env.variable FDB_CON not defined")
		fmt.Printf("\nenv.variable FDB_CON not defined")
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
		ExecCmd: func(cmdID int, idx int, cmd string) (bool, error) {
			fmt.Printf("\ncmdID: %d, idx: %d, cmd: [%s]\n", cmdID, idx, cmd)

			switch cmdID {
			case script.TkExit:
				fmt.Println("->EXIT...")
				return true, nil

			case script.TkEcho:
				// 0123456
				// $echo ..
				fmt.Println("->echo:", strings.TrimSpace(cmd[6:]))
				return false, nil

			case script.TkSet:
				s := strings.TrimSpace(cmd[5:])
				if len(s) > 0 && s[:1] == "@" {
					fmt.Println("->set:", s[1:])
					return false, nil
				}
			}

			getCmd := func(cmdID int, cmd string) string {
				switch cmdID {
				case script.TkAdd:
					ss := strings.Fields(cmd)
					le := len(ss)
					if le > 3 {
						if ss[1] == "field" {
							tbl := strings.Split(ss[2], ".")
							if len(tbl) == 2 {
								sf := tbl[1]
								for _, xf := range ss[3:] {
									sf = sf + " " + xf
								}

								return fmt.Sprintf("alter table %s add %s", tbl[0], sf)
							}
						}
					}
				}
				return cmd
			}

			sq := getCmd(cmdID, cmd)
			if sq == "" {
				return false, nil
			}

			fmt.Println("->sql :", sq)

			return false, nil
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
	dbs.ExistDomain = db.ExistDomain
	dbs.ExistException = db.ExistException
	dbs.SaveVers = xdb.SaveVers
	/*
		// select for Version
		sver := os.Getenv("FDB_VER")
		if sver == "" {
			t.Errorf("\n\nenv.variable FDB_VER not defined")
			return
		}
		q := db.ExecQ(sver)
		if q.Fetch() {
			dbs.Vinfo.Dbu = q.AsInteger(0)
			dbs.Vinfo.App = q.AsString(1)
			dbs.Vinfo.Chg = q.AsString(2)
		}
		q.Close()
	*/

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

	fmt.Println("Execute.Script OK..")
}

func TestMYD(t *testing.T) {
	conStr := os.Getenv("MYD_CON")
	if conStr == "" {
		fmt.Println("\nenv.variable MYD_CON not defined")
		return
	}

	c := dbx.ConStr2DBCfg(conStr)
	c.Print()

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

	sql := os.Getenv("MYD_SQL")
	if len(sql) > 0 {
		if strings.Index(sql, "call ") == 0 {
			qf := db.Call(db, sql)
			if qf != nil {

				fmt.Print(qf.ShowLine(true))
				fmt.Print(qf.ShowLine(false))
			}
			qf.Close()

		}
	}

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

func TestFunc(t *testing.T) {
	var ar = []struct {
		vs string
		vi string
	}{
		{"-20", "-20"},
		{"  0- 20  ", "-020"},
		{" + 25  ", "25"},
		{"-  40  ", "-40"},
		{"  30-  ", "-30"},
	}

	f := dbx.SqxField{Typ: "INT"}

	for _, a := range ar {
		f.Value = []byte(a.vs)
		s := f.CleanedString(false)

		if a.vi != s {
			t.Errorf("CleanString('%s'): soll '%s', ist '%s'", a.vs, a.vi, s)
		}
	}
}
