package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func main() {
    host := "ep-crimson-truth-a5ttv8vs.us-east-2.aws.neon.tech"

    // Try to resolve using different DNS servers
    configs := []string{"8.8.8.8:53", "1.1.1.1:53"}

    for _, dnsServer := range configs {
        r := &net.Resolver{
            PreferGo: true,
            Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                d := net.Dialer{
                    Timeout: time.Second * 10,
                }
                return d.Dial(network, dnsServer)
            },
        }

        ips, err := r.LookupHost(context.Background(), host)
        if err != nil {
            fmt.Printf("Error resolving with %s: %v\n", dnsServer, err)
            continue
        }

        fmt.Printf("Successfully resolved %s using %s:\n", host, dnsServer)
        for _, ip := range ips {
            fmt.Printf("  IP: %s\n", ip)
        }
    }
}