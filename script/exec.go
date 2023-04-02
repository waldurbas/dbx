package script

// ----------------------------------------------------------------------------------
// exec.go for Go's dbx.script package
// Copyright 2020,2023 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------
// HISTORY
// ----------------------------------------------------------------------------------
// 2023.04.02 ExistDom,ExistExc
// 2020.07.19 TkField abfragen
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
	Vinfo          VersInfo
	ExecCmd        func(a int, ix int, cmd string) (bool, error)
	ExecEcv        func(a *Lines) error
	ExistTrigger   func(sName string) bool
	ExistTable     func(sName string) bool
	ExistTableCol  func(sName string) bool
	ExistIndex     func(sName string) bool
	ExistProc      func(sName string) bool
	ExistFunc      func(sName string) bool
	ExistDomain    func(sName string) bool
	ExistException func(sName string) bool
	SaveVers       func(v int) error
}

// NewScript #
func NewScript() *DbScript {
	return &DbScript{Vinfo: VersInfo{App: "none", Hide: true}}
}

// LoadFile #
func (dbs *DbScript) LoadFile(fileName string) (*Parser, error) {

	px := NewParser()

	err := px.LoadFile(fileName)
	if err != nil {
		return nil, err
	}

	return px, nil
}

// Execute #
func (dbs *DbScript) Execute(px *Parser) (int, error) {
	ndbu := 0
	nupd := 0
	xdbu := 0
	cmdID := 0
	a := 0
	lastDBU := -dbs.Vinfo.Dbu

	for _, tk := range px.Token {
		ok := false
		debug("token:", tk.ID, Token2String(int(tk.ID)))

		switch tk.ID {
		// APP_VERSION
		case TkAppVersion:
			op, val := tk.FieldKeyVal()
			if op != TkEQ {
				return a, errors.New("Parser.App: bad operator")
			}

			if val != dbs.Vinfo.App {
				return a, errors.New("Parser.App: wrong database")
			}
			continue

		case TkExit:
			cmdID = TkExit
			ok = true

		case TkEcho:
			cmdID = TkEcho
			ok = true

		case TkSet:
			cmdID = TkSet
			ok = true

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

			if xdbu > dbs.Vinfo.Dbu {
				dbs.Vinfo.Dbu = xdbu
				lastDBU = xdbu
			} else if xdbu < dbs.Vinfo.Dbu {
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
				ok = dbs.Vinfo.Dbu == xdbu
			case TkGT:
				ok = (dbs.Vinfo.Dbu/100 == ndbu) && (dbs.Vinfo.Dbu%100 == nupd-1)
				if !ok {
					ok = (dbs.Vinfo.Dbu/100 == ndbu-1) && (nupd == 0)
				}
			case TkGE:
				ok = dbs.Vinfo.Dbu == xdbu

				if !ok {
					ok = (dbs.Vinfo.Dbu/100 == ndbu) && (dbs.Vinfo.Dbu%100 == nupd-1)
				}

				if !ok {
					ok = (dbs.Vinfo.Dbu/100 == ndbu-1) && (nupd == 0)
				}
			}

			if !ok {
				return a, errors.New("Parser.Dbu: bad version")
			}

			continue

		case TkShow:
			dbs.Vinfo.Show = true
			continue
		case TkNoShow:
			dbs.Vinfo.Show = false
			continue
		case TkHide:
			dbs.Vinfo.Hide = true
			continue
		case TkNoHide:
			dbs.Vinfo.Hide = false
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
						ok = dbs.ExistTableCol(val)
					} else {
						ok = dbs.ExistTable(val)
					}
				case TkField:
					ss := strings.Split(val, ".")
					if len(ss) == 2 {
						ok = dbs.ExistTableCol(val)
					}
				case TkIndex:
					ok = dbs.ExistIndex(val)
				case TkTrigger:
					ok = dbs.ExistTrigger(val)
				case TkFunction:
					ok = dbs.ExistFunc(val)
				case TkProcedure:
					ok = dbs.ExistProc(val)
				case TkException:
					ok = dbs.ExistException(val)
				case TkDomain:
					ok = dbs.ExistDomain(val)

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
			ao, typ, val := tk.FieldIE()
			cmdID = ao

			switch typ {
			case TkTable:
				ss := strings.Split(val, ".")
				if len(ss) == 2 {
					ok = dbs.ExistTableCol(val)
				} else {
					ok = dbs.ExistTable(val)
				}
			case TkIndex:
				ok = dbs.ExistIndex(val)
			case TkTrigger:
				ok = dbs.ExistTrigger(val)
			case TkFunction:
				ok = dbs.ExistFunc(val)
			case TkProcedure:
				ok = dbs.ExistProc(val)
			case TkField:
				ok = dbs.ExistTableCol(val)
			case TkException:
				ok = dbs.ExistException(val)
			case TkDomain:
				ok = dbs.ExistDomain(val)

			default:
				return a, errors.New("Parser.IF: bad object")
			}

			if tk.ID == TkOneNotIf {
				ok = !ok
			}

		case TkEcv:
			if dbs.ExecEcv != nil {
				lin := tk.Cmds2Data()

				err := dbs.ExecEcv(lin)
				if err != nil {
					return a, err
				}
			}
			cmdID = TkNone
			a += len(tk.Cmds[0]) + 2
			continue
		default:
			ok = true
		}

		if ok {
			for i := 0; i < len(tk.Cmds); i++ {
				sq := tk.GetData(i)
				end, err := dbs.ExecCmd(cmdID, i, sq)
				cmdID = TkNone
				if err != nil {
					return a, err
				}
				if end {
					ok = false
					break
				}
			}

			if !ok {
				break
			}
		}
		a++
	}

	if lastDBU > 0 {
		err := dbs.SaveVers(lastDBU)
		if err != nil {
			return a, err
		}
	} else {
		dbs.Vinfo.Dbu = -lastDBU
	}

	return a, nil
}
