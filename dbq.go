package dbx

// ----------------------------------------------------------------------------------
// dbq.go for Go's dbx package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	// StmtUnknown #
	StmtUnknown = 0
	// StmtSelect #
	StmtSelect = 1
	// StmtExec #
	StmtExec = 2
	// StmtBlock #
	StmtBlock = 4
	// StmtOutf #
	StmtOutf = 8
	// StmtExecProc #
	StmtExecProc = 16
)

// SQLX #
type SQLX struct {
	r          *sql.Rows
	db         *DB
	scanValues []interface{}
	closed     bool

	ColNum    int
	Err       error
	Fields    []SqxField
	Lfd       int
	StmtType  int
	Statement string
	QName     string
	NewLine   bool
	Replacer  *strings.Replacer
}

// NewSQLX #new instance
func NewSQLX(v *DB) *SQLX {
	return &SQLX{db: v, closed: true, NewLine: true}
}

// PrepareStmt Sql.StmtType
func (q *SQLX) prepareStmt(sqlText string) {
	q.Statement = sqlText
	qf := strings.Fields(sqlText)

	a := -1
	q.StmtType = StmtUnknown
	for i, f := range qf {
		f = strings.ToLower(f)
		if f == "as" || f == ")as" {
			break
		}

		switch f {
		case "returns":
			fallthrough
		case ")returns(":
			fallthrough
		case ")returns":
			fallthrough
		case "returns(":
			q.StmtType |= StmtOutf

		case "select":
			q.StmtType |= StmtSelect | StmtOutf
		case "execute":
			a = i
		case "procedure":
			if (a + 1) == i {
				q.StmtType |= StmtExecProc
			}
		case "block":
			if (a + 1) == i {
				q.StmtType |= StmtBlock
			}
		}
	}

	if q.StmtType == StmtUnknown {
		q.StmtType = StmtExec
	}

	if (q.StmtType & StmtSelect) > 0 {
		qn := -1
		for _, key := range qf {
			if qn == -1 {
				if strings.ToLower(key) == "from" {
					qn = 0
					continue
				}
			} else {
				q.QName = key
				break
			}
		}
	}
}

// execSelect #
func (q *SQLX) execSelect(sq string) bool {
	q.r, q.Err = q.db.Query(sq)
	if q.Err != nil {
		return false
	}

	cols, _ := q.r.ColumnTypes()
	q.ColNum = len(cols)
	q.scanValues = make([]interface{}, q.ColNum)
	q.Fields = make([]SqxField, q.ColNum)

	//	lgx("sql: [%v]", sq)

	for i, col := range cols {
		q.scanValues[i] = &q.Fields[i].Value
		f := &q.Fields[i]

		f.Idx = i
		f.Q = q

		_len, ok := col.Length()
		if ok {
			f.Len = int(_len)
		}
		f.Name = col.Name()
		f.Typ = col.DatabaseTypeName()
		f.rTyp = col.ScanType()

		switch f.Typ {
		case "TINYINT":
			f.OutLen = 4
		case "MEDIUMINT":
			f.OutLen = 8
		case "SMALLINT":
			f.Typ = "SHORT"
			fallthrough
		case "SHORT":
			f.OutLen = 6
		case "LONG":
			f.Typ = "INT"
			fallthrough
		case "INT":
			f.OutLen = 11
		case "DATETIME":
			f.Typ = "TIMESTAMP"
			fallthrough
		case "TIMESTAMP":
			f.OutLen = 19

		default:
			f.OutLen = f.Len
		}

		le := len(f.Name)
		if f.OutLen > 0 && f.OutLen < le {
			f.OutLen = le
		}
	}
	q.closed = false
	q.Lfd = 0
	return true
}

// Fetch Sql
func (q *SQLX) Fetch() (readed bool) {
	readed = q.r.Next()
	if readed {
		q.Lfd++
		q.Err = q.r.Scan(q.scanValues...)
	}

	return readed
}

// Exec #
func (q *SQLX) Exec(sq string) bool {
	//	log.Printf("sq: [%s]", sq)

	q.prepareStmt(sq)

	if (q.StmtType & StmtOutf) == StmtOutf {
		return q.execSelect(sq)
	}

	_, q.Err = q.db.Exec(sq)
	if q.Err != nil {
		return false
	}

	return true
}

// Close Sql
func (q *SQLX) Close() error {
	if !q.closed {
		q.closed = true
		q.Lfd = 0
		q.Err = q.r.Err()
		q.r.Close()

		q.Fields = q.Fields[:0]
		q.ColNum = 0
		q.scanValues = q.scanValues[:0]

		return q.Err
	}

	return nil
}

// ShowLine #
func (q *SQLX) ShowLine(isTitle bool) string {
	s := ""

	if isTitle {
		for i, f := range q.Fields {
			if i > 0 {
				s += " "
			}
			s = s + f.FormattedTitle()
		}
	} else {
		for i, f := range q.Fields {
			if i > 0 {
				s += " "
			}
			s = s + f.FormattedValue()
		}
	}

	if q.NewLine {
		s += "\n"
	}

	return s
}

