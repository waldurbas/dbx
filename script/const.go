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
	TkAny
)

// Cmds #
var cmds = map[string]TokenID{
	"$ie":          TkOneIf,
	"$ine":         TkOneNotIf,
	"$if":          TkIf,
	"$fi":          TkFi,
	"$endif":       TkFi,
	"$ecv_start":   TkEcvStart,
	"$ecv_stop":    TkEcvStop,
	"$dbu_start":   TkDbuStart,
	"$dbu_end":     TkDbuEnd,
	"$lastdbu":     TkDbuLast,
	"$dbu":         TkDbu,
	"$show":        TkShow,
	"$noshow":      TkNoShow,
	"$set":         TkSet,
	"$hide":        TkHide,
	"$nohide":      TkNoHide,
	"$exit":        TkExit,
	"$echo":        TkEcho,
	"$drop":        TkDrop,
	"$app_version": TkAppVersion,
	"$dbu_version": TkDbuVersion,
	"#":            TkComment,
	"//":           TkComment,
	"&&":           TkEOL,
	"exist":        TkExist,
	"recreate":     TkRecreate,
	"field":        TkField,
	"!":            TkNot,
	"!=":           TkNE,
	"=":            TkEQ,
	">":            TkGT,
	"<":            TkLT,
	">=":           TkGE,
	"<=":           TkLE,
	")":            TkBracketClose,
	"(":            TkBracketOpen,
}

// Sqls #
var sqls = map[string]TokenID{
	"add":        TkAdd,
	"exists":     TkExist,
	"not":        TkNot,
	"table":      TkTable,
	"column":     TkField,
	"modify":     TkModify,
	"alter":      TkAlter,
	"rename":     TkRename,
	"procedure":  TkProcedure,
	"function":   TkFunction,
	"index":      TkIndex,
	"trigger":    TkTrigger,
	"on":         TkOn,
	"to":         TkTo,
	"first":      TkFirst,
	"after":      TkAfter,
	"create":     TkCreate,
	"drop":       TkDrop,
	"select":     TkSelect,
	"update":     TkUpdate,
	"delete":     TkDelete,
	"ascending":  TkAscending,
	"descending": TkDescending,
	"primary":    TkPrimary,
	"foreign":    TkForeign,
	"unique":     TkUnique,
	"key":        TkKey,
	"if":         TkIf,
	")":          TkBracketClose,
	"(":          TkBracketOpen,
}
