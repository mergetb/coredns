package nex

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"

	"gitlab.com/mergetb/nex"
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

	from := state.IP()
	qname := strings.Trim(state.QName(), ".")
	log.Infof("nex: name=%s from=%s", qname, from)
	var rr dns.RR

	switch state.Family() {
	case 1:
		rr = &dns.A{}
		rr.(*dns.A).Hdr = dns.RR_Header{
			Name:   state.QName(),
			Rrtype: dns.TypeA,
			Class:  state.QClass(),
		}

		addrs, err := nex.ResolveName(qname)
		if err != nil {
			return -1, fmt.Errorf("Failed to resolve name - %v", err)
		}
		log.Infof("addrs=%#v", addrs)
		rr.(*dns.A).A = net.ParseIP(addrs.Ip4).To4()
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

	a.Answer = []dns.RR{rr, srv}
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
