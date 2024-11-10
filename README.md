# Distributed Grep

YAAP (Yet Another Ambitious Project)

## About
[KISS](https://en.wikipedia.org/wiki/KISS_principle) implementation of [MapReduce](https://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf), distributed grep is one of the example
given in the paper that can ideally utilize MapReduce based distributed system 
constructs and pragmatics.

## Goal
Initially the system should be able to process large log based files for log analysis,
but overarching goal is to understand symantics of any text based file (or files).
But still intially it should be able to properly process logs. 

## Core components (might change later)
- Redis: For state management
- MinIO: For storage utilites
- NATS: For messaging

## Services 
- API: The API gateway
- Manager: The job manager and orchestrator
- Mapper: The Map() executer
- Reducer: The Reduce() executer
- Splitter: The Chonks generator 

## Deployment plan
- Would start with deploying on k3s
- If I get some compute, then can stress a bit more
