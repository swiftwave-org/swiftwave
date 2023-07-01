package main

import (
	"errors"

	"github.com/domainr/dnsr"
)

func fetchDNSRecord(qname string, qtype string) ([]string, error){
	r := dnsr.NewResolver(dnsr.WithCache(10))
	result := []string{}
	for _, rr := range r.Resolve(qname, qtype) {
	  if rr.Type == qtype {
		result = append(result, rr.Value)
	  }
	}
	if len(result) > 0 {
		return result, nil
	}
	return result, errors.New("no record found")
}