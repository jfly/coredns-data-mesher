package data_mesher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	golog "log"
	"os"
	"path"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/stretchr/testify/assert"

	"github.com/miekg/dns"
)

type TestContext struct {
	dataMesher  DataMesher
	stateDir    string
	logOutput   *bytes.Buffer
	ogLogOutput io.Writer
}

var ErrNextHandler = fmt.Errorf("the next handler was called")

func NewTestContext(t *testing.T) *TestContext {
	stateDir := t.TempDir()
	dm := DataMesher{
		stateDir: stateDir,
		Next:     test.NextHandler(dns.RcodeServerFailure, ErrNextHandler),
	}

	// Setup a new output buffer that is *not* standard output, so we can check if
	// example is really being printed.
	b := &bytes.Buffer{}
	golog.SetOutput(b)
	ogOutput := golog.Writer()

	return &TestContext{
		dataMesher:  dm,
		stateDir:    stateDir,
		logOutput:   b,
		ogLogOutput: ogOutput,
	}
}

func (tc *TestContext) Close() {
	golog.SetOutput(tc.ogLogOutput)
}

func TestJsonMissing(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	testCase := test.Case{
		Qname: "example.org.",
		Qtype: dns.TypeAAAA,

		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("@ 60	IN	A 127.0.0.1"),
		},
	}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	rCode, err := tc.dataMesher.ServeDNS(context.Background(), rec, testCase.Msg())
	assert.ErrorContains(t, err, "dns.json: no such file or directory")
	assert.Equal(t, dns.RcodeServerFailure, rCode)
	assert.Contains(t, tc.logOutput.String(), fmt.Sprintf("[ERROR] plugin/data-mesher: %s/dns.json not found", tc.stateDir))
}

func TestSuccessfulQueries(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Close()

	{
		jsonPath := path.Join(tc.stateDir, "dns.json")
		err := os.WriteFile(jsonPath, []byte(`
			{"hostname":"example.org","ips":["2001:db8::1", "2001:db8::2"]}
			{"hostname":"s1.example.org","ips":["2001:db8::3"]}
		`), 0o644)
		assert.NoError(t, err)
	}

	testCases := []test.Case{
		{
			Qname: "example.org.",
			Qtype: dns.TypeAAAA,

			Rcode: dns.RcodeSuccess,
			Answer: []dns.RR{
				test.AAAA("example.org. 60	IN	AAAA 2001:db8::1"),
				test.AAAA("example.org. 60	IN	AAAA 2001:db8::2"),
			},
		},
		{
			Qname: "s1.example.org.",
			Qtype: dns.TypeAAAA,

			Rcode: dns.RcodeSuccess,
			Answer: []dns.RR{
				test.AAAA("s1.example.org. 60	IN	AAAA 2001:db8::3"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("query type=%d name=%s", testCase.Qtype, testCase.Qname), func(t *testing.T) {
			rec := dnstest.NewRecorder(&test.ResponseWriter{})
			_, err := tc.dataMesher.ServeDNS(context.Background(), rec, testCase.Msg())
			assert.NoError(t, err)

			assert.Equal(t, "", tc.logOutput.String()) // Verify nothing was logged.
			assert.NoError(t, test.SortAndCheck(rec.Msg, testCase))
		})
	}
}

func TestQueryWrongType(t *testing.T) {
	// Query for a domain that DataMesher knows about, but the wrong type
	// (here we query for an A record when there are only AAAA records).
	// This should succeed and return no answers.

	tc := NewTestContext(t)
	defer tc.Close()

	{
		jsonPath := path.Join(tc.stateDir, "dns.json")
		err := os.WriteFile(jsonPath, []byte(`
			{"hostname":"example.org","ips":["2001:db8::1"]}
		`), 0o644)
		assert.NoError(t, err)
	}

	testCase := test.Case{
		Qname: "example.org.",
		Qtype: dns.TypeA,

		Rcode:  dns.RcodeSuccess,
		Answer: []dns.RR{},
	}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	_, err := tc.dataMesher.ServeDNS(context.Background(), rec, testCase.Msg())
	assert.NoError(t, err)

	assert.Equal(t, "", tc.logOutput.String()) // Verify nothing was logged.
	assert.NoError(t, test.SortAndCheck(rec.Msg, testCase))
}

func TestQueryUnknownDomain(t *testing.T) {
	// Query for a domain that Data Mesher doesn't know about. This should
	// forward onto the next plugin in the CoreDNS chain.

	tc := NewTestContext(t)
	defer tc.Close()

	{
		jsonPath := path.Join(tc.stateDir, "dns.json")
		err := os.WriteFile(jsonPath, []byte(""), 0o644)
		assert.NoError(t, err)
	}

	testCase := test.Case{
		Qname: "unknown.example.org.",
		Qtype: dns.TypeAAAA,
	}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	rCode, err := tc.dataMesher.ServeDNS(context.Background(), rec, testCase.Msg())
	assert.Equal(t, dns.RcodeServerFailure, rCode)
	assert.ErrorIs(t, err, ErrNextHandler)
}
