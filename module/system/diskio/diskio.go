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
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/cfgwarn"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/metricbeat/mb"
	"unsafe"
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
	stats         []C.perfstat_disk_t
	first         *C.perfstat_id_t
	sc_clk_tck    uint64
	sc_xint       uint64
	sc_xfrac      uint64
	udid_map      map[string]string
	logger        *logp.Logger
	last_counters map[string]DiskCounters
}

type DiskCounters struct {
	read_count  uint64
	read_bytes  uint64
	read_time   uint64
	write_count uint64
	write_bytes uint64
	write_time  uint64
}

// New creates a new instance of the MetricSet. New is responsible for unpacking
// any MetricSet specific configuration options if there are any.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	cfgwarn.Experimental("The system diskio metricset is experimental.")

	logger := logp.NewLogger("diskio")

	config := struct{}{}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	first := new(C.perfstat_id_t)
	C.strcpy((*C.char)(unsafe.Pointer(&first.name)), C.CString(C.FIRST_DISK))

	num := C.perfstat_disk(nil, nil, C.sizeof_perfstat_disk_t, 0)
	stats := make([]C.perfstat_disk_t, num, num)
	last_counters := make(map[string]DiskCounters)

	udid_map, err := odm.Get_attribute_map("unique_id")
	if err != nil {
		logger.Error(err.Error())
	}

	return &MetricSet{
		BaseMetricSet: base,
		stats:         stats,
		first:         first,
		sc_clk_tck:    uint64(C.sysconf(C._SC_CLK_TCK)),
		sc_xint:       uint64(C._system_configuration.Xint),
		sc_xfrac:      uint64(C._system_configuration.Xfrac),
		udid_map:      udid_map,
		logger:        logger,
		last_counters: last_counters,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch() ([]common.MapStr, error) {
	C.perfstat_disk(m.first, &m.stats[0], C.sizeof_perfstat_disk_t, (C.int)(len(m.stats)))

	events := make([]common.MapStr, 0, len(m.stats))
	for _, counters := range m.stats {

		name := C.GoString(&counters.name[0])

		current := DiskCounters{
			read_count:  uint64(counters.xrate),
			read_bytes:  uint64(counters.rblks * counters.bsize),
			read_time:   uint64(counters.rserv) * m.sc_xint / m.sc_xfrac / 1e+6,
			write_count: uint64(counters.xfers - counters.xrate),
			write_bytes: uint64(counters.wblks * counters.bsize),
			write_time:  uint64(counters.wserv) * m.sc_xint / m.sc_xfrac / 1e+6,
		}

		event := common.MapStr{
			"name":   name,
			"vgname": C.GoString(&counters.vgname[0]),
			"udid":   m.udid_map[name],
			"read": common.MapStr{
				"count": current.read_count,
				"bytes": current.read_bytes,
				"time":  current.read_time,
			},
			"write": common.MapStr{
				"count": current.write_count,
				"bytes": current.write_bytes,
				"time":  current.write_time,
			},
			"io": common.MapStr{
				"time": uint64(counters.time) * 1000 / m.sc_clk_tck,
			},
		}

		if last, ok := m.last_counters[name]; ok {
			event["derivative"] = common.MapStr{
				"read": common.MapStr{
					"count": current.read_count - last.read_count,
					"bytes": current.read_bytes - last.read_bytes,
					"time":  current.read_time - last.read_time,
				},
				"write": common.MapStr{
					"count": current.write_count - last.write_count,
					"bytes": current.write_bytes - last.write_bytes,
					"time":  current.write_time - last.write_time,
				},
			}
		}
		m.last_counters[name] = current

		events = append(events, event)
	}

	return events, nil
}
