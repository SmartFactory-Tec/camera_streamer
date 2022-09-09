package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>

static GOptionEntry* allocOptionEntries(int entryCount) {
	return calloc(entryCount, sizeof(GOptionEntry));
}
*/
import "C"
import (
	"unsafe"
)

func Init() {
	argc, argv := createCArgs()
	// the gst C library is responsible for freeing this memory
	//defer freeCArgs(argv)

	C.gst_init(&argc, &argv)
}

func InitWithOptions(options []OptionsEntry) {
	// Create options context
	ctx := C.g_option_context_new(C.CString(""))
	defer C.g_option_context_free(ctx)

	// Size of options array is one more than required, to leave the last array element null
	var gOptionsPtr *C.GOptionEntry = C.allocOptionEntries(C.int(len(options) + 1))
	defer C.free(unsafe.Pointer(gOptionsPtr))

	// Get a go slice referencing the options array
	//
	gOptions := unsafe.Slice(gOptionsPtr, len(options))

	for i, option := range options {
		cEntry := option.cEntry()
		gOptions[i] = cEntry
	}

	// Add the entries and option group
	C.g_option_context_add_main_entries(ctx, gOptionsPtr, nil)
	C.g_option_context_add_group(ctx, C.gst_init_get_option_group())

	argc, argv := createCArgs()
	// the gst C library is responsible for freeing this memory
	//defer freeCArgs(argv)

	// Parse the context and cli args
	C.g_option_context_parse(ctx, &argc, &argv, nil)
}
