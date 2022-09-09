package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"

import (
	"go/types"
	"unsafe"
)

type Argument interface {
	getType() C.GOptionArg
	getValue() unsafe.Pointer
}

type OptionsEntry struct {
	longName       string
	shortName      byte
	flags          int
	argument       Argument
	description    string
	argDescription string
}

func (entry *OptionsEntry) cEntry() C.GOptionEntry {
	return C.GOptionEntry{C.CString(entry.longName), C.char(entry.shortName), C.int(entry.flags), entry.argument.getType(), C.gpointer(entry.argument.getValue()), C.CString(entry.description), C.CString(entry.argDescription)}
}

type NoneArg types.Nil

func (arg *NoneArg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_NONE
}
func (arg *NoneArg) getValue() unsafe.Pointer {
	return nil
}

type StringArg string

func (arg *StringArg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_STRING
}

func (arg *StringArg) getValue() unsafe.Pointer {
	return unsafe.Pointer(arg)
}

type IntArg int

func (arg *IntArg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_INT
}

func (arg *IntArg) getValue() unsafe.Pointer {
	return unsafe.Pointer(arg)
}

type CallbackArg types.Func

func (arg *CallbackArg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_CALLBACK
}

func (arg *CallbackArg) getValue() unsafe.Pointer {
	return unsafe.Pointer(arg)
}

type FilenameArg string

func (arg *FilenameArg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_FILENAME
}

func (arg *FilenameArg) getValue() unsafe.Pointer {
	return unsafe.Pointer(arg)
}

type StringArrayArg []string

func (arg *StringArrayArg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_STRING_ARRAY
}

func (arg *StringArrayArg) getValue() unsafe.Pointer {
	return unsafe.Pointer(arg)
}

type FilenameArrayArg []string

func (arg *FilenameArrayArg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_FILENAME_ARRAY
}

func (arg *FilenameArrayArg) getValue() unsafe.Pointer {
	return unsafe.Pointer(arg)
}

type DoubleArg float64

func (arg *DoubleArg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_DOUBLE
}

func (arg *DoubleArg) getValue() unsafe.Pointer {
	return unsafe.Pointer(arg)
}

type Int64Arg int64

func (arg *Int64Arg) getType() C.GOptionArg {
	return C.G_OPTION_ARG_INT64
}

func (arg *Int64Arg) getValue() unsafe.Pointer {
	return unsafe.Pointer(arg)
}
