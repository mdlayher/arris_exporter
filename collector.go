package arrisexporter

import (
	"github.com/mdlayher/arris"
	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &collector{}

// A collector is a prometheus.Collector for a device.
type collector struct {
	UptimeSecondsTotal *prometheus.Desc

	DownstreamPowerDBMV                 *prometheus.Desc
	DownstreamBytesTotal                *prometheus.Desc
	DownstreamCorrectedSymbolsTotal     *prometheus.Desc
	DownstreamUncorrectableSymbolsTotal *prometheus.Desc

	UpstreamPowerDBMV  *prometheus.Desc
	UpstreamSymbolRate *prometheus.Desc

	InterfacesInfo        *prometheus.Desc
	InterfacesProvisioned *prometheus.Desc
	InterfacesUp          *prometheus.Desc

	c *arris.Client
}

// newCollector constructs a collector using a device.
func newCollector(c *arris.Client) prometheus.Collector {
	return &collector{
		UptimeSecondsTotal: prometheus.NewDesc(
			"arris_uptime_seconds_total",
			"Device uptime in seconds.",
			nil,
			nil,
		),

		DownstreamPowerDBMV: prometheus.NewDesc(
			"arris_downstream_power_dbmv",
			"Current power level for the downstream connection in dBmV.",
			[]string{"name"},
			nil,
		),

		DownstreamBytesTotal: prometheus.NewDesc(
			"arris_downstream_bytes_total",
			"Number of downstream bytes total.",
			[]string{"name"},
			nil,
		),

		DownstreamCorrectedSymbolsTotal: prometheus.NewDesc(
			"arris_downstream_corrected_symbols_total",
			"Number of downstream corrected symbols total.",
			[]string{"name"},
			nil,
		),

		DownstreamUncorrectableSymbolsTotal: prometheus.NewDesc(
			"arris_downstream_uncorrectable_symbols_total",
			"Number of downstream uncorrectable symbols total.",
			[]string{"name"},
			nil,
		),

		UpstreamPowerDBMV: prometheus.NewDesc(
			"arris_upstream_power_dbmv",
			"Current power level for the upstream connection in dBmV.",
			[]string{"name"},
			nil,
		),

		UpstreamSymbolRate: prometheus.NewDesc(
			"arris_upstream_symbols_per_second",
			"Current symbol rate for the upstream connection in symbols per second.",
			[]string{"name"},
			nil,
		),

		InterfacesInfo: prometheus.NewDesc(
			"arris_interfaces_info",
			"Information about an interface.",
			[]string{"name", "speed", "mac"},
			nil,
		),

		InterfacesProvisioned: prometheus.NewDesc(
			"arris_interfaces_provisioned",
			"Whether or not a network interface is provisioned (0 - false, 1 - true).",
			[]string{"name"},
			nil,
		),

		InterfacesUp: prometheus.NewDesc(
			"arris_interfaces_up",
			"Whether or not a network interface is up (0 - false, 1 - true).",
			[]string{"name"},
			nil,
		),

		c: c,
	}
}

// Describe implements prometheus.Collector.
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.UptimeSecondsTotal,
		c.DownstreamPowerDBMV,
		c.DownstreamBytesTotal,
		c.DownstreamCorrectedSymbolsTotal,
		c.DownstreamUncorrectableSymbolsTotal,
		c.UpstreamPowerDBMV,
		c.UpstreamSymbolRate,
		c.InterfacesInfo,
		c.InterfacesProvisioned,
		c.InterfacesUp,
	}

	for _, d := range ds {
		ch <- d
	}
}

// Collect implements prometheus.Collector.
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	status, err := c.c.Status()
	if err != nil {
		ch <- prometheus.NewInvalidMetric(c.UptimeSecondsTotal, err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.UptimeSecondsTotal,
		prometheus.CounterValue,
		status.Uptime.Seconds(),
	)

	c.collectDownstream(ch, status.Downstream)
	c.collectUpstream(ch, status.Upstream)
	c.collectInterfaces(ch, status.Interfaces)
}

func (c *collector) collectDownstream(ch chan<- prometheus.Metric, downstream []arris.Downstream) {
	for _, ds := range downstream {
		pairs := []struct {
			d *prometheus.Desc
			t prometheus.ValueType
			v float64
		}{
			{d: c.DownstreamPowerDBMV, t: prometheus.GaugeValue, v: ds.Power},
			{d: c.DownstreamBytesTotal, v: float64(ds.Octets)},
			{d: c.DownstreamCorrectedSymbolsTotal, v: float64(ds.Corrected)},
			{d: c.DownstreamUncorrectableSymbolsTotal, v: float64(ds.Uncorrectable)},
		}

		for _, p := range pairs {
			if p.t == 0 {
				p.t = prometheus.CounterValue
			}

			ch <- prometheus.MustNewConstMetric(p.d, p.t, p.v, ds.Name)
		}
	}
}

func (c *collector) collectUpstream(ch chan<- prometheus.Metric, upstream []arris.Upstream) {
	for _, us := range upstream {
		pairs := []struct {
			d *prometheus.Desc
			v float64
		}{
			{d: c.UpstreamPowerDBMV, v: us.Power},
			// Convert kSym/s to sym/s.
			{d: c.UpstreamSymbolRate, v: float64(us.SymbolRate * 1000)},
		}

		for _, p := range pairs {
			ch <- prometheus.MustNewConstMetric(p.d, prometheus.GaugeValue, p.v, us.Name)
		}
	}
}

func (c *collector) collectInterfaces(ch chan<- prometheus.Metric, ifis []arris.Interface) {
	for _, ifi := range ifis {
		ch <- prometheus.MustNewConstMetric(
			c.InterfacesInfo,
			prometheus.GaugeValue,
			1.0,
			ifi.Name, ifi.Speed, ifi.MAC.String(),
		)

		pairs := []struct {
			d *prometheus.Desc
			b bool
		}{
			{d: c.InterfacesProvisioned, b: ifi.Provisioned},
			{d: c.InterfacesUp, b: ifi.Up},
		}

		for _, p := range pairs {
			var f float64
			if p.b {
				f = 1.0
			}

			ch <- prometheus.MustNewConstMetric(p.d, prometheus.GaugeValue, f, ifi.Name)
		}
	}
}
