Cassandra Wannabe
=======
[![Go Report Card](https://goreportcard.com/badge/github.com/freddygv/cassandra-wannabe)](https://goreportcard.com/report/github.com/freddygv/cassandra-wannabe) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ffreddygv%2Fcassandra-wannabe.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Ffreddygv%2Fcassandra-wannabe?ref=badge_shield)

[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/freddygv/cassandra-wannabe/blob/master/LICENSE)

### What?
Distributed data store **experiment** for MovieLens ratings dataset. Schema is hard-coded to focus on other areas.

### Core Goals
- [ ] Consistent hashing 
- [ ] At most 1-hop routing
- [ ] Data replication on 2 nodes with Cassandra's SimpleStrategy
- [x] Expose REST api for upsert/read/delete
- [x] gRPC CRUD service for the nodes
- [ ] Observability with OpenCensus.io, exporting to Prometheus and OpenZipkin

### Stretch Goals
- [ ] Basic implementation of Gossip protocol with UDP packets
- [ ] Data streaming to new nodes when cluster grows
- [ ] Virtual nodes instead of real nodes
- [ ] Memtable that flushes to disk when full

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ffreddygv%2Fcassandra-wannabe.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Ffreddygv%2Fcassandra-wannabe?ref=badge_large)