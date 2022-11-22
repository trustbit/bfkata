# Black Friday Kata


Status: stabilizing the domain internally at Trustbit. Not for public use, yet.

This is a scaffolding for DDD and Event Sourcing kata from Trustbit.

The idea is to implement backend for a minimal inventory management system. 
It has to be a separate application that is written in a language of your choice. You are free
to pick database engine and internal design that fits you.

API definitions and event contracts are pre-defined in gRPC/Protobuf. So you can easily
generate scaffolding and contracts in your language.

The only things that are fixed and shared by everybody:

- API methods (they are defined in gRPC spec)
- Event contracts
- 


We provide:

- API and Event definition in gRPC;
- event-driven spec tests that define behaviors;
- platform-independent test runner.


## API Definitions

API and Event definitions are available in [api/api.proto](api/api.proto).

There are two sets of APIs  implement:

- Inventory service - the real service.
- Spec service - service that allows to test individual specs.



You need to implement two different APIs



[api/api.proto](api/api.proto)

## Inventory API

Inventory API is as simple as we could get it. It has methods:

- AddLocations 
- AddProducts 
- ListLocations
- MoveLocation
- UpdateInventory
- GetLocInventory
- Reserve

There are following domain events:

- LocationAdded
- LocationMoved
- ProductAdded
- InventoryUpdated
- Reserved

You can find full definition in .


## Event-Driven Specs

Event-driven specs look like this:

```
move locations
------------------------------------------
GIVEN:
  LocationAdded id:1  name:"Warehouse"
  LocationAdded id:2  name:"Container"
WHEN:
  MoveLocationReq id:2  newParent:1
THEN:
  MoveLocationResp 
EVENTS:
  LocationMoved id:2  newParent:
```

They are bundled as text file in [specs/bundle.txt](specs/bundle.txt).

This repository includes a test runner that can run specs against gRPC implementation.

## Spec API

## Test Runner

If you have go 1.19 installed on your system (install 1.19, it has generics!), then:

```

> go install github.com/trustbit/bfkata@v1.0.1
go: downloading github.com/trustbit/bfkata v1.0.1

> bfkata  

Loaded 29 specs from <bundled>
Connecting to 127.0.0.1:50051...
connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:50051: connect: connection refused"

Test endpoint is not found. Did you start it?
```

