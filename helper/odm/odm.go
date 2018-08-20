package odm

/*
#cgo LDFLAGS: -lodm -lcfg
#include <odmi.h>
#include <sys/cfgodm.h>
*/
import "C"

import (
	"regexp"
	"strings"
	"unsafe"
)

type odmError struct {
	odmerrno C.int
}

var whitespace = regexp.MustCompile("\\s+")

func (e odmError) Error() string {
	var message *C.char
	C.odm_err_msg(e.odmerrno, &message)
	return "ODM subroutine failed: " + strings.TrimSpace(whitespace.ReplaceAllLiteralString(C.GoString(message), " "))
}

func Get_attribute_map(attribute string) (map[string]string, error) {
	m := make(map[string]string)
	if C.odm_initialize() == 0 {
		defer C.odm_terminate()
		class := C.odm_open_class_rdonly((C.CLASS_SYMBOL)(unsafe.Pointer(&C.CuAt_CLASS)))
		if ^uintptr(unsafe.Pointer(class)) != 0 {
			defer C.odm_close_class(class)
			var obj C.struct_CuAt
			obj_ptr := unsafe.Pointer(&obj)
			query := C.CString("attribute = " + attribute)
			for res := C.odm_get_first(class, query, obj_ptr); res != nil; res = C.odm_get_next(class, obj_ptr) {
				if ^uintptr(res) != 0 {
					m[C.GoString(&obj.name[0])] = C.GoString(&obj.value[0])
				} else {
					return m, odmError{C.odmerrno}
				}
			}
			return m, nil
		}
	}
	return m, odmError{C.odmerrno}
}
