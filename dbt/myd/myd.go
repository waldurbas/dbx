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
// 2020.06.24 zurück zu eigener Version (fork) meine Änderung ist wieder weg
// 2020.05.26 meine Änderung 2020.02.22 @Wald.Urbas for strings (x.length >> 2)
//            wurde vom github.com/go-sql-driver/mysql übernommen
//            switch to "github.com/go-sql-driver/mysql"
// ----------------------------------------------------------------------------------

import (
	"fmt"
	"strings"

	"github.com/waldurbas/dbx"

	// mysql #
	_ "github.com/waldurbas/mysql"
)

// NewDatabase #new instance
func NewDatabase(a interface{}) *dbx.DB {
	db := dbx.NewDB("mysql", a)

	db.AddOp(dbx.OpExistTable, "select count(*) from INFORMATION_SCHEMA.TABLES where TABLE_SCHEMA='"+db.Cfg.DBName+"' and TABLE_NAME='%s'")
	db.AddOp(dbx.OpExistTableCol, "select count(*) from INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA='"+db.Cfg.DBName+"' and TABLE_NAME='%s' and COLUMN_NAME='%s'")
	db.AddOp(dbx.OpExistIdx, "select count(*) from INFORMATION_SCHEMA.STATISTICS where TABLE_SCHEMA='"+db.Cfg.DBName+"' and TABLE_NAME='%s' and INDEX_NAME='%s'")

	db.Call = Call
	return db
}

func Call(db *dbx.DB, sql string) *dbx.SQLX {
	if strings.Index(sql, "call ") == 0 {
		sf := strings.FieldsFunc(strings.TrimSpace(sql[5:]), func(r rune) bool {
			return r == ',' || r == '(' || r == ')'
		})

		if len(sf) > 1 {
			type ProcField struct {
				Idx  int
				Name string
				Typ  string
				Len  int
			}

			sq := `select p.ordinal_position,p.parameter_name,p.data_type,p.character_maximum_length as char_length,p.numeric_precision,p.numeric_scale
from information_schema.routines r
join information_schema.parameters p on p.specific_schema = r.routine_schema and p.specific_name = r.specific_name
where r.routine_schema = Database() and r.routine_type='PROCEDURE' and r.specific_name='` + sf[0] +
				`' and p.ordinal_position is not NULL and p.PARAMETER_MODE ='OUT' order by 1`

			fmt.Println("--> exec ProcFields")
			pFields := []*ProcField{}
			q := db.ExecQ(sq)
			for q.Fetch() {
				b := &ProcField{Idx: q.AsInteger(0), Name: q.AsString(1), Typ: strings.ToUpper(q.AsString(2)), Len: q.AsInteger(3)}
				pFields = append(pFields, b)

			}
			q.Close()

			nsql := "select "
			i := 0
			for _, f := range sf {
				if f[0] == '@' {
					if i > 0 {
						nsql = nsql + ","
					}
					nsql = nsql + f
					i++
				}
			}

			db.Exec(sql)
			qf := db.ExecQ(nsql)
			if qf.Fetch() {
				for a := 0; a < qf.ColNum; a++ {
					for _, b := range pFields {
						if b.Name == qf.Fields[a].Name[1:] {
							qf.Fields[a].Typ = b.Typ
							qf.Fields[a].Name = b.Name
							switch b.Typ {
							case "INT":
								qf.Fields[a].OutLen = 11
							case "VARCHAR":
								qf.Fields[a].OutLen = b.Len
							}
							break
						}
					}
				}
				return qf
			}
		}
	}

	return nil
}
