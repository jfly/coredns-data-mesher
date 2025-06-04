package data_mesher

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"path"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("data-mesher")

type DataMesher struct {
	stateDir string

	Next plugin.Handler
}

type HostEntry struct {
	Hostname string   `json:"hostname"`
	IPs      []string `json:"ips"`
}

func (dm DataMesher) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Export metric with the server label set to the current server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	state := request.Request{W: w, Req: r}

	dataMesherJsonPath := path.Join(dm.stateDir, "dns.json")
	_, err := os.Stat(dataMesherJsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("%s/dns.json not found", dm.stateDir)
			return dns.RcodeServerFailure, err
		}
	}

	file, err := os.Open(dataMesherJsonPath)
	if err != nil {
		log.Errorf("Error opening file: %v", err)
		return dns.RcodeServerFailure, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		var entry HostEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			log.Errorf("Error parsing JSON line: %s\nError: %v", line, err)
			continue
		}

		if entry.Hostname+"." == state.Name() {
			m := new(dns.Msg)
			m.SetReply(r)
			m.Authoritative = true

			// Data Mesher only supports AAAA records.
			if state.QType() == dns.TypeAAAA {
				for _, ip := range entry.IPs {
					rr := &dns.AAAA{
						Hdr: dns.RR_Header{
							Name:   state.Name(),
							Rrtype: dns.TypeAAAA,
							Class:  dns.ClassINET,
							Ttl:    60,
						},
						AAAA: net.ParseIP(ip).To16(),
					}
					m.Answer = append(m.Answer, rr)
				}
			}

			w.WriteMsg(m)
			return dns.RcodeSuccess, nil
		}
	}

	// We couldn't find the given hostname. Forward onto the next plugin (if any).
	return plugin.NextOrFailure("data-mesher", dm.Next, ctx, w, r)
}

// Name implements the Handler interface.
func (e DataMesher) Name() string { return "data-mesher" }

// ResponsePrinter wrap a dns.ResponseWriter and will write "data-mesher" to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// `WriteMsg` calls the underlying `ResponseWriters`'s `WriteMsg` method and prints "data-mesher" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("data-mesher")
	return r.ResponseWriter.WriteMsg(res)
}
