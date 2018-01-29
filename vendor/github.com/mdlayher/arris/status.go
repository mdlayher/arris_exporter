package arris

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Status is the status of an Arris modem.
type Status struct {
	Downstream []Downstream
	Upstream   []Upstream
	Uptime     time.Duration
	Interfaces []Interface
}

// Downstream indicates the status of a downstream connection.
type Downstream struct {
	Name          string  //
	DCID          int     //
	Frequency     float64 // MHz
	Power         float64 // dBmV
	SNR           float64 // dB
	Modulation    string  //
	Octets        uint64  //
	Corrected     uint64  //
	Uncorrectable uint64  //
}

// Upstream indicates the status of an upstream connection.
type Upstream struct {
	Name        string  //
	UCID        int     //
	Frequency   float64 // MHz
	Power       float64 // dBmV
	ChannelType string  //
	SymbolRate  int     // kSym/s
	Modulation  string  //
}

// An Interface indicates the status of a network interface.
type Interface struct {
	Name        string
	Provisioned bool
	Up          bool
	Speed       string
	MAC         net.HardwareAddr
}

// parse is the entry point for raw data to be parsed into Status.
func (s *Status) parse(rows [][]string) error {
	if len(rows) == 0 || len(rows[0]) == 0 {
		return errors.New("arris: no status rows available to parse")
	}

	// Some parsers will skip the first row of column names.
	switch rows[0][0] {
	// Downstream status data.
	case "DCID":
		return s.parseDownstream(rows[1:])
	// Upstream status data.
	case "UCID":
		return s.parseUpstream(rows[1:])
	// System information (starts with uptime).
	case "System Uptime:":
		return s.parseSystem(rows)
	case "Interface Name":
		return s.parseInterfaces(rows[1:])
	}

	return nil
}

// parseDownstream parses downstream status data from incoming rows.
func (s *Status) parseDownstream(rows [][]string) error {
	for _, r := range rows {
		if len(r) != 9 {
			return errors.New("arris: incorrect number of row elements for downstream")
		}

		// int and uint64 value parsing.
		ints := make([]uint64, 0, 4)
		for _, n := range []int{1, 6, 7, 8} {
			ii, err := strconv.ParseUint(r[n], 10, 64)
			if err != nil {
				return err
			}

			ints = append(ints, ii)
		}

		// float64 value parsing, stripping the unit suffix.
		floats := make([]float64, 0, 3)
		for _, n := range []int{2, 3, 4} {
			ss := strings.Fields(r[n])
			if len(ss) != 2 {
				return fmt.Errorf("arris: malformed downstream float64: %q", r[n])
			}

			ff, err := strconv.ParseFloat(ss[0], 64)
			if err != nil {
				return err
			}

			floats = append(floats, ff)
		}

		s.Downstream = append(s.Downstream, Downstream{
			Name:          r[0],
			DCID:          int(ints[0]),
			Frequency:     floats[0],
			Power:         floats[1],
			SNR:           floats[2],
			Modulation:    r[5],
			Octets:        ints[1],
			Corrected:     ints[2],
			Uncorrectable: ints[3],
		})
	}

	return nil
}

// parseUpstream parses upstream status data from incoming rows.
func (s *Status) parseUpstream(rows [][]string) error {
	for _, r := range rows {
		if len(r) != 7 {
			return errors.New("arris: incorrect number of row elements for upstream")
		}

		ucid, err := strconv.ParseUint(r[1], 10, 64)
		if err != nil {
			return err
		}

		// float64 value parsing, stripping the unit suffix.
		floats := make([]float64, 0, 3)
		for _, n := range []int{2, 3, 5} {
			ss := strings.Fields(r[n])
			if len(ss) != 2 {
				return fmt.Errorf("arris: malformed upstream float64: %q", r[n])
			}

			// Turn "n/a" into zero value.
			if ss[0] == "----" {
				ss[0] = "0.0"
			}

			ff, err := strconv.ParseFloat(ss[0], 64)
			if err != nil {
				return err
			}

			floats = append(floats, ff)
		}

		s.Upstream = append(s.Upstream, Upstream{
			Name:        r[0],
			UCID:        int(ucid),
			Frequency:   floats[0],
			Power:       floats[1],
			ChannelType: r[4],
			SymbolRate:  int(floats[2]),
			Modulation:  r[6],
		})
	}

	return nil
}

// parseSystem parses system status information.
func (s *Status) parseSystem(rows [][]string) error {
	for _, r := range rows {
		var err error
		switch r[0] {
		case "System Uptime:":
			err = s.parseUptime(r)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// parseUptime parses the system's uptime.
func (s *Status) parseUptime(row []string) error {
	if len(row) != 2 {
		return fmt.Errorf("arris: malformed uptime: %v", row)
	}

	// Example uptime format:
	// "0 d: 2 h: 43 m"
	fields := strings.Fields(row[1])

	// int and uint64 value parsing.
	durs := make([]time.Duration, 0, 3)
	for _, n := range []int{0, 2, 4} {
		d, err := strconv.ParseUint(fields[n], 10, 64)
		if err != nil {
			return err
		}

		durs = append(durs, time.Duration(d))
	}

	s.Uptime = (durs[0] * 24 * time.Hour) +
		(durs[1] * time.Hour) +
		(durs[2] * time.Minute)

	return nil
}

// parseInterfaces parses interface information from incoming rows.
func (s *Status) parseInterfaces(rows [][]string) error {
	for _, r := range rows {
		if len(r) != 5 {
			return errors.New("arris: incorrect number of row elements for interface")
		}

		speed := r[3]
		if speed == "-----" {
			speed = "n/a"
		}

		mac, err := net.ParseMAC(r[4])
		if err != nil {
			return err
		}

		s.Interfaces = append(s.Interfaces, Interface{
			Name: r[0],
			// TODO(mdlayher): determine if a string is actually more appropriate
			// for these.
			Provisioned: r[1] == "Enabled",
			Up:          r[2] == "Up",
			Speed:       speed,
			MAC:         mac,
		})
	}

	return nil
}
