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
}

func handleFallback(w dns.ResponseWriter, r *dns.Msg) {
    log.Printf("fallback")
    c := new(dns.Client)
    c.Net = "udp"
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
                m.Answer = append(m.Answer, &dns.A{
                    Hdr: dns.RR_Header{
                        Name: domain, 
                        Rrtype: dns.TypeA,
                        Class: dns.ClassINET,
                        Ttl: 60,
                    },
                    A: net.ParseIP(ip),
                })
            }
        }
    }

    log.Printf("%d", len(m.Answer))
    // if len(m.Answer) == 0 {
    //     m.SetRcode(r, dns.RcodeNameError)
    // }
    w.WriteMsg(m)
    // log.Printf("%s", r.String())
}
