package diskio

/*
#cgo LDFLAGS: -lperfstat
#include <string.h>
#include <unistd.h>
#include <libperfstat.h>
*/
import "C"

import (
	"github.com/WuerthIT/perfstatbeat/helper/odm"
	"unsafe"
)

type diskLayout struct {
	stats    []C.perfstat_disk_t
	first    *C.perfstat_id_t
	udid_map map[string]string
}

func gatherDiskLayout() (*diskLayout, error) {
	first := new(C.perfstat_id_t)
	C.strcpy((*C.char)(unsafe.Pointer(&first.name)), C.CString(C.FIRST_DISK))

	num := C.perfstat_disk(nil, nil, C.sizeof_perfstat_disk_t, 0)
	stats := make([]C.perfstat_disk_t, num)

	udid_map, err := odm.GetAttributeMap("unique_id")

	return &diskLayout{
		stats:    stats,
		first:    first,
		udid_map: udid_map,
	}, err
}

func (m *MetricSet) updateDiskLayout() error {
	m.logger.Debug("updating disk layout")
	layout, err := gatherDiskLayout()
	if err != nil {
		if _, ok := err.(odm.OdmError); ok {
			m.logger.Error(err.Error())
		}
		return err
	}
	m.mux.Lock()
	m.layout = layout
	m.mux.Unlock()
	return nil
}
