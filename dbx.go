package dbx

// ----------------------------------------------------------------------------------
// dbx.go for Go's dbx package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// OpExistTable #
const OpExistTable string = "existTable"

// OpExistTableCol #
const OpExistTableCol string = "existTableCol"

// OpExistProc #
const OpExistProc = "existProc"

// OpExistFunc #
const OpExistFunc = "existFunc"

// OpExistIdx #
const OpExistIdx = "existIdx"

// Debug #
var Debug int

// DBCfg #
type DBCfg struct {
	User     string
	Pass     string
	Instance string
	DBName   string
	OrgDbn   string
}

// DB #
type DB struct {
	*sql.DB
	DrvName string
	opened  bool
	Cfg     DBCfg
	Err     error
	dbOp    map[string]string
	ExitF   func(int)
}

// DBCfg2ConStr #
func DBCfg2ConStr(c DBCfg) string {
	return c.User + ":" + c.Pass + "@" + c.Instance + "/" + c.DBName
}

// ConStr2DBCfg #
func ConStr2DBCfg(s string) *DBCfg {
	sup := strings.Split(s, "@")

	if len(sup) > 1 {
		su := strings.Split(sup[0], ":")
		if len(su) == 2 {
			c := DBCfg{}
			c.User = su[0]
			c.Pass = su[1]

			sd := strings.Split(sup[1], "/")
			le := len(sd)
			if le > 1 {
				c.DBName = sd[le-1]
				le = len(sup[1])
				c.Instance = (sup[1])[0 : le-(len(c.DBName)+1)]
				return &c
			}
		}
	}

	return nil
}

// NewDB #new instance
func NewDB(drvName string, a interface{}) *DB {
	var c *DBCfg

	db := &DB{DrvName: drvName}
	switch v := a.(type) {
	case string:
		c = ConStr2DBCfg(v)
	case DBCfg:
		c = &v
	}

	if c == nil {
		fmt.Println("NewDB.error: bad conString")
		os.Exit(1)
	}

	db.dbOp = make(map[string]string)
	db.Cfg = *c

	cstr := DBCfg2ConStr(db.Cfg)
	db.DB, db.Err = sql.Open(db.DrvName, cstr)
	if db.Err != nil {
		fmt.Println("NewDB.error:", db.Err)
		os.Exit(1)
	}

	db.ExitF = db.exitFunc

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db
}

func lgx(format string, x ...interface{}) {
	if Debug > 0 {
		log.Printf(format, x...)
	}
}

// Print #DBCfg
func (c *DBCfg) Print() {
	fmt.Println("User    :", c.User)
	fmt.Println("Pass    :", c.Pass)
	fmt.Println("Instance:", c.Instance)
	fmt.Println("DBName  :", c.DBName)
}

func (v *DB) exitFunc(sta int) {
	v.Close()
	os.Exit(sta)
}

// ErrMsg #
func (v *DB) ErrMsg(err error) string {
	serr := fmt.Sprintf("%v", err)

	return strings.ReplaceAll(serr, "\n", "|")
}

// Fatal #mit zus. Message
func (v *DB) Fatal(msg string, err error) {
	fmt.Println("error: ", msg, v.ErrMsg(err))
	v.ExitF(1)
}

// Connect to Database
func (v *DB) Connect() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	v.Err = v.DB.PingContext(ctx)
	v.opened = v.Err == nil
	return v.opened
}

// Close #close database
func (v *DB) Close() {
	if v.opened {
		v.opened = false
		v.DB.Close()
	}
}

// AddOp #
func (v *DB) AddOp(key string, val string) {
	v.dbOp[key] = val
}

func (v *DB) prepareSqText(format string, x ...interface{}) string {
	return fmt.Sprintf(format, x...)
}

//ExecI64 # asInteger
func (v *DB) ExecI64(format string, x ...interface{}) int64 {
	sq := v.prepareSqText(format, x...)
	r := v.QueryRow(sq)

	var n int64
	v.Err = r.Scan(&n)
	if v.Err == nil {
		return n
	}

	return 0
}

// ExecI #
func (v *DB) ExecI(format string, x ...interface{}) int {
	sq := v.prepareSqText(format, x...)
	r := v.QueryRow(sq)

	var n int
	v.Err = r.Scan(&n)
	if v.Err == nil {
		return n
	}

	return 0
}

// ExecS # asString
func (v *DB) ExecS(format string, x ...interface{}) string {
	r := v.QueryRow(v.prepareSqText(format, x...))

	var s string
	v.Err = r.Scan(&s)
	if v.Err == nil {
		return s
	}

	return ""
}

// ExecQ #
func (v *DB) ExecQ(format string, x ...interface{}) *SQLX {
	q := NewSQLX(v)
	q.Exec(v.prepareSqText(format, x...))
	return q
}

// CreateSqlx #
func (v *DB) CreateSqlx() *SQLX {
	return NewSQLX(v)
}

// ExecSqlx #
func (v *DB) ExecSqlx(format string, x ...interface{}) *SQLX {
	sq := fmt.Sprintf(format, x...)

	q := NewSQLX(v)
	q.prepareStmt(sq)
	if (q.StmtType & StmtOutf) == StmtOutf {
		q.execSelect(sq)
	} else {
		_, q.Err = q.db.Exec(sq)
	}

	if q.Err != nil {
		v.Fatal("exec", q.Err)
	}

	return q
}

// ExecuteF #
func (v *DB) ExecuteF(statement string) {
	if v.Execute(statement) != nil {
		v.ExitF(1)
	}
}

// Execute #
func (v *DB) Execute(statement string) error {
	_, err := v.DB.Exec(statement)

	if err != nil {
		le := len(statement)
		if le > 256 {
			le = 256
		}
		pfx := statement[0:le]
		fmt.Println("error: db.Execute."+pfx, v.ErrMsg(err))
		return err
	}

	return nil
}

// ExistTable #
func (v *DB) ExistTable(sName string) bool {
	sq := v.dbOp[OpExistTable]
	return len(sq) > 9 && v.ExecI(sq, sName) > 0
}

// ExistTableCol #
func (v *DB) ExistTableCol(sName string) bool {
	sq := v.dbOp[OpExistTableCol]
	elem := strings.Split(sName, ".")
	if len(sq) > 9 && len(elem) == 2 {
		return v.ExecI(sq, elem[0], elem[1]) > 0
	}

	return false
}

// ExistIndex #
func (v *DB) ExistIndex(sName string) bool {
	sq := v.dbOp[OpExistIdx]

	elem := strings.Split(sName, ".")
	if len(sq) > 9 && len(elem) == 2 {
		return v.ExecI(sq, elem[0], elem[1]) > 0
	}

	return false
}

// ExistProc #
func (v *DB) ExistProc(sName string) bool {
	sq := v.dbOp[OpExistProc]
	return len(sq) > 9 && v.ExecI(sq, sName) > 0
}

// ExistFunc #
func (v *DB) ExistFunc(sName string) bool {
	sq := v.dbOp[OpExistFunc]
	return len(sq) > 9 && v.ExecI(sq, sName) > 0
}
