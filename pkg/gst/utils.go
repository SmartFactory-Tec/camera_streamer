package gst

/*
#include <stdlib.h>

static char** allocArgv(int argc) {
    return malloc(sizeof(char *) * argc);
}
*/
import "C"
import (
	"os"
	"unsafe"
)

func createCArgs() (argc C.int, argv **C.char) {
	argc = C.int(len(os.Args))
	argv = C.allocArgv(argc)
	slice := unsafe.Slice(argv, len(os.Args))

	for i, arg := range os.Args {
		slice[i] = C.CString(arg)
	}
	return
}

func freeCArgs(argv **C.char) {
	C.free(unsafe.Pointer(argv))
}
