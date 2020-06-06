package script

// ----------------------------------------------------------------------------------
// parse.go for Go's dbx.script package
// Copyright 2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------
// HISTORY
// ----------------------------------------------------------------------------------
// 2020.05.23 init
// ----------------------------------------------------------------------------------

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
)

// Parser #
type Parser struct {
	Token []*Token
}

// Token #
type Token struct {
	ID         TokenID
	Key        string
	Fields     []Field
	nextCmdIdx int
	Cmds       []([]string)
}

// Field #
type Field struct {
	Key string
	ID  TokenID
}

// NewParser #
func NewParser() *Parser {
	return &Parser{}
}

// LoadFile #
func (x *Parser) LoadFile(fname string) error {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	return x.load(&b)
}

func (x *Parser) load(b *[]byte) error {
	lines := strings.Split(strings.Replace(string(*b), "\r", "", -1), "\n")

	var cTok *Token

	for i := 0; i < len(lines); i++ {
		r := []rune(lines[i])
		le := skipRight(&r)

		aix := 0
		skipLeft(&aix, &r, le)

		if aix >= le {
			continue
		}

		if r[aix] == '#' {
			continue
		}

		eol := false
		if le > 1 && r[le-1] == '&' && r[le-2] == '&' {
			eol = true
			le -= 2
			r = r[:le]
		}

		if eol && le < 1 {
			continue
		}

		var tk *Token
		var lk string

		if r[aix] == '$' {
			tk, _, lk = getDollarToken(r[aix:le])
			if tk.ID == TkFi {
				debug("#decode: cTok.ENDFI")

				cTok = nil
				continue
			}

			if tk.ID == TkEcvStop {
				debug("#decode: cTok.ENDECV")
				cTok = nil
				continue
			}

			if cTok == nil {
				cTok = &Token{}
				*cTok = *tk
				x.Token = append(x.Token, cTok)
				debug("\n#decode: cTok.New #", len(x.Token))

				cTok.nextCmdIdx = 0
			}

			if tk == nil {
				return errors.New("error line #" + strconv.Itoa(i+1) + " bad token: " + lk)
			}

			if !eol && tk.ID < TkEOL {
				eol = true
			}
			debug("#decode #", len(x.Token), "T=[", string(r[aix:le]), "]: cTok.ID=", cTok.ID, "tk.ID=", tk.ID, "eol=", eol)

			if tk.ID == TkIf || tk.ID == TkEcvStart {
				continue
			}

			if tk.ID == TkOneIf || tk.ID == TkOneNotIf {
				aixo := aix
				getNextWordIdx(&aix, &r, le)
				r = r[aix:]
				le = le - (aix - aixo)
			}
		} else {
			if cTok == nil {
				cTok = &Token{TkAny, "sql", []Field{}, 0, []([]string){}}
				x.Token = append(x.Token, cTok)
				debug("\n#decode: cTok.New.Any #", len(x.Token))
				cTok.nextCmdIdx = 0
			}
		}

		if eol {
			cTok.Add(string(r[:le]))
			cTok.nextCmdIdx++

			if tk != nil && tk.ID < TkEOL {
				debug("#decode: cTok.ENDA")
				cTok = nil
				continue
			}

			if cTok.ID == TkOneIf || cTok.ID == TkOneNotIf || cTok.ID == TkAny {
				debug("#decode: cTok.ENDX")
				cTok = nil
			}
			//}
		} else {
			cTok.Add(string(r[:le]))
		}
	}

	return nil
}

// Add #
func (x *Token) Add(s string) {
	if len(x.Cmds) == x.nextCmdIdx {
		x.Cmds = append(x.Cmds, []string{})
		debug("#decode.addCmd #", x.nextCmdIdx+1, ":", s)
	}
	x.Cmds[x.nextCmdIdx] = append(x.Cmds[x.nextCmdIdx], s)
}

// GetData #
func (x *Token) GetData(ix int) string {
	ss := ""
	if ix < len(x.Cmds) {
		for _, s := range x.Cmds[ix] {
			if ss == "" {
				ss = s
			} else {
				ss = ss + "\n" + s
			}
		}

	}

	return ss
}

// FieldKeyVal #
func (x *Token) FieldKeyVal() (int, string) {
	op := TkNone
	val := ""
	// => eql_1.txt ID= 1 $app_version len(Fields)= 2 tkf: [{= 23} {NSF 0}]
	le := len(x.Fields)
	if le > 0 {
		op = int(x.Fields[0].ID)
	}

	if le > 1 {
		val = x.Fields[1].Key
	}

	return op, val
}

