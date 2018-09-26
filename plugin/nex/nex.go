package nex

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("nex")

type Nex struct {
	Next plugin.Handler
}

func (x Nex) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (
	int, error) {

	log.Debug("received request")

	state := request.Request{W: w, Req: r}

	a := &dns.Msg{}
	a.SetReply(r)
	a.Compress = true
	a.Authoritative = true

	//ip := state.IP()
	var rr dns.RR

	switch state.Family() {
	case 1:
		rr = &dns.A{}
		rr.(*dns.A).Hdr = dns.RR_Header{
			Name:   state.QName(),
			Rrtype: dns.TypeA,
			Class:  state.QClass(),
		}
		rr.(*dns.A).A = net.ParseIP("1.2.3.4").To4()
	}

	srv := &dns.SRV{}
	srv.Hdr = dns.RR_Header{
		Name:   "_" + state.Proto() + "." + state.QName(),
		Rrtype: dns.TypeSRV,
		Class:  state.QClass(),
	}
	port, _ := strconv.Atoi(state.Port())
	srv.Port = uint16(port)
	srv.Target = "."

	a.Extra = []dns.RR{rr, srv}
	state.SizeAndDo(a)
	w.WriteMsg(a)

	pw := NewResponsePrinter(w)

	return plugin.NextOrFailure(x.Name(), x.Next, ctx, pw, r)

}

func (x Nex) Name() string {
	return "nex"
}

type ResponsePrinter struct {
	dns.ResponseWriter
}

func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	fmt.Fprintln(out, ex)
	return r.ResponseWriter.WriteMsg(res)
}

var out io.Writer = os.Stdout

const ex = "nex"
