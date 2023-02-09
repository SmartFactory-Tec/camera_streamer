package gst

import "C"
import (
	"runtime"
)

type RefCounter interface {
	ref()
	unref()
}

func enableGarbageCollection(ref RefCounter) {
	runtime.SetFinalizer(ref, func(object RefCounter) {
		ref.unref()
	})
}
