# Tasks to-do

## Stage One
- [x] Read the [paper](https://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf) and watch [lecture](https://youtu.be/cQP8WApzIQQ?si=zYO9Z2i5jRHkeYxn)
- [x] Understand MapReduce on high level 
- [x] Figure out the core components 
    - [x] Messaging
    - [x] Job Management
    - [x] Storage Management
- [x] System Design (v1)

## Stage Two
- [x] Define the API spec
- [x] Implement Logger
- [ ] Core Services Design 
    - [ ] Storage Interface (MinIO)
        - [x] File Operations
        - [x] Chunk Management 
        - [x] Implementation
        - [ ] Testing
    - [ ] Job Management (Redis)
        - [ ] Job Defintions
        - [ ] Data Structures
        - [ ] Implementation
        - [ ] Testing
    - [ ] Messaging Design (NATS)
        - [ ] Architecture
        - [ ] Pub/Sub Design
        - [ ] Implementation
        - [ ] Testing

## Stage Three
- [ ] Manager 
    - [ ] Design
    - [ ] Implementation
    - [ ] Integration with NATS
    - [ ] Dockerfile
    - [ ] k8s spec
- [ ] Splitter
    - [ ] Design
    - [ ] Implementation
    - [ ] Integration with NATS and MinIO
    - [ ] Testing
    - [ ] Dockerfile
    - [ ] k8s spec
- [ ] Mapper
    - [ ] Design
    - [ ] Grepping algorithm design (use stuff already available)
    - [ ] Implementation
    - [ ] Integration with NATS and MinIO
    - [ ] Testing
    - [ ] Dockerfile
    - [ ] k8s spec
- [ ] Reducer 
    - [ ] Design
    - [ ] Integration and reduction mechanism design
    - [ ] Implementation
    - [ ] Integration with NATS and MinIO
    - [ ] Testing 
    - [ ] Dockerfile
    - [ ] k8s spec
- [ ] Final Integration tests
    
## Stage Four
- [ ] Bencmarking
- [ ] Complete Deployment
