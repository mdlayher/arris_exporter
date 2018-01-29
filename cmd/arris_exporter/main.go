// Command arris_exporter implements a Prometheus exporter for Arris cable
// modem devices.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/mdlayher/arris"
	"github.com/mdlayher/arris_exporter"
)

func main() {
	var (
		metricsAddr = flag.String("metrics.addr", ":9393", "address for Arris exporter")
		metricsPath = flag.String("metrics.path", "/metrics", "URL path for surfacing collected metrics")

		arrisTimeout = flag.Duration("arris.timeout", 5*time.Second, "timeout value for requests to an Arris cable modem; use 0 for no timeout")
	)

	flag.Parse()

	// dial is the function used to connect to an Arris device on each
	// metrics scrape request.
	dial := func(addr string) (*arris.Client, error) {
		return arris.New(addr, &http.Client{
			Timeout: *arrisTimeout,
		})
	}

	h := arrisexporter.NewHandler(dial)

	mux := http.NewServeMux()
	mux.Handle(*metricsPath, h)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})

	log.Printf("starting Arris exporter on %q", *metricsAddr)

	if err := http.ListenAndServe(*metricsAddr, mux); err != nil {
		log.Fatalf("cannot start Arris exporter: %v", err)
	}
}
