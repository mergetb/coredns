# Nex: Scalable Name and Address Software Infrastructure

Nex is the spirit of dnsmasq meets modern scale out architecture. The design is inspired and motivated by the recent success Facebook has had with deploying Kea using a stateless replecated model. Nex is based on a similar stateless service model for both Kea and CoreDNS - with an etcd deployment that brings them together into a cohesive system - ala dnsmasq.

The DNS side of things is a fork of CoreDNS plugin and is maintained in this repository

https://github.com/mergetb/coredns