package arrisexporter

import (
	"fmt"
	"net"
	"net/http"

	"github.com/mdlayher/arris"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ http.Handler = &handler{}

// A handler is an http.Handler that serves Prometheus metrics for
// arris devices.
type handler struct {
	dial func(addr string) (*arris.Client, error)
}

// NewHandler returns an http.Handler that serves Prometheus metrics for
// arris devices. The dial function specifies how to connect to a
// device with the specified address on each HTTP request.
//
// Each HTTP request must contain a "target" query parameter which indicates
// the network address of the device which should be scraped for metrics.
// If no port is specified, the arris device default of 65001 will be used.
func NewHandler(dial func(addr string) (*arris.Client, error)) http.Handler {
	return &handler{
		dial: dial,
	}
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Prometheus is configured to send a target parameter with each scrape
	// request. This determines which device should be scraped for metrics.
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	host, port, err := net.SplitHostPort(target)
	if err != nil {
		// Assume no port was provided and use the default.
		host = target
		port = "80"
	}

	addr := net.JoinHostPort(host, port)

	c, err := h.dial("http://" + addr)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("failed to dial arris device at %q: %v", addr, err),
			http.StatusInternalServerError,
		)
		return
	}

	metrics := serveMetrics(c)
	metrics.ServeHTTP(w, r)
}

// serveMetrics creates a Prometheus metrics handler for a Device.
func serveMetrics(c *arris.Client) http.Handler {
	reg := prometheus.NewRegistry()
	reg.MustRegister(newCollector(c))

	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
