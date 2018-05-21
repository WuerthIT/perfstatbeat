package diskio

// #cgo LDFLAGS: -lperfstat
// #include <string.h>
// #include <unistd.h>
// #include <libperfstat.h>
import "C"

import (
	"reflect"
	"unsafe"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/cfgwarn"
	"github.com/elastic/beats/metricbeat/mb"
)

// init registers the MetricSet with the central registry as soon as the program
// starts. The New function will be called later to instantiate an instance of
// the MetricSet for each host defined in the module's configuration. After the
// MetricSet has been created then Fetch will begin to be called periodically.
func init() {
	mb.Registry.MustAddMetricSet("system", "diskio", New)
}

// MetricSet holds any configuration or state information. It must implement
// the mb.MetricSet interface. And this is best achieved by embedding
// mb.BaseMetricSet because it implements all of the required mb.MetricSet
// interface methods except for Fetch.
type MetricSet struct {
	mb.BaseMetricSet
	stats []C.perfstat_disk_t
	first *C.perfstat_id_t
	buffer *C.perfstat_disk_t
	num C.int
	sc_clk_tck uint64
	xint uint64
	xfrac uint64
}

// New creates a new instance of the MetricSet. New is responsible for unpacking
// any MetricSet specific configuration options if there are any.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	cfgwarn.Experimental("The system diskio metricset is experimental.")

	config := struct{}{}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	var first_data C.perfstat_id_t
	C.strcpy((*C.char)(unsafe.Pointer(&first_data.name)), C.CString(C.FIRST_DISK))

	num :=  C.perfstat_disk(nil, nil, C.sizeof_perfstat_disk_t, 0)
	stats := make([]C.perfstat_disk_t, num, num)

	// get a pointer to the array backing the newly created slice
	buffer := (*C.perfstat_disk_t)(unsafe.Pointer(((*reflect.SliceHeader)(unsafe.Pointer(&stats))).Data))

	return &MetricSet{
		BaseMetricSet: base,
		stats: stats,
		first: &first_data,
		buffer: buffer,
		num: num,
		sc_clk_tck: uint64(C.sysconf(C._SC_CLK_TCK)),
		xint: uint64(C._system_configuration.Xint),
		xfrac: uint64(C._system_configuration.Xfrac),
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch() ([]common.MapStr, error) {
	C.perfstat_disk(m.first, m.buffer, C.sizeof_perfstat_disk_t, m.num)

	events := make([]common.MapStr, 0, len(m.stats))
	for _, counters := range m.stats {

		event := common.MapStr{
			"name": C.GoString((*C.char)(unsafe.Pointer(&counters.name))),
			"vgname": C.GoString((*C.char)(unsafe.Pointer(&counters.vgname))),
			"read": common.MapStr{
				"count": counters.xrate,
				"bytes": counters.rblks * counters.bsize,
				"time": uint64(counters.rserv) * m.xint / m.xfrac / 1e+6,
			},
			"write": common.MapStr{
				"count": counters.xfers - counters.xrate,
				"bytes": counters.wblks * counters.bsize,
				"time": uint64(counters.wserv) * m.xint / m.xfrac / 1e+6,
			},
			"io": common.MapStr{
				"time": uint64(counters.time) * 1000 / m.sc_clk_tck,
			},
		}

		events = append(events, event)
	}

	return events, nil
}
