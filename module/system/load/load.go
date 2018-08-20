package load

/*
#cgo LDFLAGS: -lperfstat
#include <libperfstat.h>
*/
import "C"

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/cfgwarn"
	"github.com/elastic/beats/metricbeat/mb"
)

// init registers the MetricSet with the central registry as soon as the program
// starts. The New function will be called later to instantiate an instance of
// the MetricSet for each host defined in the module's configuration. After the
// MetricSet has been created then Fetch will begin to be called periodically.
func init() {
	mb.Registry.MustAddMetricSet("system", "load", New)
}

// MetricSet holds any configuration or state information. It must implement
// the mb.MetricSet interface. And this is best achieved by embedding
// mb.BaseMetricSet because it implements all of the required mb.MetricSet
// interface methods except for Fetch.
type MetricSet struct {
	mb.BaseMetricSet
	cpustat *C.perfstat_cpu_total_t
}

// New creates a new instance of the MetricSet. New is responsible for unpacking
// any MetricSet specific configuration options if there are any.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	cfgwarn.Experimental("The system load metricset is experimental.")

	config := struct{}{}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	return &MetricSet{
		BaseMetricSet: base,
		cpustat:       new(C.perfstat_cpu_total_t),
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch(report mb.ReporterV2) {
	C.perfstat_cpu_total(nil, m.cpustat, C.sizeof_perfstat_cpu_total_t, 1)

	load_factor := float64(1 << C.SBITS)
	cores_factor := float64(m.cpustat.ncpus)
	load_1 := float64(m.cpustat.loadavg[0]) / load_factor
	load_5 := float64(m.cpustat.loadavg[1]) / load_factor
	load_15 := float64(m.cpustat.loadavg[2]) / load_factor

	report.Event(mb.Event{
		MetricSetFields: common.MapStr{
			"1":     load_1,
			"5":     load_5,
			"15":    load_15,
			"cores": m.cpustat.ncpus,
			"norm": common.MapStr{
				"1":  load_1 / cores_factor,
				"5":  load_5 / cores_factor,
				"15": load_15 / cores_factor,
			},
		},
	})
}
