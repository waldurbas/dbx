package script

// ----------------------------------------------------------------------------------
// const.go for Go's dbx.script package
// Copyright 2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices. Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------
// HISTORY
// ----------------------------------------------------------------------------------
// 2023.04.02 TkException,TkDomain
// 2020.09.14 tokArray, func init(), Token2String
// 2020.05.23 init
// ----------------------------------------------------------------------------------

// TokenID #
type TokenID int

// Token consts
const (
	TkNone = iota
	TkAppVersion
	TkDbuVersion
	TkDbu
	TkDbuStart
	TkDbuEnd
	TkDbuLast
	TkEcho
	TkShow
	TkNoShow
	TkHide
	TkNoHide
	// 12:
	TkComment
	TkFi
	TkSet

	// 15: wichtig wegen IID < TkEOL
	TkEOL
	TkExit
	TkEcvStart
	TkEcvStop
	TkOneIf
	TkOneNotIf
	TkIf

	// :22
	TkExist
	TkNot
	TkNE
	TkEQ
	TkGT
	TkLT
	TkGE
	TkLE
	TkEOF
	TkBracketOpen

	// :32
	TkBracketClose
	TkCreate
	TkDrop

	// :35
	TkAdd
	TkSelect
	TkUpdate
	TkRecreate
	TkAscending
	TkDescending
	TkPrimary
	TkKey
	TkDelete
	TkTable

	// :45
	TkField
	TkProcedure
	TkIndex
	TkTrigger
	TkOn
	TkTo
	TkFirst
	TkAfter
	TkUnique
	TkForeign
	TkFunction
	TkModify
	TkAlter
	TkRename
	TkException
	TkDomain
	TkEcv
	TkAny
)

const (
	scrTypNone = 0
	scrTypCmd  = 1
	scrTypSQL  = 2
)

var tokArray = []struct {
	txt string
	iid TokenID
	typ int
}{
	{"$ie", TkOneIf, scrTypCmd},
	{"$ine", TkOneNotIf, scrTypCmd},
	{"$if", TkIf, scrTypCmd},
	{"$fi", TkFi, scrTypCmd},
	{"$endif", TkFi, scrTypCmd},
	{"$ecv", TkEcv, scrTypNone},
	{"$ecv_start", TkEcvStart, scrTypCmd},
	{"$ecv_stop", TkEcvStop, scrTypCmd},
	{"$dbu_start", TkDbuStart, scrTypCmd},
	{"$dbu_end", TkDbuEnd, scrTypCmd},
	{"$lastdbu", TkDbuLast, scrTypCmd},
	{"$dbu", TkDbu, scrTypCmd},
	{"$show", TkShow, scrTypCmd},
	{"$noshow", TkNoShow, scrTypCmd},
	{"$set", TkSet, scrTypCmd},
	{"$hide", TkHide, scrTypCmd},
	{"$nohide", TkNoHide, scrTypCmd},
	{"$exit", TkExit, scrTypCmd},
	{"$echo", TkEcho, scrTypCmd},
	{"$drop", TkDrop, scrTypCmd},
	{"$app_version", TkAppVersion, scrTypCmd},
	{"$dbu_version", TkDbuVersion, scrTypCmd},
	{"#", TkComment, scrTypCmd},
	{"//", TkComment, scrTypCmd},
	{"&&", TkEOL, scrTypCmd},
	{"exist", TkExist, scrTypCmd},
	{"recreate", TkRecreate, scrTypCmd},
	{"field", TkField, scrTypCmd},
	{"!", TkNot, scrTypCmd},
	{"!=", TkNE, scrTypCmd},
	{"=", TkEQ, scrTypCmd},
	{">", TkGT, scrTypCmd},
	{"<", TkLT, scrTypCmd},
	{">=", TkGE, scrTypCmd},
	{"<=", TkLE, scrTypCmd},
	{")", TkBracketClose, scrTypCmd | scrTypSQL},
	{"(", TkBracketOpen, scrTypCmd | scrTypSQL},
	{"add", TkAdd, scrTypSQL},
	{"exists", TkExist, scrTypSQL},
	{"not", TkNot, scrTypSQL},
	{"table", TkTable, scrTypSQL},
	{"column", TkField, scrTypSQL},
	{"modify", TkModify, scrTypSQL},
	{"alter", TkAlter, scrTypSQL},
	{"rename", TkRename, scrTypSQL},
	{"procedure", TkProcedure, scrTypSQL},
	{"function", TkFunction, scrTypSQL},
	{"index", TkIndex, scrTypSQL},
	{"trigger", TkTrigger, scrTypSQL},
	{"exception", TkException, scrTypSQL},
	{"domain", TkDomain, scrTypSQL},
	{"on", TkOn, scrTypSQL},
	{"to", TkTo, scrTypSQL},
	{"first", TkFirst, scrTypSQL},
	{"after", TkAfter, scrTypSQL},
	{"create", TkCreate, scrTypSQL},
	{"drop", TkDrop, scrTypSQL},
	{"select", TkSelect, scrTypSQL},
	{"update", TkUpdate, scrTypSQL},
	{"delete", TkDelete, scrTypSQL},
	{"ascending", TkAscending, scrTypSQL},
	{"descending", TkDescending, scrTypSQL},
	{"primary", TkPrimary, scrTypSQL},
	{"foreign", TkForeign, scrTypSQL},
	{"unique", TkUnique, scrTypSQL},
	{"key", TkKey, scrTypSQL},
	{"if", TkIf, scrTypSQL},
	{"#any", TkAny, scrTypNone},
}

var sqls = make(map[string]TokenID)
var cmds = make(map[string]TokenID)

func debug(a ...interface{}) {
	// fmt.Fprint(os.Stdout, "-> ")
	// fmt.Fprintln(os.Stdout, a...)
}

func init() {
	for _, e := range tokArray {
		if (e.typ & scrTypCmd) > 0 {
			cmds[e.txt] = TokenID(e.iid)
		}

		if (e.typ & scrTypSQL) > 0 {
			cmds[e.txt] = TokenID(e.iid)
		}
	}
}

// Token2String #
func Token2String(itok int) string {
	for _, e := range tokArray {
		if int(e.iid) == itok {
			return e.txt
		}
	}

	return "none"
}
