package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
)

const prefix = "lo"

func main() {
    server := &dns.Server{
        Addr: "127.0.0.1:53",
        Net: "udp",
    }
    dns.HandleFunc(fmt.Sprintf("%s.", prefix), handleRequest)
    dns.HandleFunc(".", handleFallback)
    server.ListenAndServe()
}

type DNSFilter interface {
    LookupIPV4(host string) net.IP
}

var records = map[string]string{
    "baguncinha.com": "69.69.69.69",
    "gugou.com": "142.251.128.14",
}

func handleFallback(w dns.ResponseWriter, r *dns.Msg) {
    c := new(dns.Client)
    c.Net = "udp"
    if len(r.Question) > 0 {
        q := r.Question[0]
        log.Printf("fallback query: %s", q.String())
    }
    ret, _, err := c.Exchange(r, "1.1.1.1:53")
    msg := new(dns.Msg)
    msg.Answer = ret.Answer
    if err != nil {
        log.Printf("upstream failed: %s", err.Error())
    }
    msg.SetReply(r)
    w.WriteMsg(msg)
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
    m := new(dns.Msg)
    m.SetReply(r)
    for _, q := range m.Question {
        log.Printf("Query: %s", q.String())
        switch q.Qtype {
        case dns.TypeA:
            domain := q.Name
            domain = strings.Trim(domain, ".")
            domain = strings.TrimSuffix(domain, prefix)
            domain = strings.Trim(domain, ".")
            log.Printf("Query: %s", domain)
            ip, ok := records[domain]
            if ok {
                log.Printf("found ip for %s: %s", domain, ip)
                m.Answer = append(m.Answer, &dns.A{
                    Hdr: dns.RR_Header{
                        Name: q.Name, 
                        Rrtype: dns.TypeA,
                        Class: dns.ClassINET,
                        Ttl: 60,
                    },
                    A: net.ParseIP(ip),
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
