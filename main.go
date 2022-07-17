package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"

	"github.com/miekg/dns"
)

var records = DNSFilterMap{
    "baguncinha.com": "69.69.69.69",
    "gugou.com": "142.251.128.14",
}

func main() {
    sd := NewSimpleDNS("lo", "1.1.1.1:53")
    sd.Filters = []DNSFilter{records}
    sd.Run()
}

type DNSFilter interface {
    LookupIPV4(host string) *net.IP
}

type DNSFilterMap map[string]string

func (f DNSFilterMap) LookupIPV4(host string) *net.IP {
    ipStr, ok := f[host]
    if !ok {
        return nil
    }
    ip := net.ParseIP(ipStr)
    return &ip
}


func NewSimpleDNS(prefix string, upstreams ...string) *SimpleDNS {
    ret := &SimpleDNS{
        Filters: []DNSFilter{},
        Upstreams: upstreams,
        prefix: prefix,
    }
    return ret
}

type SimpleDNS struct {
    Filters []DNSFilter
    Upstreams []string
    prefix string
}

func (s *SimpleDNS) Run() {
    server := &dns.Server{
        Addr: "127.0.0.1:53",
        Net: "udp",
    }
    dns.HandleFunc(fmt.Sprintf("%s.", s.prefix), s.handleRequest)
    dns.HandleFunc(".", s.handleFallback)
    server.ListenAndServe()
}

func (s *SimpleDNS) handleFallback(w dns.ResponseWriter, r *dns.Msg) {
    c := new(dns.Client)
    c.Net = "udp"
    if len(r.Question) > 0 {
        q := r.Question[0]
        log.Printf("fallback query: %s", q.String())
    }
    upstream := s.Upstreams[rand.Intn(len(s.Upstreams))]
    ret, _, err := c.Exchange(r, upstream)
    msg := new(dns.Msg)
    msg.Answer = ret.Answer
    if err != nil {
        log.Printf("upstream failed: %s", err.Error())
    }
    msg.SetReply(r)
    w.WriteMsg(msg)
}

func (s *SimpleDNS) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
    m := new(dns.Msg)
    m.SetReply(r)
    for _, q := range m.Question {
        log.Printf("Query: %s", q.String())
        switch q.Qtype {
        case dns.TypeA:
            domain := q.Name
            domain = strings.Trim(domain, ".")
            domain = strings.TrimSuffix(domain, s.prefix)
            domain = strings.Trim(domain, ".")
            log.Printf("Query: %s", domain)
            var ip *net.IP
            for _, filter := range s.Filters {
                ip = filter.LookupIPV4(domain)
                if ip != nil {
                    break
                }
            }
            if ip != nil {
                log.Printf("found ip for %s: %s", domain, ip)
                m.Answer = append(m.Answer, &dns.A{
                    Hdr: dns.RR_Header{
                        Name: q.Name, 
                        Rrtype: dns.TypeA,
                        Class: dns.ClassINET,
                        Ttl: 60,
                    },
                    A: *ip,
                })
            }
        }
    }

    // if len(m.Answer) == 0 {
    //     m.SetRcode(r, dns.RcodeNameError)
    // }
    w.WriteMsg(m)
    // log.Printf("%s", r.String())
}
