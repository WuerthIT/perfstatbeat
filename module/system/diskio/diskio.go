package diskio

// #cgo LDFLAGS: -lperfstat
// #include <stdlib.h>
// #include <string.h>
// #include <unistd.h>
// #include <libperfstat.h>
// u_longlong_t get_rxfers(perfstat_disk_t *stat) { return stat->__rxfers; }
// u_longlong_t ticks2ms(u_longlong_t ticks) { return ticks * _system_configuration.Xint / _system_configuration.Xfrac / 1e6; }
import "C"

import (
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
	first C.perfstat_id_t
	num C.int
	buffer *C.perfstat_disk_t
}

// New creates a new instance of the MetricSet. New is responsible for unpacking
// any MetricSet specific configuration options if there are any.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	cfgwarn.Experimental("The system diskio metricset is experimental.")

	config := struct{}{}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	var first C.perfstat_id_t
	C.strcpy((*C.char)(unsafe.Pointer(&first.name)), C.CString(C.FIRST_DISK))

	num :=  C.perfstat_disk(nil, nil, C.sizeof_perfstat_disk_t, 0)
	buffer := (*C.perfstat_disk_t)(C.malloc((C.size_t)(num * C.sizeof_perfstat_cpu_total_t)))

	return &MetricSet{
		BaseMetricSet: base,
		first: first,
		num: num,
		buffer: buffer,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch() ([]common.MapStr, error) {
	C.perfstat_disk(&m.first, m.buffer, C.sizeof_perfstat_disk_t, m.num)
	stats := (*[1 << 30]C.perfstat_disk_t)(unsafe.Pointer(m.buffer))[:m.num:m.num]

	events := make([]common.MapStr, 0, len(stats))
	for _, counters := range stats {

		rxfers := C.get_rxfers(&counters)

		event := common.MapStr{
			"name": C.GoString((*C.char)(unsafe.Pointer(&counters.name))),
			"vgname": C.GoString((*C.char)(unsafe.Pointer(&counters.vgname))),
			"read": common.MapStr{
				"count": rxfers,
				"bytes": counters.rblks * counters.bsize,
				"time": C.ticks2ms(counters.rserv),
			},
			"write": common.MapStr{
				"count": counters.xfers - rxfers,
				"bytes": counters.wblks * counters.bsize,
				"time": C.ticks2ms(counters.wserv),
			},
			"io": common.MapStr{
				"time": uint64(counters.time) * 1000 / uint64(C.sysconf(C._SC_CLK_TCK)),
			},
		}

		events = append(events, event)
	}

	return events, nil
}
