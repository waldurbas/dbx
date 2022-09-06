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
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SqxField #
type SqxField struct {
	Idx    int
	Name   string
	Typ    string
	Len    int
	OutLen int
	Value  []byte
	rTyp   reflect.Type
	Q      *SQLX
}

// IsNull #
func (f *SqxField) IsNull() bool {
	return f.Value == nil
}

// AsInt64 #
func (f *SqxField) AsInt64() int64 {
	if f.Value == nil {
		return 0
	}

	i, e := strconv.ParseInt(fmt.Sprintf("%v", string(f.Value)), 10, 64)
	if e != nil {
		return 0
	}

	return i
}

// AsInteger #
func (f *SqxField) AsInteger() int {
	if f.Value == nil {
		return 0
	}

	i, e := strconv.Atoi(fmt.Sprintf("%v", string(f.Value)))
	if e != nil {
		return 0
	}

	return i
}

// AsString #
func (f *SqxField) AsString() string {
	if f.Value == nil {
		return "NULL"
	}

	switch f.Typ {
	case "TIMESTAMP", "DATETIME":
		s := string(f.Value)
		s = strings.ReplaceAll(s, "T", " ")
		if len(s) > 19 {
			s = s[:19]
		}

		return s
	case "DATE":
		s := string(f.Value)
		if len(s) > 10 {
			s = s[:10]
		}
		return s
	}

	return strings.Trim(string(f.Value), " ")
}

// AsDateTime #
func (f *SqxField) AsDateTime() *time.Time {
	if f.Value == nil {
		return nil
	}

	s := string(f.Value)
	s = strings.ReplaceAll(s, "T", " ")
	if len(s) > 19 {
		s = s[:19]
	}
	layout := "2006-01-02 15:04:05" //Z07:00"
	t, _ := time.Parse(layout, s)

	return &t
}

// AsDate #
func (f *SqxField) AsDate() *time.Time {
	if f.Value == nil {
		return nil
	}

	s := string(f.Value)
	s = strings.ReplaceAll(s, "T", " ")
	if len(s) > 10 {
		s = s[:10]
	}
	layout := "2006-01-02"
	t, _ := time.Parse(layout, s)

	return &t
}

// FormattedTitle Value As Formatted String
func (f *SqxField) FormattedTitle() string {
	switch f.Typ {
	case "SHORT", "INT", "MEDIUMINT", "TINYINT", "BIGINT":
		return fmt.Sprintf("%*.*s", f.OutLen, f.OutLen, f.Name)
	default:
		if f.OutLen > 0 && f.Idx < (f.Q.ColNum-1) {
			return fmt.Sprintf("%-*.*s", f.OutLen, f.OutLen, f.Name)
		}

		return f.Name
	}
}

// FormattedValue Value As Formatted String
func (f *SqxField) FormattedValue() string {
	switch f.Typ {
	case "SHORT", "INT", "MEDIUMINT", "TINYINT", "BIGINT":
		return fmt.Sprintf("%*.*s", f.OutLen, f.OutLen, f.AsString())
	default:
		if f.OutLen > 0 && f.Idx < (f.Q.ColNum-1) {
			return fmt.Sprintf("%-*.*s", f.OutLen, f.OutLen, f.AsString())
		}

		return f.AsString()
	}
}

func (f *SqxField) CleanedString() string {
	if f.IsNull() {
		return "NIL"
	} else {
		switch f.Typ {
		case "SHORT", "INT", "MEDIUMINT", "TINYINT", "BIGINT":
			return prepareIntField(f.AsString())
		}
	}
	return prepareStringField(f.AsString())
}

// PrintOut Formatted Value auf stdout
func (f *SqxField) PrintOut() {
	fmt.Print(f.FormattedValue(), " ")
}

func prepareStringField(s string) string {
	rr := make([]rune, len(s)+2)

	rr[0] = rune('"')
	n := 1
	for _, c := range s {
		r := rune(c)
		switch r {
		case '\n':
			rr[n] = '/'
			n++
		case '\r':
		case '^':
		case '"':
			rr[n] = '\''
			n++
		default:
			rr[n] = r
			n++
		}
	}
	rr[n] = rune('"')

	return string(rr[:n])
}

func prepareIntField(s string) string {
	rr := make([]rune, len(s))

	n := 0
	for _, c := range s {
		r := rune(c)
		if r >= '0' && r <= '9' {
			rr[n] = r
			n++
		}
	}

	return string(rr[:n])
}
