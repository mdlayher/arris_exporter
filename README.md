arris_exporter [![Build Status](https://travis-ci.org/mdlayher/arris_exporter.svg?branch=master)](https://travis-ci.org/mdlayher/arris_exporter) [![GoDoc](https://godoc.org/github.com/mdlayher/arris_exporter?status.svg)](https://godoc.org/github.com/mdlayher/arris_exporter) [![Go Report Card](https://goreportcard.com/badge/github.com/mdlayher/arris_exporter)](https://goreportcard.com/report/github.com/mdlayher/arris_exporter)
==============

Command `arris_exporter` implements a Prometheus exporter for Arris cable
modem devices.  MIT Licensed.

Configuration
-------------

The `arris_exporter`'s Prometheus scrape configuration (`prometheus.yml`) is
configured in a similar way to the official Prometheus
[`blackbox_exporter`](https://github.com/prometheus/blackbox_exporter).

The `targets` list under `static_configs` should specify the addresses of any
Arris devices which should be monitored by the exporter.  The address of
the `arris_exporter` itself must be specified in `relabel_configs` as well.

```yaml
scrape_configs:
  - job_name: 'arris'
    static_configs:
      - targets:
        - '192.168.100.1' # arris cable modem.
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: '127.0.0.1:9393' # arris_exporter.
```