package script

// ----------------------------------------------------------------------------------
// exec.go for Go's dbx.script package
// Copyright 2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------
// HISTORY
// ----------------------------------------------------------------------------------
// 2020.06.06 New,VersInfo
// 2020.05.24 init
// ----------------------------------------------------------------------------------

import (
	"errors"
	"strconv"
	"strings"
)

// VersInfo #
type VersInfo struct {
	Dbu  int
	App  string
	Chg  string
	Show bool
	Hide bool
}

// DbScript #
type DbScript struct {
	Vinfo         VersInfo
	ExecCmd       func(a int, ix int, cmd string) (bool, error)
	ExistTable    func(sName string) bool
	ExistTableCol func(sName string) bool
	ExistIndex    func(sName string) bool
	ExistProc     func(sName string) bool
	ExistFunc     func(sName string) bool
	SaveVers      func(v int) error
}

// NewScript #
func NewScript() *DbScript {
	return &DbScript{Vinfo: VersInfo{App: "none", Hide: true}}
}

// LoadFile #
func (db *DbScript) LoadFile(fileName string) (*Parser, error) {

	px := NewParser()

	err := px.LoadFile(fileName)
	if err != nil {
		return nil, err
	}

	return px, nil
}

// Execute #
func (db *DbScript) Execute(px *Parser) (int, error) {
	ndbu := 0
	nupd := 0
	xdbu := 0
	a := 0

	for _, tk := range px.Token {
		ok := false

		switch tk.ID {
		// APP_VERSION
		case TkAppVersion:
			op, val := tk.FieldKeyVal()
			if op != TkEQ {
				return a, errors.New("Parser.App: bad operator")
			}

			if val != db.Vinfo.App {
				return a, errors.New("Parser.App: wrong database")
			}
			continue

		// LASTDBU
		case TkDbuLast:
			_, val := tk.FieldKeyVal()
			ss := strings.Split(val, ".")
			if len(ss) != 2 {
				return a, errors.New("Parser.LastDbu: bad value")
			}

			ndbu, _ = strconv.Atoi(ss[0])
			nupd, _ = strconv.Atoi(ss[1])
			xdbu = ndbu*100 + nupd

			if xdbu > db.Vinfo.Dbu {
				err := db.SaveVers(xdbu)
				if err != nil {
					return a, err
				}
				db.Vinfo.Dbu = xdbu
			} else if xdbu < db.Vinfo.Dbu {
				return a, errors.New("Parser.LastDbu: bad value")
			}
			continue

		// DBU_VERSION
		case TkDbuVersion:
			op, val := tk.FieldKeyVal()
			ss := strings.Split(val, ".")
			if len(ss) != 2 {
				return a, errors.New("Parser.Dbu: bad value")
			}
			ndbu, _ = strconv.Atoi(ss[0])
			nupd, _ = strconv.Atoi(ss[1])
			xdbu = ndbu*100 + nupd
			switch op {
			case TkEQ:
				ok = db.Vinfo.Dbu == xdbu
			case TkGT:
				ok = (db.Vinfo.Dbu/100 == ndbu) && (db.Vinfo.Dbu%100 == nupd-1)
				if !ok {
					ok = (db.Vinfo.Dbu/100 == ndbu-1) && (nupd == 0)
				}
			case TkGE:
				ok = db.Vinfo.Dbu == xdbu

				if !ok {
					ok = (db.Vinfo.Dbu/100 == ndbu) && (db.Vinfo.Dbu%100 == nupd-1)
				}

				if !ok {
					ok = (db.Vinfo.Dbu/100 == ndbu-1) && (nupd == 0)
				}
			}

			if !ok {
				return a, errors.New("Parser.Dbu: bad version")
			}

			continue

		case TkShow:
			db.Vinfo.Show = true
			continue
		case TkNoShow:
			db.Vinfo.Show = false
			continue
		case TkHide:
			db.Vinfo.Hide = true
			continue
		case TkNoHide:
			db.Vinfo.Hide = false
			continue

		// IF
		case TkIf:
			op, neg, typ, val := tk.FieldIfExist()

			switch op {
			case TkExist:
				switch typ {
				case TkTable:
					ss := strings.Split(val, ".")
					if len(ss) == 2 {
						ok = db.ExistTableCol(val)
					} else {
						ok = db.ExistTable(val)
					}
				case TkIndex:
					ok = db.ExistIndex(val)
				case TkFunction:
					ok = db.ExistFunc(val)
				case TkProcedure:
					ok = db.ExistProc(val)
				default:
					return a, errors.New("Parser.IF: bad object")
				}

				if neg {
					ok = !ok
				}
			default:
				return a, errors.New("Parser.IF: bad operator")
			}

		case TkOneIf, TkOneNotIf:
			_, typ, val := tk.FieldIE()

			switch typ {
			case TkTable:
				ss := strings.Split(val, ".")
				if len(ss) == 2 {
					ok = db.ExistTableCol(val)
				} else {
					ok = db.ExistTable(val)
				}
			case TkIndex:
				ok = db.ExistIndex(val)
			case TkFunction:
				ok = db.ExistFunc(val)
			case TkProcedure:
				ok = db.ExistProc(val)
			default:
				return a, errors.New("Parser.IF: bad object")
			}

			if tk.ID == TkOneNotIf {
				ok = !ok
			}

		default:
			ok = true
		}

		if ok {
			for i := 0; i < len(tk.Cmds); i++ {
				sq := tk.GetData(i)
				end, err := db.ExecCmd(a, i, sq)
				if err != nil {
					return a, err
				}
				if end {
					break
				}
			}
		}
		a++
	}

	if xdbu > db.Vinfo.Dbu {
		err := db.SaveVers(xdbu)
		if err != nil {
			return a, err
		}
		db.Vinfo.Dbu = xdbu
	}

	return a, nil
}
