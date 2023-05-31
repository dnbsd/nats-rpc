# nats-rpc

Protocol agnostic Remote Procedure Call (RPC) framework over NATS.

WARNING: Work in progress! Expect API changes and things to break in unexpected ways.

## Supported protocols

| Protocol     | Project                                                 |
|--------------|---------------------------------------------------------|
| JSON-RPC 2.0 | [nats-json-rpc](https://github.com/dnbsd/nats-json-rpc) |

## Concepts

### Server

`Server` consumes request messages on NATS subjects and forwards them to registered `Services`, as well as produces
reply
messages.

`Server` can consume one or many subjects, as well as have one or many registered `Services` per subject.

All services must implement `Service` interface.

### Service

`Service` encodes/decodes request/response messages according to the rules of its protocol, and calls appropriate RPC
methods in bound `Receivers`.

`Service` can bind one or many `Receivers`, however each `Receiver` must have a unique name.

### Receiver

`Receiver` provides implementation of RPC methods. It receives the decoded data from a `Service` as a parameter to
method.

`Receiver` must be `Service` agnostic, meaning, that a `Receiver` must be compatible with any `Service`.