// FieldIfExist #
func (x *Token) FieldIfExist() (int, bool, int, string) {
	neg := false
	op := TkNone
	typ := TkNone
	val := ""

	for _, f := range x.Fields {
		switch f.ID {
		case TkTable,
			TkIndex,
			TkFunction,
			TkTrigger,
			TkProcedure:
			typ = int(f.ID)

		case TkNot:
			neg = true
		case TkExist:
			op = TkExist
		case TkNone:
			if val == "" {
				val = f.Key
			}
		}
	}

	return op, neg, typ, val
}

// FieldIE #
// => $ine tkf: [{CREATE 32} {TABLE 44} {xFile 0} {( 30}]
// => $ie  tkf: [{DROP 33} {TABLE 44} {endFile 0}]
// => $ine tkf: [{CREATE 32} {UNIQUE 48} {INDEX 45} {ix_amFields 0} {on 47} {amFields 0} {( 30} {name 0} {) 31}]
func (x *Token) FieldIE() (int, int, string) {
	op := TkNone
	typ := TkNone
	val := ""
	tbl := ""
	onField := 0

	for _, f := range x.Fields {
		switch f.ID {
		case TkTable,
			TkIndex,
			TkFunction,
			TkTrigger,
			TkProcedure:
			typ = int(f.ID)

		case TkOn:
			onField++
		case TkCreate:
			op = TkCreate
		case TkDrop:
			op = TkDrop
		case TkNone:
			if onField == 0 && val == "" {
				val = f.Key
			}

			if onField == 1 && tbl == "" {
				tbl = f.Key
			}
		}
	}

	if tbl != "" {
		val = tbl + "." + val
	}

	return op, typ, val
}

func getDollarToken(s []rune) (*Token, bool, string) {
	fi := splitLine(s)
	if len(fi) < 1 {
		return nil, false, ""
	}

	k := strings.ToLower(fi[0])
	if k[0] == '$' {

		if c, ok := cmds[k]; ok {
			tok := &Token{c, k, []Field{}, 0, []([]string){}}
			for _, w := range fi[1:] {
				if c, ok := cmds[strings.ToLower(w)]; ok {
					tok.Fields = append(tok.Fields, Field{w, c})
				} else {
					tok.Fields = append(tok.Fields, Field{w, TkNone})
				}
			}

			// sql-fields updateten
			for i := 0; i < len(tok.Fields); i++ {
				lo := strings.ToLower(tok.Fields[i].Key)
				if q, ok := sqls[lo]; ok {
					tok.Fields[i].ID = q
				}
			}

			return tok, true, k
		}
	}

	return nil, true, k
}

func splitLine(rs []rune) []string {
	type span struct {
		start int
		end   int
	}
	spans := make([]span, 0, 32)

	fromIndex := 0
	le := len(rs)
	i := 0
	for i < le {
		ok := false
		skipLeft(&i, &rs, le)
		fromIndex = i

		// check word
		for i < le && isWord(rs[i]) {
			i++
		}

		if fromIndex != i {
			ok = true
			spans = append(spans, span{fromIndex, i})
		}

		// check operator
		if !ok {
			for i < le && isOperator(rs[i]) {
				i++
			}

			if fromIndex != i {
				ok = true
				spans = append(spans, span{fromIndex, i})
				fromIndex = i
			}
		}

		// check bracket
		if !ok {
			for i < le && isBracket(rs[i]) {
				i++
			}

			if fromIndex != i {
				ok = true
				spans = append(spans, span{fromIndex, i})
				fromIndex = i
			}
		}

		if !ok {
			i++
		}
	}

	a := make([]string, len(spans))
	for i, span := range spans {
		a[i] = string(rs[span.start:span.end])
	}

	return a
}

func hasPrefix(s *string, le int, pfx string) bool {
	lx := len(pfx)
	return le >= lx && (*s)[0:lx] == pfx
}

func hasSuffix(s *string, le int, sfx string) bool {
	lx := len(sfx)
	return le >= lx && (*s)[le-lx:] == sfx
}

func skipLeft(i *int, r *[]rune, le int) bool {
	for *i < le && isWhitespace((*r)[*i]) {
		(*i)++
	}

	return *i >= le
}

func skipRight(r *[]rune) int {
	i := len(*r) - 1
	for i >= 0 && isWhitespace((*r)[i]) {
		i--
	}

	return i + 1
}

func getNextWordIdx(i *int, r *[]rune, le int) {
	for *i < le && isWord((*r)[*i]) {
		(*i)++
	}

	skipLeft(i, r, le)
}

func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }
func isLetter(ch rune) bool     { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }
func isDigit(ch rune) bool      { return (ch >= '0' && ch <= '9') }

func isWord(ch rune) bool {
	return ch == '.' || ch == '$' || ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

func isOperator(ch rune) bool { return ch == '!' || ch == '>' || ch == '<' || ch == '=' }
func isBracket(ch rune) bool  { return ch == '(' || ch == ')' || ch == '[' || ch == ']' }

func debug(a ...interface{}) {
	//	fmt.Fprintln(os.Stdout, a...)
}