// ShowLineAsEcv QLIne As ecv
func (q *SQLX) ShowLineAsEcv(isTitle bool) string {
	s := ""

	if isTitle {
		s = "@" + q.QName
		for _, f := range q.Fields {
			s = s + "," + f.Name
			switch f.Typ {
			case "VARYING", "VARCHAR":
				s = s + "[char_" + strconv.Itoa(f.Len) + "]"
			case "SHORT":
				s = s + "[short]"
			case "INT", "LONG":
				s = s + "[int]"
			case "TIMESTAMP", "DATETIME":
				s = s + "[timestamp]"
			case "TEXT":
				if f.Len > 0 {
					s = s + "[char_" + strconv.Itoa(f.Len) + "]"
				} else {
					s = s + "[str]"
				}
			default:
				s = s + "[str]"
			}
		}
	} else {
		for i, f := range q.Fields {
			if i > 0 {
				s = s + "^"
			}

			if q.Replacer != nil {
				s = s + q.Replacer.Replace(f.AsString())
			} else {
				s = s + f.AsString()
			}
		}
	}

	if q.NewLine {
		s += "\n"
	}

	return s
}

// AsDelimitedText #
func (q *SQLX) AsDelimitedText(delimiter string) ([]string, error) {

	var ss []string

	s := ""
	for _, f := range q.Fields {
		s = s + f.Name + delimiter
	}
	ss = append(ss, s)

	for q.Fetch() {
		if q.Err != nil {
			return nil, q.Err
		}

		s = ""
		for _, f := range q.Fields {
			s = s + string(f.Value) + delimiter
		}
		ss = append(ss, s)
	}

	return ss, nil
}

// AsJSON #
func (q *SQLX) AsJSON() ([]byte, error) {
	// an array of JSON objects
	// the map key is the field name
	var objects []map[string]interface{}

	for q.Fetch() {
		if q.Err != nil {
			return nil, q.Err
		}

		values := make([]interface{}, q.ColNum)
		object := map[string]interface{}{}
		for i, f := range q.Fields {
			v := reflect.New(f.rTyp).Interface()
			switch v.(type) {
			case *int32, *int16:
				ii := f.AsInteger()
				object[f.Name] = ii
			default:
				ss := f.AsString()
				object[f.Name] = ss
			}

			values[i] = object[f.Name]
		}

		objects = append(objects, object)
	}
	q.Close()
	return json.MarshalIndent(objects, "", "\t")
}

// PrintOut #
func (q *SQLX) PrintOut(w io.Writer) (err error) {

	if q.Lfd == 1 {
		st := ""
		for i, f := range q.Fields {
			if i > 0 {
				st += " "
			}
			st = st + f.FormattedTitle()
		}
		fmt.Fprintf(w, "%v\n", st)
	}

	s := ""
	for i, f := range q.Fields {
		if i > 0 {
			s += " "
		}
		s = s + f.FormattedValue()
	}
	fmt.Fprintf(w, "%v\n", s)

	return
}

// PrintTo #
func (q *SQLX) PrintTo(w io.Writer, frm string) (err error) {
	/*	for i, f := range q.Fields {
			fmt.Fprintf(w, "%d. %-20s, len=%2d, typ=%v\n", i+1, f.Name, f.OutLen, f.Typ)
		}
	*/

	if frm == "json" {
		b, _ := q.AsJSON()
		if q.Err != nil {
			fmt.Fprintf(w, "asJSON: %v\n", q.Err)
			return q.Err
		}

		fmt.Fprintf(w, string(b))
		return nil
	}

	if frm == "table" {
		s := q.ShowLine(true)
		fmt.Fprintf(w, "%v", s)
		for q.Fetch() {
			if q.Err != nil {
				return q.Err
			}
			s := q.ShowLine(false)
			fmt.Fprintf(w, "%v", s)
		}
		return nil
	}

	delimiter := ";"
	s := ""
	for i, f := range q.Fields {
		if i > 0 {
			s = s + delimiter
		}
		s = s + f.Name
	}
	fmt.Fprintf(w, "%v\n", s)

	for q.Fetch() {
		if q.Err != nil {
			return q.Err
		}

		s = ""
		for i, f := range q.Fields {
			if i > 0 {
				s = s + delimiter
			}

			s = s + string(f.Value)
		}
		fmt.Fprintf(w, "%v\n", s)
	}

	return nil
}

// AsString (columnIdx)
func (q *SQLX) AsString(ix int) string {
	return q.Fields[ix].AsString()
}

// AsInteger (columnIdx)
func (q *SQLX) AsInteger(ix int) int {
	return q.Fields[ix].AsInteger()
}

// AsInt64 (columnIdx)
func (q *SQLX) AsInt64(ix int) int64 {
	return q.Fields[ix].AsInt64()
}

// Value (columnIdx)
func (q *SQLX) Value(ix int) []byte {
	return q.Fields[ix].Value
}

// AsDateTime #
func (q *SQLX) AsDateTime(ix int) *time.Time {
	return q.Fields[ix].AsDateTime()
}
