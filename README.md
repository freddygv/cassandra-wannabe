Cassandra Wannabe
=======
[![Go Report Card](https://goreportcard.com/badge/github.com/freddygv/cassandra-wannabe)](https://goreportcard.com/report/github.com/freddygv/cassandra-wannabe) 
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/freddygv/cassandra-wannabe/blob/master/LICENSE)

### What?
Distributed data store **experiment** for MovieLens ratings dataset. Schema is hard-coded to focus on other areas.

### Core Goals
* Consistent hashing with at most 1-hop routing
* Data replication on 2 nodes with Cassandra's SimpleStrategy
* Expose REST api for upsert/read/delete ✅
* gRPC CRUD service for the nodes
* Observability instrumentation with OpenCensus.io
* Metrics exported to Prometheus
* Traces exported to Zipkin

### Stretch Goals
* Virtual nodes instead of real nodes
* Basic implementation of Gossip protocol with UDP packets
