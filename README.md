# Full stack demo-intro

This is a modern full-stack demo developed independently by Leon LiangNing individual .

In terms of functionality, half of it is a feed and half is e-commerce, referenced from the XiaoHongShu[小红书 – 你的生活指南 on the App Store (apple.com)](https://apps.apple.com/us/app/%E5%B0%8F%E7%BA%A2%E4%B9%A6-%E4%BD%A0%E7%9A%84%E7%94%9F%E6%B4%BB%E6%8C%87%E5%8D%97/id741292507).

## projects

- [Golang-framework kit-demo](https://github.com/lightning560/mio) is named mio
- [Golang Backend microservice-demo](https://github.com/lightning560/go-microservice)
- [SRE & Full chain stress test](https://github.com/lightning560/mio/blob/main/SRE&Test.md)
- [Web-demo](https://github.com/lightning560/Web-demo)
- [iOS-demo](https://github.com/lightning560/iOS-demo)
- [Flutter-demo](https://github.com/lightning560/Flutter-demo)

# Golang Framework Demo Syllabus

## High Availability Module

1. Isolation
2. Timeout and backoff retry,refer to doordash
3. Overload Protection, token bucket and adaptive rate limiting bbr algorithm,refer to alibaba/Sentinel
4. Quota ,distributed Rate Limit
5. Circuit Breaker use gRPC SRE algorithm
6. Degrade on BFF layer ,base on in-house cache middleware on gin library
7. load balance use p2c algorithm

## Base Module

- build script,refer to scripts from Istio and Envoy.  

- standard mircoservice layout

- Startup Parameters

- configuration ,remote dynamic configuration and hot loading

- Application cycle manage

- Service discovery and register

- Http server base on Gin

- RPC server base on gRPC
  
  - gRPC api design,refer to netflix

- Job server base on XXLjob

- MetaData propagate

- Custom Error Struct,refer to gRPC error design.combine with gRPC status, can pass error between services.

- Service online govern,reference to Etcd

- Log base on Zap

- Metric base on Prometheus,refer to go-metrics

- Tracing base on otel

- client
  
  1. redis
  2. etcd
  3. kafka
  4. mongo
  5. ...

# HA1,Isolation

## problems

### thumbnail service problem case

thumbnail service, being consumed all CPU by real-time scaling of large images, causing normal small image thumbnails to be discarded and a large number of 503 errors. The gif feature goes online, blocking other thumbnail services.
solution: Gif goes to other service

### log problem case

The amount of INFO logs is too large, causing a delay in the collection of ERROR logs.
solution: Log isolation, INFO and ERROR divided into different topics or even different ES clusters.

## Isolation Mode

### static and dynamic contents Isolation

Separate static resources from dynamic APIs. For example, CDN is for static resources.

### Fast-Slow Isolation

Distinguish by frequency of change.
Archive table: once created, rarely modified, archive_stat table: statistical table, frequently modified.

### Read-Write Isolation

Distinguish by reading and writing

- leader and follower
  1 leader and 3 followers; the leader table is written, and 3 follower tables are read.

- CQRS architectural pattern.

### Light-Heavy Isolation

Distinguish by business importance.
The logs are all placed in a Kafka topic for sequential IO, flowing to different logids. The logids will have different sink ends. There will be differences in speed between them, such as htfs jitter, es normal water level, and the overall data will be backed up globally. Isolate according to various dimensions: sink, part, business, logid, importance sabc Business logs belong to a certain logid, and the log level can be used as an isolation channel Live broadcast business s9 packet loss buried point, log surge.

### Hot Spot Isolation

For example, services for big promotions are written separately.

## Implementation Method

### Thread Isolation

The total control of goroutine also needs to be combined with breaker or based on semaphore, and adaptive current limiting is the most suitable.

### Process Isolation

Tools:

- KVM
- Docker Swarm
- Kubernetes

### Cluster Isolation

Multi-Cluster and multi-live construction
The naming method of the service: region.zone.cluster.appid

Main functions:

1. Only suitable for some businesses, not suitable for all businesses
2. shardID
3. Gateway，Switch traffic
4. Data synchronization, use CRDT to handle data conflicts

### Node High Availability

The main points are redundant services and the ability to discover service nodes.

Discover service nodes:

- Client realization
- Proxy

# HA2,Timeout

Timeouts determine the depletion of service threads.

Controlling chain failures is best achieved through timeout control. Fail fast timeout control is the first line of defense for microservice availability. A well-designed timeout strategy can prevent request accumulation, quickly clear high-latency requests, and free up Goroutines.

Generally, it's 100ms for internal and 300, 500ms for special circumstances; the external access should not exceed 1s.

## Timeout problems

- No configuration to limit retries, causing downstream crashes.
- SLB entry Nginx did not configure timeout leading to chain failure. Nearly 1 million long connections, nginx did not hang, all the requests were blocked, service has been timeout
- Network issues cause some servers to lose packets frequently.
- The timeout of the DB connection pool that the service depends on is not configured properly, causing the request to block and eventually leading to a collective OOM (Out of Memory) of the service.

## Best Practice

### Provide “Default” config

Have to set a timeout, don’t set it to never timeout, Golang Kit Mio provides a default timeout as a last resort, for example, 100ms.

if forget config `timeout` to cover default, service will use `default` configuration

### Timeout Propagation

To solve the problem of high latency services causing the client to waste resources by waiting,
Implement the timeout propagation in the following ways:

- Agree on a protocol in Proto
- Pass between processes
- Pass across processes.

#### Agree on a protocol in Proto

```ProtoBuf
package google.example.library.v1;
service LibraryService{
    // Lagency SLO: 95th in 100ms, 99th in 150ms
    rpc CreateBook(CreateBookRequest) returns(Book);
    rpc GetBook(GetBookRequest)returns (Book);
    rpc ListBooks(ListBooksRequest) returns (ListBooksResponse);
}
```

#### In-process timeout control: DDL+504 error code+context

When the upstream service has already timed out and returned a 504 error, but the downstream service is still executing, it can result in resource waste. Timeout transfer refers to passing the remaining quota of the current service to the downstream service, inheriting the timeout policy, and controlling the global timeout at the request level. It is necessary to use quota to check if there is sufficient remaining quota to process the request at each stage (network request) before it starts, and to inherit the timeout policy using the context.WithTimeout function from the Go standard library.

```go
func (c *asiiConn) Get(ctx context.Context, key string) (result *Item, err error) {
 c.conn.SetWriteDeadline(shrinkDeadline(ctx, c.writeTimeout))// 有分配的比如redis 50ms,mysql 100ms
 if _, err = fmt.Fprintf(c.rw, "gets %s\r\n", key); err != nil {
```

Pingpong also takes time, and usually 2 TTLs are deducted, which is about 10ms.

#### Across processes timeout control: gRPC metadata transmission

In the gRPC framework, gRPC Metadata Exchange is relied upon for transmitting the grpc-timeout field through HTTP2 headers automatically to the downstream service, constructing a context with a timeout.

### Insufficient remaining DDL detected

1. If we find there only a few milliseconds left, wait for the context.timeout error and throw the timeout error up.
2. If we find that there is no remaining timeout time at all, the client reports an error directly and does not send a request.

## Timeout data distribution

refer to Google SRE

- Bimodal distribution: 95% of requests are completed within 100ms, while 5% of requests may never complete (long timeouts). The long tail is the most concerning part; it is important to focus on the 5%, for example, through tracing.
- When monitoring, do not only rely on the mean value; it is useful to consider the distribution statistics, such as the 95th and 99th percentiles.
- Set reasonable timeouts, reject excessively long requests, or proactively fail when the server is unavailable.

![](/img/F8820F9F-ECB2-428B-9EB2-39CB3C373D98.png)

# HA2,retry and backoff

## retry Problems

- Large-scale failures caused by manual retries by users.
- Inconsistent timeout strategies between the client and server lead to resource waste.
- The magnifying effect of requests after the retry, multiple levels of retry propagation, resource number magnification, and traffic magnification cause avalanches.

## Refer to Doordash and ByteDance

<https://doordash.engineering/2018/12/21/enforce-timeout-a-doordash-reliability-methodology/>

## gRPC's backoff

[GitHub - grpc/grpc-go: The Go language implementation of gRPC. HTTP/2 based RPC](https://github.com/grpc/grpc-go)

## best practice

- Simple retry is not recommended
- Don't retry indefinitely
- Exponential backoff
- Jitter randomization
- Limit the number of retries and strategies based on retry distribution (retry ratio: 10%) ,for example gRPC should not exceed 3 times,set a retry ratio limit
- Retry should only be done at the failed level. When retry still fails, the global error code "overloaded, no need to retry" is agreed to avoid cascading retries.for example return 503 global error code, no more retries
- Write operations are not recommended for retry

## implement

```go
type Backoff struct {
 attempt uint64
 Factor float64
 Jitter bool
 Min time.Duration
 Max time.Duration
}

func (b *Backoff) Duration() time.Duration {
 d := b.ForAttempt(float64(atomic.AddUint64(&b.attempt, 1) - 1))
 return d
}

const maxInt64 = float64(math.MaxInt64 - 512)

func (b *Backoff) ForAttempt(attempt float64) time.Duration {
 min := b.Min
 if min <= 0 {
  min = 100 * time.Millisecond
 }
 max := b.Max
 if max <= 0 {
  max = 10 * time.Second
 }
 if min >= max {
  return max
 }
 factor := b.Factor
 if factor <= 0 {
  factor = 2
 }
 minf := float64(min)
 durf := minf * math.Pow(factor, attempt)
 if b.Jitter {
  durf = rand.Float64()*(durf-minf) + minf
 }
 if durf > maxInt64 {
  return max
 }
 dur := time.Duration(durf)
 //keep within bounds
 if dur < min {
  return min
 }
 if dur > max {
  return max
 }
 return dur
}

func (b *Backoff) Reset() {
 atomic.StoreUint64(&b.attempt, 0)
}

func (b *Backoff) Attempt() float64 {
 return float64(atomic.LoadUint64(&b.attempt))
}
func (b *Backoff) Copy() *Backoff {
 return &Backoff{
  Factor: b.Factor,
  Jitter: b.Jitter,
  Min:    b.Min,
  Max:    b.Max,
 }
}
```

## will Better

- The downstream service's release increases the time consumption, and the upstream service's timeout configuration is too short, leading to the failure of the upstream request. The downstream service modifies the timeout, but the upstream service doesn't know. Define the timeout in the .proto.
- The user's app client also needs to have retry limits.
- Service failures caused by abnormal clients (query of death). This can be limited by using a distributed quota.
- The _client_ side records the histogram of the number of retries, passes it to the server for distribution judgment, and the server can decide to reject. The server can know how many retries there are and can choose how to handle it.

## retry need Idempotency

The service is not idempotent, and retries will result in duplicate data.

- Global Unique ID: Generate a global unique ID according to the business. During the call, this ID will be passed. The provider checks if this global _ID_ exists in the corresponding storage system like Redis. If it exists, this means the operation has already been executed, and this service request will be rejected; otherwise, the service request will be responded to and the global _ID_ will be stored. After that, requests with the same business ID parameter will be rejected.
- Deduplication Table: This method is suitable for insert scenarios with unique identifiers in the business. For example, in the payment scenario, an order will only be paid once. We can establish a deduplication table and use the order ID as a unique index. The payment and the writing of the payment bill into the deduplication table are put into a transaction. In this way, when a repeated payment occurs, the database will throw a unique constraint exception, and the operation will be rolled back. This guarantees that an order will only be paid once.
- Multiversion concurrency control: This is suitable for controlling the idempotency of update requests. For example, updating the name of a product, we can add a version number to the update interface to control idempotency.

# HA3. Overload Protection

- When the server is nearly overloaded, it actively discards a certain amount of load. The goal is for the service to preserve itself.
- Under the premise of system stability, maintain the system's throughput.

## bbr algorithm:adaptive rate limiting

Referenced the BBR algorithm implementation from Alibaba/Sentinel.
An adaptive current limiting based on the BBR algorithm. The BBR algorithm is a TCP congestion control algorithm, which has certain similarities with the current limiting in microservices.

### Little's Law

Use Little's law to calculate the system's throughput.
The formula to calculate throughput is: `L = λ * W`
Where `λ` is considered as QPS, and `W` is the time each request takes.

### Calculate throughput

Since adaptive current limiting is still different from TCP congestion control, TCP clients can control the send rate, thereby detecting maxPass, but RPC cannot control the rate of traffic, so CPU/IOPS  is used to determine the maximum capacity of the system, formula is `cpu > 800 AND InFlight > (maxPass * minRtt * windows / 1000).`

- `CPU`: Use a separate thread for sampling, triggered once every 250ms. When calculating the average, a moving average is used to eliminate the influence of peak values.
- `Inflight`: The number of requests currently being processed in the service.
- `Pass`: In the last 5 seconds, pass is the number of successful requests in each 100ms sampling window
- `rt`: The average response time in a single sampling window.

```go
// cpu = cpuᵗ⁻¹ * decay + cpuᵗ * (1 - decay)
func cpuproc() {
 ticker := time.NewTicker(time.Millisecond * 500) // same to cpu sample rate
 defer func() {
  ticker.Stop()
  if err := recover(); err != nil {
   go cpuproc()
  }
 }()

 // EMA algorithm:
 for range ticker.C {
  stat := &cpu.Stat{}
  cpu.ReadStat(stat)
  prevCPU := atomic.LoadInt64(&gCPU)
  curCPU := int64(float64(prevCPU)*decay + float64(stat.Usage)*(1.0-decay))
  atomic.StoreInt64(&gCPU, curCPU)
 }
}


func (l *BBR) maxPASS() int64 {
 passCache := l.maxPASSCache.Load()
 if passCache != nil {
  ps := passCache.(*counterCache)
  if l.timespan(ps.time) < 1 {
   return ps.val
  }
 }
 rawMaxPass := int64(l.passStat.Reduce(func(iterator window.Iterator) float64 {
  var result = 1.0
  for i := 1; iterator.Next() && i < l.opts.Bucket; i++ {
   bucket := iterator.Bucket()
   count := 0.0
   for _, p := range bucket.Points {
    count += p
   }
   result = math.Max(result, count)
  }
  return result
 }))
 l.maxPASSCache.Store(&counterCache{
  val:  rawMaxPass,
  time: time.Now(),
 })
 return rawMaxPass
}

// timespan returns the passed bucket count
// since lastTime, if it is one bucket duration earlier than
// the last recorded time, it will return the BucketNum.
func (l *BBR) timespan(lastTime time.Time) int {
 v := int(time.Since(lastTime) / l.bucketDuration)
 if v > -1 {
  return v
 }
 return l.opts.Bucket
}

func (l *BBR) minRT() int64 {
 rtCache := l.minRtCache.Load()
 if rtCache != nil {
  rc := rtCache.(*counterCache)
  if l.timespan(rc.time) < 1 {
   return rc.val
  }
 }
 rawMinRT := int64(math.Ceil(l.rtStat.Reduce(func(iterator window.Iterator) float64 {
  var result = math.MaxFloat64
  for i := 1; iterator.Next() && i < l.opts.Bucket; i++ {
   bucket := iterator.Bucket()
   if len(bucket.Points) == 0 {
    continue
   }
   total := 0.0
   for _, p := range bucket.Points {
    total += p
   }
   avg := total / float64(bucket.Count)
   result = math.Min(result, avg)
  }
  return result
 })))
 if rawMinRT <= 0 {
  rawMinRT = 1
 }
 l.minRtCache.Store(&counterCache{
  val:  rawMinRT,
  time: time.Now(),
 })
 return rawMinRT
}

func (l *BBR) maxInFlight() int64 {
 return int64(math.Floor(float64(l.maxPASS()*l.minRT()*l.bucketPerSecond)/1000.0) + 0.5)
}

func (l *BBR) shouldDrop() bool {
 now := time.Duration(time.Now().UnixNano())
 fmt.Println("l.cpu()", l.cpu())
 if l.cpu() < l.opts.CPUThreshold {
  // current cpu payload below the threshold
  prevDropTime, _ := l.prevDropTime.Load().(time.Duration)
  if prevDropTime == 0 {
   // haven't start drop,
   // accept current request
   return false
  }
  if time.Duration(now-prevDropTime) <= time.Second {
   // just start drop one second ago,
   // check current inflight count
   inFlight := atomic.LoadInt64(&l.inFlight)
   return inFlight > 1 && inFlight > l.maxInFlight()
  }
  l.prevDropTime.Store(time.Duration(0))
  return false
 }
 // current cpu payload exceeds the threshold
 inFlight := atomic.LoadInt64(&l.inFlight)
 drop := inFlight > 1 && inFlight > l.maxInFlight()
 if drop {
  prevDrop, _ := l.prevDropTime.Load().(time.Duration)
  if prevDrop != 0 {
   // already started drop, return directly
   return drop
  }
  // store start drop time
  l.prevDropTime.Store(now)
 }
 return drop
}

// Stat tasks a snapshot of the bbr limiter.
func (l *BBR) Stat() Stat {
 return Stat{
  CPU:         l.cpu(),
  MinRt:       l.minRT(),
  MaxPass:     l.maxPASS(),
  MaxInFlight: l.maxInFlight(),
  InFlight:    atomic.LoadInt64(&l.inFlight),
 }
}

// Allow checks all inbound traffic.
// Once overload is detected, it raises limit.ErrLimitExceed error.
func (l *BBR) Allow() (DoneFunc, error) {
 if l.shouldDrop() {
  return nil, errors.ErrLimitExceed
 }
 atomic.AddInt64(&l.inFlight, 1)
 start := time.Now().UnixNano()
 return func(DoneInfo) {
  rt := (time.Now().UnixNano() - start) / int64(time.Millisecond)
  l.rtStat.Add(rt)
  atomic.AddInt64(&l.inFlight, -1)
  l.passStat.Add(1)
 }, nil
}
```

example

```go
func Test_example(t *testing.T) {
 allowed := func(ctx context.Context, req interface{}) (interface{}, error) {
  return "Hello valid", nil
 }
 limiter := bbr.NewLimiter()
 for i := 0; i < 1000; i++ {
  done, e := limiter.Allow()
  if e != nil {
   // rejected
   fmt.Println("error", xerrors.ErrLimitExceed)
   // return
  }
  // allowed
  _, err := allowed(context.Background(), nil)
  done(bbr.DoneInfo{Err: err})
 }
}
```

### Add a cooldown time

Without CD, there will be severe jitter

- After the current limiting effect takes effect, if no cooldown time is used, a short-term CPU drop may cause a large number of requests to be released, and in severe cases, it will max out the CPU.
- After the cooldown time, for example, a 5-second cooldown time, re-judge whether the threshold (CPU > 800) continues to enter overload protection.

## CoDel quesue

refer to TCP

Rejected requests are not discarded, but are put into a queue. However, the traditional first-in-first-out (FIFO) queue is not suitable for microservices because there can be timeouts and it's not possible to wait indefinitely. It's possible that SLP has already set an 800-millisecond timeout. If an old request is admitted at this time, its success rate will be lower because it may have been queuing for a long time.

Therefore, a queue based on discarding by processing time is needed. When the system is under high load, it enforces a Last-In-First-Out (LIFO) policy. This means it actively discards requests that have been queued for a long time and lets new requests pass through directly. This queue is used to compensate for the buffering problem in the previous algorithm and absorb a sudden increase in traffic.

# HA4. ratelimit

The above three High Availability (HA) methods are very important, because if the previous three methods can't withstand, the server will directly crash, and there is no chance to intervene.
However, rate limiting also costs every time a request is rejected, and it can't solve all problems. It's just to buy time for scaling up.

## Single-machine static flow control

### Insufficient

- Aimed at single nodes
- It is difficult to set good values
- Unable to control the total amount, unable to distribute flow control.
- Unable to set different priorities for requests

### Token bucket, static

`golang.org/x/time/rate`
Token bucket  is a rate limit algorithm we have been using from the beginning, and until now, many of the services are still using this algorithm.

At the beginning, the token bucket will contain some tokens, and new tokens will be received in the token bucket every few seconds. When the interceptor takes tokens from the token bucket, if it can be taken, it will continue to pass, and if it cannot be taken, it will be discarded.

#### Insufficient

- It only targets local server-side current limiting and cannot control global resources.
- The capacity of the token bucket and the rate of putting tokens cannot be well evaluated, because the system load is always changing. If the system scales in or out for some reasons, it needs to be manually modified, which results in high operation and maintenance costs.
- There is no prioritization, so important requests cannot be prioritized to pass through.

### Funnel algorithm, static

uber/ratelimit library traffic shaping, no matter how much come in, the come out is will always constant

#### Insufficient

Numbers are hard to set.

# HA5. Quota:distributed rate limit

## Architecture

using Redis for counting
![](/img/Pasted%20image%2020230720093920.png)

## Obtain quota tokens

- Use the default value for the first time. Once we have data from the past history window, we can request quota based on historical window data.
- Upgrade from getting a single quota to batch quotas. _quota:_ indicates the rate. After obtaining it, use the token bucket algorithm to limit.
- After each heartbeat, asynchronously batch get quota. This can greatly reduce the frequency of requesting Redis. After getting it, consume locally and intercept based on the token bucket.

### Active Pull

In addition to passively obtaining node data when requesting, there is also an active pull mechanism. The following two situations will apply:
1、When the Quota is exhausted;
2、When the Lease for applying for Quota is about to expire;

## Quota Allocation: Max-Min Fairness Algorithm

At the algorithm level, the Max-Min Fairness algorithm is used to solve the starvation caused by a large consumer and avoid the occurrence of unfair situations.

- Often facing the problem of allocating scarce resources to a group of users, they all enjoy equivalent rights to obtain resources, but some users actually need less resources than others.
- Intuitively, fair sharing assigns the minimum demand that each user wants to be met, and then evenly distributes the unused resources to users who need 'large resources'.

### Formal definition of the Max-Min Fairness algorithm

- Resources are allocated in increasing order of demand.
- There are no users who get more resources than they need.
- Unsatisfied users share resources equally.

## Quota service degradation

If the passive pull really fails, for example, if the QuotaServer fails, consider degrading to a local strategy or even allowing all directly pass;

## Better

- The underlying core infrastructure must do distributed current limiting.Cache penetration at the second level and large-scale fallback cause the core service to fail.
- Service failures caused by abnormal clients (query of death) ,this client can be limited by distributed quota.

# HA6. Circuit Breaker: Client Rate Limiting

## Purpose

- When a user exceeds their resource quota, the backend tasks will quickly reject requests and return an "insufficient quota" error. However, the rejection response still consumes certain resources. It is possible for the backend to be busy sending continuous rejection requests, leading to overload.

## Problem of Hystrix Circuit Breaker

The traditional circuit breaker problem is an all-or-nothing failure. The failure rate is relatively high

## Google SRE's Circuit Breaker

Using a sliding window with ring record data, the Google SRE circuit breaker algorithm handles overload. More information can be found at: Handling Overload: Google SRE Circuit Breaker Algorithm
<https://sre.google/sre-book/handling-overload/#eq2101>

Each client task retains the following information in its historical records from the last two minutes:

- Requests: The number of requests attempted by the client.
- Accepts: The number of requests accepted by the backend (total client requests minus the number of requests rejected by the backend).

Under normal circumstances, the values of accepts and requests are equal. As the backend starts rejecting traffic, the number of accepts becomes smaller than the number of requests. The client can continue sending requests to the backend until its requests are K times the accepts. Once this threshold is reached, the client starts to self-regulate and rejects new requests locally (i.e., on the client side) according to the probability calculated in the client's request rejection rate.

The probability of a client request being rejected:
`max(0, (requests - K * accepts) / (requests + 1))`.
![](/img/Pasted%20image%2020230612211759.png)

If failed requests increase while accepts remain the same, the float value will become smaller and smaller.

When the client itself begins to reject requests, requests will continue to exceed accepts. Although this seems counterintuitive, given that locally rejected requests are actually not propagated to the backend, this is the preferred behavior. As the rate at which the application tries to issue requests to the client increases (relative to the rate at which the backend accepts requests), we want to increase the likelihood of dropping new requests.

For services where the cost of handling a request is very close to the cost of rejecting that request, it may be unacceptable for rejected requests to consume about half of the backend resources.
In this case, the solution is simple: modify the acceptance multiplier K in the client request rejection probability. In this way:

- Reducing the multiplier will make adaptive throttling behavior more proactive
- Increasing the multiplier will make the adaptive throttling behavior less proactive

### Gutter  solution: Double Circuit Breaker

Based on the fuse of gutter kafka, it is used to take over the load during the operation of the automatic repair system, so that part of the system availability problem can be solved with only 10% of the resources.
a fuse uses b, b fuse uses a, b is a small kafka cluster with 10% resources 10% resources process overflow kafka's message

## Mobile app Traffic Control

Positive feedback: Users always actively retry and access an unreachable service. When an accident occurs, the traffic will double.

- The client needs to limit the request frequency and do some request concession with retry backoff.
- Can be mounted in the response of every API returned by the interface-level error_details

### Agree on some error code not to retry

For example, error codes like 5xx, 4xx, cannot retry within 5 seconds.

## implement:google SRE's circuit breaker

```Go
// Allow request if error returns nil.
func (b *sreBreaker) Allow() error {
    success, total := b.stat.Value()
    k:= b.k * float64 (success)

  if total < b.request II float64(total) < k {
        return nil
    }
    dr := math.Max(0, (float64(total) - k)/float64(total+1)) 
    rr := b.r.Float64()

    if dr <= rr {
      return nil
    }
     return code. ServiceUnavailable
}

// MarkSuccess mark requeest is success.
func (b *Breaker) MarkSuccess() {
 b.stat.Add(1)
}

// MarkFailed mark request is failed.
func (b *Breaker) MarkFailed() {
 b.stat.Add(0)
}

// trueOnProba
func (b *Breaker) trueOnProba(proba float64) (truth bool) {
 b.randLock.Lock()
 truth = b.r.Float64() < proba
 b.randLock.Unlock()
 return
}
```

### example

```go
func example(t *testing.T) {
 // allowedFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
  // return "Hello valid", nil
 // }

 faildFunc := func(ctx context.Context, req interface{}) (interface{}, error) {
  return nil, errors.InternalServer("", "")
 }

 breaker := sre.NewBreaker()
 if err := breaker.Allow(); err != nil {
  // rejected
  breaker.MarkFailed()
  return
 }

 _, err := faildFunc(context.Background(), nil)
 if err != nil && (errors.IsInternalServer(err) || errors.IsServiceUnavailable(err) || errors.IsGatewayTimeout(err)) {
  breaker.MarkFailed()
 } else {
  breaker.MarkSuccess()
 }
}
```

# HA7. Degrade

The essence of degradation is to provide lossy service.

Degrade responses to reduce workload, or discard unimportant requests. It's necessary to understand which traffic can be downgraded, and to differentiate requests. We usually offer reduced-quality responses to cut down on computational demands or time.

## Degradation Metrics

Determine which specific metric to use as the decisive factor for traffic evaluation and graceful degradation (e.g., CPU, latency, queue length, thread count, errors, etc.).

## Layer of Implementation for Degradation

Degrade strategies should trigger at the BFF (Backend for Frontend) layer to avoid data contamination and errors.

## Degradation Handling Strategies

- Modularize the UI and degrade non-core modules. Typically, the app should cache the last successfully opened page to avoid showing a blank screen when a request fails.
- Cache a copy of the last successful request page.
- Return default values, popular recommendations, etc. For example, a recommendation system can degrade by returning popular data.
- Traffic interception + regular data caching (expired copy policy).
- Page degradation, delayed services, write/read degradation, cache degradation.
- Throwing exceptions, returning agreed-upon protocols, using mock data, fallback processing.
- Delayed services, for instance, if there are issues with the account, only display the default avatar for the comment feature while continuing to show the comment content.

## Considerations

- Graceful degradation should not be frequently triggered.
- Regular drills should be conducted, where the code is not triggered or used regularly, to ensure the normal functioning of the mode.
- It should be simple enough to implement.

## Implementation

- There isn't much to say about the implementation of degradation since it varies.
- The BFF layer can handle data caching and regular data caching. Fallback data can be used here when content fails. Mock data can also be used.

### Degradation switch + dynamic config

```go
if hs == nil || err != nil || s.c.HonorDegradeSwitch {
 return s.degrade(c, h)
}
```

## Better

- The client fails to parse the protocol, causing the app to crash. Agree on the protocol for degraded data.
- Part of the client protocol is incompatible, causing the page to fail. Partial failure, do not cause global failure.
- Local cache data source caching, release invalidation + dependent interface failure, leading to a white screen. Put a copy in both local cache and remote redis, multiple copies.

# HA8.load balance

## Implementation methods

There are two options: client-side and server-side.
Because microservices decentralize, we chose the client-side method.

## WRR

Referring to the Nginx WRR load balancing algorithm:
Weighted Round Robin. The weights of NodeA, NodeB, and NodeC are set as 3:2:1 respectively. This means NodeA will be called 3 times, NodeB will be called 2 times, and NodeC will be called once. This method distributes the load efficiently.
The algorithm works as follows: for each peer selection, we increase the current_weight of each eligible peer by its weight, select the peer with the highest current_weight and reduce its current_weight by the total number of weight points distributed among peers.

### Drawbacks

However, this version has a few issues. Firstly, it cannot quickly remove problematic nodes. Secondly, it cannot balance backend loads, and thirdly, it cannot reduce overall latency.

## P2C Simplified version

1. Obtain the most up-to-date information as much as possible: Use the Exponentially Weighted Moving Average (EWMA) with time decay to update information such as latency and success rate in real-time.
2. Introduce the "best of two random choices" algorithm to add some randomness. The horizontal axis represents the time delay of information, and the vertical axis represents the average request response time. The best algorithm and load balancer 2.0 are similar when the horizontal coordinate is close to 0, but the difference becomes evident when the horizontal coordinate approaches 40 or 50.
3. Take inflight as a reference to balance traffic to bad nodes. The higher the inflight, the less likely it is to be scheduled.

Use the number of `inflight` as the indicator.

## P2C Complex version

Calculate weight score. We will update the delay for each incoming request and decay the previously obtained time delay by weighting. The newly obtained time gets a higher weight, implementing rolling updates this way.
`success*metaWeight / cpu*math.Sqrt(lag)*(inflight+ 1)`

`success`: success rate on the client side.
`metaWeight`: weight value set in service discovery.
`cpu`: the CPU usage rate on the server side for the recent period.
`lag`: request delay.
`inflight`: the number of requests currently being handled.

### lag

```go
//获取&设置上次测算的时间点
stamp := atomic. SwapInt64(&pc.stamp, now)
//获得时间间隔
td := now - stamp

//获取时间衰减系数
w := math. Exp(float64(-td) / float64(tau))

//获得延迟
lag := now - start 
oldLag := atomic.LoadUint64 (&pc. lag)

//计算出平均延迟
lag = int64(float64(oldLag)* + float64 (lag)*(1.0-w))
atomic.StoreUint64(&pc. lag, uint64(lag))
```

### EWMA

The metric calculation combines the moving average, applies time decay, and uses the formula `vt = v(t-1) * β + at * (1-β)`, where β is the reciprocal of several powers, i.e., `Math.Exp((-span) / 600ms)`.

### p2c

Best of Two Algorithm,Each time, a node A and B are randomly picked from all nodes. After the score comparison algorithm, the weight value in the code refers to the weight value set in Discovery.

P2C algorithm[https://ieeexplore.ieee.org/document/963420](https://ieeexplore.ieee.org/document/963420) , randomly selects two nodes to score, and chooses the better node: a very simple idea.

- Choose backend: CPU, client: health, inflight (QPS situation, this alone is not enough, combined with CPU), latency as indicators, and use a simple linear equation for scoring. This is consistent with the earlier self-protection data collection.
- For newly launched nodes (especially JVM), use a constant penalty value and minimize the amount of traffic with a probing method for preheating.
- For nodes with lower scores, to avoid entering a "permanent blacklist" and unable to recover, use statistical decay to gradually restore the node indicators to their initial state (_i.e., default values_).
- If the current request sent out exceeds the predicted latency, a penalty will be added.

```go
type leastLoaded struct {
 items []*leastLoadedNode
 mu sync.Mutex
 rand *rand.Rand
}

func (p *leastLoaded) Add(item interface{}) {
 p.items = append(p.items, &leastLoadedNode{item: item})
}

func (p *leastLoaded) Next() (interface{}, func(balancer.DoneInfo)) {
 var sc, backsc *leastLoadedNode

 switch len(p.items) {
 case 0:
  return nil, func(balancer.DoneInfo) {}
 case 1:
  sc = p.items[0]
 default:
  // rand needs lock
  p.mu.Lock()
  a := p.rand.Intn(len(p.items))
  b := p.rand.Intn(len(p.items) - 1)
  p.mu.Unlock()

  if b >= a {
   b = b + 1
  }
  sc, backsc = p.items[a], p.items[b]

  // choose the least loaded item based on inflight
  if sc.inflight > backsc.inflight {
   sc, _ = backsc, sc
  }
 }

 atomic.AddInt64(&sc.inflight, 1)

 return sc.item, func(balancer.DoneInfo) {
  atomic.AddInt64(&sc.inflight, -1)
 }
}
```

### Test RPC Load Balancing

This test is relatively important. If not careful during the launch, it might lead to an avalanche, so caution is needed. In addition to basic unit testing, the test code will simulate multi-client, multi-server scenarios, and randomly add network jitter, long-tail requests, server load mutations, request failures, and other situations that might occur in real scenarios, and print out the results at the end to determine whether the new function is effective.

Add some analysis to online, such as the current score, success rate, etc.

# Build

## Injecting Information during the Compilation Phase

Refer to scripts from Istio and Envoy.  
Because command line arguments are usually specified by developers, specifying a large number of startup parameters for each project can be tedious.  

Usually, common configurations such as environment, region, zone, configuration path, and startup IP are set as environment variables through infrastructure to **simplify** the startup parameters for developers.  

At the same time, environment variables can be used to enforce a **development standard** within the company. For example, using `dev.toml` for configurations in the `dev` environment and `prod.toml` for configurations in the `prod` environment.  

During the compilation phase, necessary information such as application name, application version, framework version, compilation machine host name, and compilation time can be injected using the `-ldflags` directive.

### script

`gobuild.sh`

```sh
VERBOSE=${VERBOSE:-"0"}
V=""
if [[ "${VERBOSE}" == "1" ]];then
    V="-x"
    set -x
fi

SCRIPTPATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

APP_NAME=${1:?"app name"}
APP_ID=${2:?"app id"}
OUT=${3:?"output path"}
shift

set -e

BUILD_GOOS=${GOOS:-linux}
BUILD_GOARCH=${GOARCH:-amd64}
GOBINARY=${GOBINARY:-go}
GOPKG="$GOPATH/pkg"
BUILDINFO=${BUILDINFO:-""}
STATIC=${STATIC:-1}
LDFLAGS=${LDFLAGS:--extldflags -static}
GOBUILDFLAGS=${GOBUILDFLAGS:-""}
# Split GOBUILDFLAGS by spaces into an array called GOBUILDFLAGS_ARRAY.
IFS=' ' read -r -a GOBUILDFLAGS_ARRAY <<< "$GOBUILDFLAGS"

GCFLAGS=${GCFLAGS:-}
export CGO_ENABLED=0

if [[ "${STATIC}" !=  "1" ]];then
    LDFLAGS=""
fi

# gather buildinfo if not already provided
# For a release build BUILDINFO should be produced
# at the beginning of the build and used throughout
if [[ -z ${BUILDINFO} ]];then
    BUILDINFO=$(mktemp)
    "${SCRIPTPATH}/report_build_info.sh"  ${APP_NAME} ${APP_ID}> "${BUILDINFO}"
fi

# BUILD LD_EXTRAFLAGS
LD_EXTRAFLAGS=""


while read -r line; do
    LD_EXTRAFLAGS="${LD_EXTRAFLAGS} -X ${line}"
done < "${BUILDINFO}"

# verify go version before build
# NB. this was copied verbatim from Kubernetes hack
minimum_go_version=go1.13 # supported patterns: go1.x, go1.x.x (x should be a number)
IFS=" " read -ra go_version <<< "$(${GOBINARY} version)"
if [[ "${minimum_go_version}" != $(echo -e "${minimum_go_version}\n${go_version[2]}" | sort -s -t. -k 1,1 -k 2,2n -k 3,3n | head -n1) && "${go_version[2]}" != "devel" ]]; then
    echo "Warning: Detected that you are using an older version of the Go compiler. Istio requires ${minimum_go_version} or greater."
fi

OPTIMIZATION_FLAGS="-trimpath"
if [ "${DEBUG}" == "1" ]; then
    OPTIMIZATION_FLAGS=""
fi

time GOOS=${BUILD_GOOS} GOARCH=${BUILD_GOARCH} ${GOBINARY} build \
        ${V} "${GOBUILDFLAGS_ARRAY[@]}" ${GCFLAGS:+-gcflags "${GCFLAGS}"} \
        -o "${OUT}" \
        ${OPTIMIZATION_FLAGS} \
        -pkgdir="${GOPKG}/${BUILD_GOOS}_${BUILD_GOARCH}" \
        -ldflags "${LDFLAGS} ${LD_EXTRAFLAGS}"
```

`report_build_info.sh`

```sh
APP_NAME=${1:?"app name"}
APP_ID=${2:?"app id"}

if BUILD_GIT_REVISION=$(git rev-parse HEAD 2> /dev/null); then
  if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
    BUILD_GIT_REVISION=${BUILD_GIT_REVISION}"-dirty"
  fi
else
  BUILD_GIT_REVISION=unknown
fi

# Check for local changes
if git diff-index --quiet HEAD --; then
  tree_status="Clean"
else
  tree_status="Modified"
fi

# XXX This needs to be updated to accomodate tags added after building, rather than prior to builds
RELEASE_TAG=$(git describe --match '[0-9]*\.[0-9]*\.[0-9]*' --exact-match 2> /dev/null || echo "")


GIT_DESCRIBE_TAG=$(git describe --tags)

# used by common/scripts/gobuild.sh
echo "mio.appName=${APP_NAME}"
echo "mio.appID=${APP_ID}"
echo "mio.buildAppVersion=${BUILD_GIT_REVISION}"
echo "mio.buildStatus=${tree_status}"
echo "mio.buildTag=${GIT_DESCRIBE_TAG}"
echo "mio.buildUser=$(whoami)"
echo "mio.buildHost=$(hostname -f)"
echo "mio.buildTime=$(date '+%Y-%m-%d--%T')"
```

### Version Script

After injection in this way and compilation is completed,

- In Prometheus, we can see the version of the app.
- Use `./hello --version` to see the version of the framework in the binary.

# Startup Parameters

## Priority of Obtain Parameters

1. Information injected during the source code compilation process.
2. Implementation of reading env in the util, can setting environment variables in Docker.
3. Reading flag parameters

## env Parameters

Setting ENV Parameters in Pods
Setting ENV Parameters in Dockerfile

`mio` has built-in environment variables, which allows for convenient pre-setting of some company-specific rules in the `Kubernetes` environment variables. This simplifies the startup parameters for the business side, and the startup command in the Dockerfile becomes a simple command: `CMD ["sh", "-c", "./${_APP_}"]`.

## Flag Parameters

It mainly involves a FlagSet data structure, which wraps the native FlagSet in Golang.
Based on this, the functions of register, parser, and lookup are implemented.
There is an Apply interface for various basic data structures to implement.

### some case

- config: MIO_CONFIG_PATH; config/local.toml; configuration path
- host: MIO_HOST; 0.0.0.0; startup IP
- watch: MIO_WATCH; true; default watch
- debug: MIO_DEBUG; false; enable debug mode or not
- mio_name: MIO_NAME; ; filepath.Base(os.Args[0]); application name
- mio_mode: MIO_MODE;  empty; environment
- mio_region: MIO_REGION; empty; region
- mio_zone: MIO_ZONE; empty; zone
- mio_log_path: MIO_LOG_PATH; ./logs; log configuration path
- mio_log_add_app: MIO_LOG_ADD_APP; false; whether to include application name in logs
- mio_trace_id_name: MIO_TRACE_ID_NAME; x-trace-id; trace name

### Open Source Alternatives

The total code is about 300 lines.

- ufive
- pflag

# Config

The functionality is easy to implement, but governance around config is critical.  

## Config Types

- Environment: Information such as region, zone, cluster, environment, color, discover, app ID, host, etc. These are injected into containers or physical machines by the online runtime platform for the kit library to use.  
- Static: Initialization configurations for resources like HTTP, gRPC, Redis, MySQL, etc. Changing these configurations online carries a high risk. For example, accidentally migrating SQL to different databases. On-the-fly changes are not encouraged as they can lead to unforeseen incidents. Changing static configurations should follow the same iterative release process as updating binary apps.  
- Dynamic: Simple types like int and bool. Applications may require online switches to control simple business strategies that are frequently adjusted and used. We can group these basic types (int, bool, etc.) configurations together, allowing for dynamic changes that affect business flows. We can also consider using tools like pkg.go.dev/expvar in combination.  
- Global: Copying configurations back and forth can easily lead to mistakes. Providing default configurations helps ensure consistency.

## Implementation of client part

### some libraries

The client implementation involves using the "mergo" library for merging data of type `map[string]interface{}` and the "mapstruct" library for converting a map to a struct.
The "cast" library extends more types and allows E to be extended to a map type for merging multiple maps.  

### Encryption

For encryption, we can consider using the "crypt" library or HashiCorp Vault.

### Hot Reloading of Configuration

Hot reloading of configuration can be implemented using the watch mechanism in etcd and channels in Golang.

### Open Source Alternatives for the Client

One open source alternative is `Viper`, which retrieves configurations from key-value stores.

## Functional Options Design Pattern for Loading Configuration

To load configurations, we can use the functional options design pattern.

```go
// ClientOption is HTTP client option.
type ClientOption func(*clientOptions)

// Client is an HTTP transport client.
type clientOptions struct {
 ctx          context.Context
 tlsConf      *tls.Config
 timeout      time.Duration
 endpoint     string
 userAgent    string
 encoder      EncodeRequestFunc
 decoder      DecodeResponseFunc
 errorDecoder DecodeErrorFunc
 transport    http.RoundTripper
 nodeFilters  []selector.NodeFilter
 discovery    registry.Discovery
 middleware   []middleware.Middleware
 block        bool
}

// WithTransport with client transport.
func WithTransport(trans http.RoundTripper) ClientOption {
 return func(o *clientOptions) {
  o.transport = trans
 }
}

// WithTimeout with client request timeout.
func WithTimeout(d time.Duration) ClientOption {
 return func(o *clientOptions) {
  o.timeout = d
 }
}

// WithUserAgent with client user agent.
func WithUserAgent(ua string) ClientOption {
 return func(o *clientOptions) {
  o.userAgent = ua
 }
}

// WithMiddleware with client middleware.
func WithMiddleware(m ...middleware.Middleware) ClientOption {
 return func(o *clientOptions) {
  o.middleware = m
 }
}

// WithEndpoint with client addr.
func WithEndpoint(endpoint string) ClientOption {
 return func(o *clientOptions) {
  o.endpoint = endpoint
 }
}

// Client is an HTTP client.
type Client struct {
 opts     clientOptions
 target   *Target
 r        *resolver
 cc       *http.Client
 insecure bool
 selector selector.Selector
}

// NewClient returns an HTTP client.
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
 options := clientOptions{
  ctx:          ctx,
  timeout:      2000 * time.Millisecond,
  encoder:      DefaultRequestEncoder,
  decoder:      DefaultResponseDecoder,
  errorDecoder: DefaultErrorDecoder,
  transport:    http.DefaultTransport,
 }
 for _, o := range opts {
  o(&options)
 }
 ...
}
```

Client.go defines a callback function ClientOption which accepts a pointer to a clientOptions structure that stores the actual configurations but is not exported. During the creation of NewClient, it uses variable parameters for transmission, and then modifies the related configurations by calling a for loop inside the initialization function.
The advantages of this approach are:

- Since the clientOptions structure is not exported, there is no possibility of being modified externally.
- It can distinguish between zero values and unset ones. First, we set default parameters when creating new clientOptions. If the corresponding Option is not passed from the outside, the default parameter will not be modified.
- Mandatory parameters are explicitly defined, and optional values are passed through Go's variable parameters, clearly distinguishing between mandatory and optional parameters.

## Implementation

In the `conf/init.go` file, we can register flags and implement the functions for reading configurations.

```go
type Configuration struct {

 mu sync.RWMutex
 override map[string]interface{}
 keyDelim string
 keyMap *sync.Map
 onChanges []func(*Configuration)
 onLoadeds []func(*Configuration)

 watchers map[string][]func(*Configuration)
 loaded bool

}
```

In the `init` function, we can register four flags and their corresponding actions. These actions will call the `NewDataSource` function, which should be implemented by different data sources and registered in the `init` functions of their respective packages. The `NewDataSource` function takes in the flag's name and URL to determine the appropriate data source to read from.

The results from the data source are then passed to the `LoadDataSource` function, which by default, calls the `toml` format deserialization and the implementation of the `defaultConf` interface. The `Load` method is then called to deserialize the content.

If there are any watchers, a goroutine can be started to listen for changes and trigger a refresh.

The configuration is stored in a `configuration` variable of type `map[string]interface{}`. The `apply` method is called, which executes `xmap.MergeStringMap` and stores the `keyMap`. If there are any changes, the `notifyChanges` method is called to notify the changes.

When calling the function, the NewDataSource function is first used to parse the address and format. Then errors are handled. The parsed result from NewDataSource is passed to the LoadFromDataSource function. ds.ReadConfig() is called and the result is passed to the c.Load method.

```go
//chan有通知就去调用apply。
go func() {
 for range ds.IsConfigChanged() {
  if content, err := ds.ReadConfig(); err == nil {
   _ = c.Load(content, unmarshaller)
   for _, change := range c.onChanges {
    change(c)
   }
  }
 }
}()
//ds.IsConfigChanged() 返回一个 <-chan struct{}
```

Load calls the c.Apply method
Apply is the final execution step
Apply, lock, then call map's MergeStringMap, if changes>0, exe notifyChanges(changes)

notifyChanges will call the watcher method to handle c

## Usage

For example, in the "governor" config:

```go
func StdConfig(name string) *Config {
 return RawConfig("server." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
 var config = DefaultConfig()
 if conf.Get(key) == nil {
  return config
 }

 if err := conf.UnmarshalKey(key, &config); err != nil {
  config.logger.Panic("govern server parse config panic",
  xlog.FieldErr(err), xlog.FieldKey(key),
  xlog.FieldValueAny(config),
  )
 }
 return config
}
```

1. Use the `UnmarshalKey` method in the `conf` package, which utilizes the "mapstruct" library, to parse the map into the config struct.
2. Configure business-specific settings.
3. In the `init` function, register the `OnLoaded` method of the conf struct, which executes the registered `OnLoadeds` methods during the load process.

## Improvements

### Decoupling using proto

Problem: not dependent on configuration go options, configuration load go config
Google SRE workbook2: just keep options API; config file(proto)+ options struct decoupling

### Align config with server

Config should be aligned with the binary version of the server.

### Permissions and change tracking

## Configuration Centre

### Options

- Use key value storage with consul
- Use Epollo
- Using Etcd, adopted scheme

### Configuration tool practice

Using **Monaco Editor** to fill in configurations.
proto,toml highlighting
proto,toml lint
proto,toml format

# application cycle

## Server and Client Startup Process

1. main.go creates a struct to wrap the application.
2. main.go new: grpc, http+serverclient. This step is executed in try, which is the 6th step xgo.SerialUntilError(fns...)()
3. For example, grpcsvr.StdConfig("grpc").MustBuild(). The startup of StdConfig("grpc") mainly loads the configuration, Build() is used to create services based on services.
4. StdConfig("grpc") will call RawConfig("mio.server." + name)
5. RawConfig will call DefaultConfig, and then load the corresponding &config based on the passed ("mio.server." + name), which is the default configuration of DefaultConfig
6. return config
7. Build() merely loads the original config into other configs such as \xx[ ]
8. For example, Interceptors are loaded into corresponding [], such as config.serverOptions
9. Many Options are executed in this step.
10. Finally, call newServer(Config)
11. newServer(Config),
12. The config is generated into grpc.NewServer and net.Listen()
13. Finally, return &Server{}
14. If it is grpc, the server generated in the previous step is registered into the corresponding service in pb
15. return eng.Serve(server). We can call Job, Schedule to load workers\jobs
16. This step only adds lock, and then loads server.Server into Application.servers
17. main.go new: eng, using the package struct of the application, calls eng.Startup( Inject the server, client created in the second step)
18. Startup is a method of the application. Call Application.initialize() and return the execution result of xgo.SerialUntilError(fns...)()
19. Application.initialize()
20. Use the initOnce of Application struct<->\*sync.Once  Do a func()
21. This func creates a new property of Application, which is equivalent to an instance of Application
22. Finally, call Application.parseFlags()
23. call flag.Pase() to parse all registered flags once
24. Application.printBanner()
25. Application.go's xgo.SerialUntilError(fns...)()
26. xgo executes all the input methods once, each method will call try(fn,nil)
27. util/xgo/ 's try(fn func(),cleaner func())
28. 2 defers, one executes cleaner, similar to wire
29. one recover. If an error is reported, runtime.Caller(2) is printed, and then errors are encased
30. finally returns the execution result of fn(), which is the second step
31. main.go, eng.Run()
32. Application.Run(... server.Server)
33. Loaded into Application.servers
34. app.waitSignals() //start signal listen task in goroutine
35. Wait for signal then GracefulStop or normal stop
36. app.cycle.Run(app.startServers\app.startWorks\startJobs)
37. Then startServers
    1. errgroup.Group.Go(func()),This step performs registry
38. Run
39. Use lock and wg and corresponding defer
40. Start a go func() to execute Run(fn()), if fn() returns err then c.quit <- err
41. if err := <-app.cycle.Wait() a <- chan error
42. return c.quit

## Delayed Startup

After a certain period of time, allow newly registered nodes to be exposed by the registry and reduce the weight of these nodes.

## Graceful Shutdown

Use a circuit breaker mechanism. When receiving the first shutdown signal, reject new requests, deregister from the registry, and then terminate after a 30-second delay. During this period, if a second shutdown signal is received, terminate immediately.

In the callback function, release any used resources, such as database and cache connections.

# HTTP service

Developed based on Gin.

## With MetaData

API gateways like Nginx or Envoy can write information into the HTTP headers. For example, extracting the remote IP from the header and saving it into the context's metadata.

In Gin, there is a `Keys` map[string]interface{} in the `context`, which can be used for metadata. The context of the HTTP service is available through `gin.Context.Request.Context()`.

Common information to extract includes user ID, traffic coloring information (e.g., x1-mio-color), and the remote IP of the client.

```go
 _httpHeaderUser         = "x1-mio-user"
 _httpHeaderColor        = "x1-mio-color"
 _httpHeaderTimeout      = "x1-mio-timeout"
 _httpHeaderRemoteIP     = "x-backend-mio-real-ip"
 _httpHeaderRemoteIPPort = "x-backend-mio-real-ipport"

func remoteIP(req *http.Request) (remote string) {
 if remote = req.Header.Get(_httpHeaderRemoteIP); remote != "" && remote != "null" {
  return
 }
 var xff = req.Header.Get("X-Forwarded-For")
 if idx := strings.IndexByte(xff, ','); idx > -1 {
  if remote = strings.TrimSpace(xff[:idx]); remote != "" {
   return
  }
 }
 if remote = req.Header.Get("X-Real-IP"); remote != "" {
  return
 }
 remote = req.RemoteAddr[:strings.Index(req.RemoteAddr, ":")]
 return
}
```

## Middleware

Gin has a rich ecosystem of middleware.such as

- recover
- cors
- HealthCheck
- cache
- ...

### Healthcheck Middleware

Provides liveness and readiness checks for the HTTP server, which can be utilized by Kubernetes.

### in-house Cache Middleware

Used in scenarios like request degradation and response cache

- Only cache 2xx HTTP responses,url as key,response asvalue

improvement of gin-contrib/cache.

- Cache http response in local memory or Redis.
- custom the cache strategy by per request.
- Use singleflight to avoid cache breakdown problem.

### Rate Limiting Middleware

For single-machine rate limiting, the Token Bucket algorithm is commonly used.

### Distributed rate limiting Middleware:quota

Implement distributed rate limiting to acquire quota.

## Response Wrapper

The response wrapper includes the following fields:

- `code`: Error code or status code.
- `message`: Corresponding message or reason.
- `data`: Response payload

### JSONErr(error code)

Given an error code, retrieve the corresponding message from a map based on the error code and reason.

### JSONSuccess(data interface{})

Set `code` to 0 and `message` to "success". If only `data` is provided, it will be included in the response.

### Improvement

utilize HTTP status codes based on the combination of `code` and `reason` as indicators.

# job service

## Application

Typically, there are three common ways to use the application:

1. Execute a one-time task, such as installing a program or generating mock data.
2. Delegate the application's lifecycle to platforms like Kubernetes (using Jobs) or XXLJOB. These platforms control the execution time and run binary projects.
3. Invoke a job's HTTP interface through a scheduled task.

## Parameters

- `X-Mio-Job-Name`: Used to identify the job's name. This parameter is required.
- `X-Mio-Job-RunID`: Used to record a unique execution ID for the job. This parameter is required.
- `Data`: User-defined data, optional.

#### Response

- Status Code: Returns 200 for successful execution, or 400 for parameter errors.
- Headers: The response will include the values of `X-Mio-Job-Name` and `X-Mio-Job-RunID` from the request headers.
- Parameter Error: Returns 400 with an additional `X-Mio-Job-Err` header to indicate the specific error.
  The main application of this is for domain servers.

# gRPC service

## gRPC Unit Testing

### bufconn

"google.golang.org/grpc/test/bufconn"
Use the `bufconn` package from gRPC to construct a listener, allowing tests to run without worrying about the IP and port of the gRPC service.

### Generating gRPC Test Code

Use `protoc-gen-go-test` to generate the test code for gRPC.

### Debugging with gRPC Reflection

The first step in debugging is to read the documentation. Forceful comments with the lint tool in the protobuf CI can help generate detailed documentation.

With gRPC Reflection, the server exposes its registered metadata, allowing third parties to access service and message definitions. By combining it with the Kubernetes API, users can select clusters, applications, and pods to perform gRPC interface testing directly online.

Test cases can also be archived for others to debug the interface.

### gRPC Mocking

Use the PowerMock library for mocking gRPC.

## With MetaData

### Obtaining Header Information in gRPC

- `app`
- `x-trace-id`
- `client-ip`
- `cpu`

### Setting App Name Middleware in gRPC Client

The gRPC client can set an app name middleware.

```go
func getPeerName(ctx context.Context) string {
 md, ok := metadata.FromIncomingContext(ctx)
 if !ok {
  return ""
 }
 val, ok2 := md["app"]
 if !ok2 {
  return ""
 }
 return strings.Join(val, ";")
}
```

### Obtaining Trace ID Header Information in Server

Prerequisites:

- The gRPC client uses MIO to set the trace ID middleware.
- The gRPC client has enabled tracing.
- The gRPC server has enabled tracing.

```
[trace.jaeger] # 开启链路
```

```go
// 如果开启了全局链路，获取链路id
// ExtractTraceID
// HTTP使用request.Context
func ExtractTraceID(ctx context.Context) string {
 if !IsGlobalTracerRegistered() {
  return ""
 }
 span := trace.SpanContextFromContext(ctx)
 if span.HasTraceID() {
  return span.TraceID().String()
 }
 return ""
}
```

### Obtaining Client IP Header Information in Server

The gRPC server can obtain the client IP information from the header.

```go
// getPeerIP 获取对端ip
func getPeerIP(ctx context.Context) string {
 clientIP := ig.GrpcHeaderValue(ctx, "client-ip")
 if clientIP != "" {
  return clientIP
 }

 // 从grpc里取对端ip
 pr, ok2 := peer.FromContext(ctx)
 if !ok2 {
  return ""
 }
 if pr.Addr == net.Addr(nil) {
  return ""
 }
 addSlice := strings.Split(pr.Addr.String(), ":")
 if len(addSlice) > 1 {
  return addSlice[0]
 }
 return ""
}
```

## print metadata

"github.com/davecgh/go-spew/spew"

## grpc health check

<https://github.com/grpc-ecosystem/grpc-health-probe.git>
<https://github.com/grpc/grpc/blob/master/doc/health-checking.md>

## gRPC protobuf API

Don't reuse request and response.
Don't over-abstract, there will be a lot of inappropriate fields.
map type is very suitable for passing two requests to get some nest data, to avoid the client traversed once to find the desired data

Empty reply google.protobuf.

### Problems with 0 values in proto:use WrapValue

Since there is no way to tell if a value is 0 or null

gRPC uses Protobuf v3 format by default, since required and optional keywords are removed, all fields are optional by default. If there is no assigned field, it will be the default value of the base type field, such as 0 or "".
In Protobuf v3, it is recommended to use: [https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/wrappers.proto](https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/wrappers.proto)  Warpper type field, i.e., wrapping a message, which becomes a pointer when used.

### use fieldMask for partially updated

`FieldMask` is used for response, request parameter limit returns, and updates for specified parameters.
For example, the account domain service does not provide a single interface for updating avatars and names, but only provide a large and comprehensive interface.

`FieldMask`  library:
[https://github.com/mennanov/fmutils](https://github.com/mennanov/fmutils)
[https://github.com/mennanov/fieldmask-utils](https://github.com/mennanov/fieldmask-utils)

#### refer to netflix api

[Netflix API Design Practice: Using FieldMask](https://mp.weixin.qq.com/s/L7He7M4JWi84z1emuokjbQ)

Specify paths in request and response returns the required fields based on the paths mask parameter.
[Netflix API Design Practice (II): Using FieldMask for Data Changes](https://mp.weixin.qq.com/s/uRuejsJN37hdnCN4LLeBKQ)

# Metadata

Metadata is a data structure that stores key-value pairs and is used to pass and store additional information in different applications. In a specific implementation, a `map[string]string` can be used to represent Metadata, where each key-value pair has a specific prefix.  

## Approach

Metadata can be loaded and processed in different ways in different parts of an application.
 For example, when making an HTTP request, Metadata can be loaded into the request's header.
 When working with Golang internally, Metadata can be loaded into the context.
 When communicating using gRPC, Metadata can be loaded into gRPC's metadata.
 Similarly, Kafka also has its own Metadata implementation. Redis and databases also need to support appropriate operations for Metadata.  

## prefix

Different prefixes can be used to identify different purposes for Metadata keys.
 For example, keys starting with "x-md-global-" can be used for global Metadata, while keys starting with "x-md-local-" can be used for local Metadata.  

On the server side, Metadata can be added to the context using keys prefixed with "x-md-". On the client side, Metadata can be added to the request's header using keys in the form "x-md-global-xxx".

## implement

```go
// Metadata is our way of representing request headers internally.
// They're used at the RPC level and translate back and forth
// from Transport headers.
type Metadata map[string]string

// FromClientContext returns the client metadata in ctx if it exists.
func FromClientContext(ctx context.Context) (Metadata, bool) {
 md, ok := ctx.Value(clientMetadataKey{}).(Metadata)
 return md, ok
}

// Clone returns a deep copy of Metadata
func (m Metadata) Clone() Metadata {
 md := Metadata{}
 for k, v := range m {
  md[k] = v
 }
 return md
}

// AppendToClientContext returns a new context with the provided kv merged
func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
 if len(kv)%2 == 1 {
  panic(fmt.Sprintf("metadata: AppendToOutgoingContext got an odd number of input pairs for metadata: %d", len(kv)))
 }
 md, _ := FromClientContext(ctx)
 md = md.Clone()
 for i := 0; i < len(kv); i += 2 {
  md.Set(kv[i], kv[i+1])
 }
 return NewClientContext(ctx, md)
}

// MergeToClientContext merge new metadata into ctx.
func MergeToClientContext(ctx context.Context, cmd Metadata) context.Context {
 md, _ := FromClientContext(ctx)
 md = md.Clone()
 for k, v := range cmd {
  md[k] = v
 }
 return NewClientContext(ctx, md)
}
```

# protobuf

## Application

- **Config Constraints**: Config constraints refer to the rules or limitations set for the configuration of an application. These constraints help ensure that the application is properly configured and that certain requirements are met.  

- **Error Definitions**: Error definitions are used to define different types of errors that can occur during the execution of an application. By defining specific error types, it becomes easier to handle and communicate errors in a structured and consistent manner.  

- **Struct for Data in RESTful Responses**: When designing RESTful APIs, it's common to define specific data structures that represent the response body. These structures define the format and data fields that will be returned to the client when an API request is made.  

- **Generating Swagger for RESTful Interfaces**: Swagger is a widely-used tool for documenting and exploring APIs. In the context of generating RESTful interfaces, Swagger can be used to automatically generate API documentation based on the specified endpoints, request parameters, and response structures.  

## Validate

"validate-proto-pgv" and "proto-gen-validate" are validation libraries developed based on the Envoy proxy server. It enables validation of protobuf messages using the Protobuf

example

```protobuf
// 参数必须大于 0
int64 id = 1 [(validate.rules).int64 = {gt: 0}];
// 参数必须在 0 到 120 之间
int32 age = 2 [(validate.rules).int64 = {gt:0, lte: 120}];
// 参数是 1 或 2 或 3
uint32 code = 3 [(validate.rules).uint32 = {in: [1,2,3]}];
// 参数不能是 0 或 99.99
float score = 1 [(validate.rules).float = {not_in: [0, 99.99]}];
```

gen code `pb.go`

```bash
protoc --proto_path=. \
           --proto_path=./third_party \
           --go_out=paths=source_relative:. \
           --validate_out=paths=source_relative,lang=go:. \
           xxxx.proto
```

usage

```go
req := &feedv1.GetPostByIdReq{
 Id: id,
}
err = req.Validate()
if err != nil {
 xlog.Errorf("GetPostById validate err", err)
 resp.JSONErr(c, xgin.StatusHttpRequestValidateError)
 return
}
```

### validate middleware

also use a middleware to validate.
To ensure that structs implementing the `Validate()` interface are automatically validated, we can implement middleware that intercepts the requests and performs validation before passing them to the actual handlers.

# health check

## Application

kubernetes liveness & readiness

1. to prevent discover available, but the client can not connect to the service scene, such as a colleague hand touched the optical fibre only this line packet loss
2. Rolling deployment scenarios, graceful shutdown, health check can be combined with discover double insurance.Then in combination with rolling updates, the service can be started up very elegantly.

## kubernetes liveness & readiness

[liveness&readiness | Kubernetes](https://kubernetes.io/zh-cn/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#%E5%AE%9A%E4%B9%89-grpc-%E5%AD%98%E6%B4%BB%E6%8E%A2%E9%92%88)

- livenessProbe: if the check fails, it will kill the container, according to Pod's restartPolicy
  liveness is usually a service panic, the process is gone, detection ip port does not exist, this time Kubernetes will kill our container.
  liveness Probe set a tcp detection ip port can be.

- readinessProbe: if the check fails, Kubernetes will remove the Pod from the service endpoints.
  readinessProbe is that our service may not be responding due to load issues,3 but the ip port can still be connected, this time Kubernetes will remove we from the service endpoints.

For readness, we need to set up different probing policies based on HTTP, gRPC, and so on.
When we make sure that the service interface is readness, the traffic will be imported.
liveness, readness must be set at the same time, and the policy must be different, otherwise it will cause some problems

# Service Discovery

In the early days, fixed IP-based DNS was commonly used for service discovery. However, with the emergence of RPC (Remote Procedure Call), service discovery mechanisms were also introduced.

## Implementation options

1. Client-side Discovery: In this approach, the client is responsible for discovering and locating the available services. The client typically obtains the service information from a registry or a service catalog and then communicates directly with the discovered service.

2. Server-side Discovery: In this approach, the server registers itself with a registry or a service catalog. Clients then query the registry to discover the available services. However, server-side discovery is not in line with the decentralized nature of microservices.

# Registry System

A registry system, such as etcd, serves as a central location where services can register their availability, enabling other services or clients to discover and communicate with them.

## Alternative Solutions

Consul is an alternative registry system that can be used as a replacement for etcd. Consul provides similar functionalities and can reduce some of the development effort required for service discovery.

# Deployment

## Dockerfile

In a Dockerfile, the startup command can be simplified to a straightforward command line. This allows for easy and streamlined deployment of the application within a Docker container.

```less
FROM golang:1.16 AS builder

ARG APP_RELATIVE_PATH

COPY . /src
WORKDIR /src/app/${APP_RELATIVE_PATH}

RUN GOPROXY=https://goproxy.cn make build

FROM debian:stable-slim

ARG APP_RELATIVE_PATH

RUN apt-get update && apt-get install -y --no-install-recommends \
  ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/app/${APP_RELATIVE_PATH}/bin /app

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
```

## compose

```less
version: '3'
services:
 mall-bff:
  image: redbook/mall-bff:0.1.0
  ports:
   - 80:80
 mall-domain:
  image:redbook/domain-mall:0.1.0
 mall-db:
        image: mongo:**5.0.6
        container_name: mongo5_redbook
        restart: always
        ports:
            - 27017:27017
        command: ""
        environment:
            MONGO_INITDB_ROOT_USERNAME: root
            MONGO_INITDB_ROOT_PASSWORD: password
        volumes:
            - "/mongo5/data:/data/db"
            - "/mongo5/config:/data/configdb"

    redis:
        container_name: redis6
        image: redis:6.2.6
        ports:
            - 6379:6379
        volumes:
            - "/redis/data:/data"
 kafka:
  image: docker.io/bitnami/kafka:2
  ports:
   - "9092:9092"
  environment:
   - ALLOW_PLAINTEXT_LISTENER=yes
  jaeger-in-one:
    container_name: jaeger
    image: jaegertracing/all-in-one
    ports:
      - 5775:5775/udp
      ...
      - 9411:9411
    environment:
      COLLECTOR_ZIPKIN_HOST_PORT: 9411
```

## kubenetes

Convert docker-compose.yml to kubernetes using kompose

```sh
kompose convert
```

# govern

Reference to the etcd implementation

## Problems with golang

Unlike `Java` and `PHP`, `Go` do not has a virtual machine to help the programmer observe the internals of the program running, but this observation is very important for the programmer to troubleshoot and solve performance problems.
MIO framework focuses on observability and introduces interceptors in each component to extract useful information, and implements a governance service to spit out the program running data, making it easy for users to make a governance platform to troubleshoot various problems.

## implement

use `net/http` to provide http api.

Mainly output some data of the prof package, exposing some information of service
`net/http/prof`
`runtime/debug`
`github.com/felixge/fgprof`,Call fgprof method and get the io elapsed time of the function

## Govern information

`/metrics` Monitor data
`/debug/pprof/*` pprorf information
`/configs` configuration information
`/config/json` config json data
`/config/raw` config raw data
`/module/info` Application dependency module info
`/build/info` Application compilation information
`/env/info` Application environment information
`/code/info` Status code information, to be completed
`/component/info` Component information, pending.
`/routes` Governance routes
`/debug/pprof/*` pprof info
`/buildInfo` Project build info
`/moduleInfo` Version information for project dependencies
`/status/code/list` list of status codes

## pprof

profiling refers to a picture of an application, which is a picture of how the application uses CPU and memory.
Golang is a language that puts a lot of emphasis on performance, so the language comes with a profiling library that allows we to obtain cpu, heap, block, traces, and other execution information while the application is running.

Performance optimisation in Golang is mainly in the following areas:

- CPU profile: report the CPU usage of the application, and collect the data of the application on CPU and registers according to a certain frequency.
- Memory Profile(Heap Profile): reports the memory usage of the application.
- Block Profiling: report goroutines not in running state, can be used to analyse and find deadlocks and other performance bottlenecks.
- Goroutine Profiling: reports on the usage of goroutines, what goroutines are available, and how they are invoked.

## implement

### ppof

```go
package governor
var (
 // DefaultServeMux ...
 DefaultServeMux = http.NewServeMux()
 routes          = []string{}
)

func init() {
 // 获取全部治理路由
 HandleFunc("/routes", func(resp http.ResponseWriter, req *http.Request) {
  _ = json.NewEncoder(resp).Encode(routes)
 })

 HandleFunc("/debug/pprof/", pprof.Index)
 HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
 HandleFunc("/debug/pprof/profile", pprof.Profile)
 HandleFunc("/debug/pprof/symbol", pprof.Symbol)
 HandleFunc("/debug/pprof/trace", pprof.Trace)

 if info, ok := debug.ReadBuildInfo(); ok {
  HandleFunc("/modInfo", func(w http.ResponseWriter, r *http.Request) {
   encoder := json.NewEncoder(w)
   if r.URL.Query().Get("pretty") == "true" {
    encoder.SetIndent("", "    ")
   }
   _ = encoder.Encode(info)
  })
 }
}

// HandleFunc ...
func HandleFunc(pattern string, handler http.HandlerFunc) {
 // todo: 增加安全管控
 DefaultServeMux.HandleFunc(pattern, handler)
 routes = append(routes, pattern)
}
```

### config,env,info

```go
func init() {
 conf.OnLoaded(func(c *conf.Configuration) {
  log.Print("hook config, init runtime(governor)")

 })

 registerHandlers()
}

func registerHandlers() {
 HandleFunc("/configs", func(w http.ResponseWriter, r *http.Request) {
  encoder := json.NewEncoder(w)
  if r.URL.Query().Get("pretty") == "true" {
   encoder.SetIndent("", "    ")
  }
  _ = encoder.Encode(conf.Traverse("."))
 })

 HandleFunc("/debug/config", func(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(200)
  _, _ = w.Write(xstring.PrettyJSONBytes(conf.Traverse(".")))
 })

 HandleFunc("/debug/env", func(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(200)
  _ = jsoniter.NewEncoder(w).Encode(os.Environ())
 })

 HandleFunc("/build/info", func(w http.ResponseWriter, r *http.Request) {
  serverStats := map[string]string{
   "name":       env.Name(),
   "appID":      env.AppID(),
   "appMode":    env.AppMode(),
   "appVersion": env.AppVersion(),
   "mioVersion": env.MioVersion(),
   "buildUser":  env.BuildUser(),
   "buildHost":  env.BuildHost(),
   "buildTime":  env.BuildTime(),
   "startTime":  env.StartTime(),
   "hostName":   env.HostName(),
   "goVersion":  env.GoVersion(),
  }
  _ = jsoniter.NewEncoder(w).Encode(serverStats)
 })
}
```

# Errors

## Golang Error Handling Approaches

There are three common approaches to error handling:

1. Native Errors: Using the built-in `error` type, `string` type, or pre-defined error types.
2. Error Wrapping: Introduced in Go 1.13, this approach wraps errors to provide additional context and information.
3. Custom Error Structs: Defining custom error structs that implement the `Error()` method, similar to the approach used in gRPC's `status` package.

### Custom Error Struct

Since an error in Go is an interface, any struct that implements the `Error()` method is considered an error type. For custom errors, the design of `grpc status` was referenced.

## Errors API Design

The design of the errors API takes inspiration from the error handling approach used in gRPC. It includes the following elements:

- Code: Represents the broad category of the error.
- Reason: Represents the specific reason for the error.
- Message: Describes the error information and provides guidance on how to handle it.
- Metadata: Contains additional information related to the error.

By defining error codes, the problem of defining similar "not_found" errors multiple times can be avoided.

Errors that require special handling should explicitly state the handling method. For errors that don't require special handling, only the error itself needs to be returned.

- The `Code` is used for external display and preliminary judgment. Instead of defining numerous unique `XXX_NOT_FOUND` errors, a standard `Code.NOT_FOUND` error code is used, indicating that a specific resource cannot be found. This reduces the complexity of documentation, provides better mapping in client libraries, and reduces client-side logic complexity. Additionally, the presence of standardized codes is more friendly towards external observability systems, such as using analysis of HTTP status codes in Nginx Access Logs for server monitoring and alerts.

- The `Reason` provides a more detailed explanation of the error and can be used to make finer error judgments. Each microservice defines its own unique `Reason` and should use domain-specific prefixes, e.g., `User_XXX`.

- The `Message` helps users understand and resolve API errors quickly and easily.

- The `Metadata` can be used to store additional standardized error details, such as "retryInfo" and error stack traces. This enforced standardization helps prevent sensitive information leakage by ensuring that don't directly propagate the sensitive Go error.

## Integration with gRPC

Support for status conversion between the errors API and gRPC's `status` package is provided.

### Conversion between gRPC's `Status` and Go's `Error`

As mentioned earlier, both the server and client return `error` types. However, returning a `Status` directly is not feasible. Fortunately, gRPC provides methods to convert between `Status` and Go's `Error`.

![](/img/Pasted%20image%2020230723122816.png)
So on the server side, we can use `.Err()` to convert `Status` into an `error` and return it, or we can directly create an `error` from a `Status` with `status.Errorf(codes.InvalidArgument, "invalid args")` and return it.  

### Conversion between grpc status and miopkg's error

1. Convert error to mio error  
2. Convert mio error to status  
3. Convert server status to grpc error  
4. Convert client error to status  
5. Convert status to mio error  

### Handling Errors with gRPC

- The client sends a request to the server using the `invoker` method.
- The server, through the `processUnaryRPC` method, obtains the `error` information from user code.
- The server converts the `error` to a `status.Status` using the `status.FromError` method.
- The server writes the data from the `status.Status` to the `grpc-status`, `grpc-message`, and `grpc-status-details-bin` headers using the `WriteStatus` method.
- The client receives these headers over the network and parses the `grpc-status` information using `strconv.ParseInt`, the `grpc-message` information using `decodeGrpcMessage`, and the `grpc-status-details-bin` information using `decodeGRPCStatusDetails`.
- The client retrieves the user code error using `x.Status().Err()`.

The following code is for converting the original error to mio error type:

```go
// FromError try to convert an error to *Error.
// It supports wrapped errors.
func FromError(err error) *Error {
 if err == nil {
  return nil
 }
 if se := new(Error); errors.As(err, &se) {
  return se
 }
 gs, ok := status.FromError(err)
 if ok {
  ret := New(
   httpstatus.FromGRPCCode(gs.Code()),
   UnknownReason,
   gs.Message(),
  )
  for _, detail := range gs.Details() {
   switch d := detail.(type) {
   case *errdetails.ErrorInfo:
    ret.Reason = d.Reason
    return ret.WithMetadata(d.Metadata)
   }
  }
  return ret
 }
 return New(UnknownCode, UnknownReason, err.Error())
}
```

### Correspond with HTTP's response

When handling errors, some internal error information cannot be exposed to the user, so it needs to be transformed before throwing the error to the user.
Map error codes to gin's error codes, and transform internal errors into error messages that can be shown to the user.
Switch between error codes and reasons for mapping to HTTP response error codes.
Error code <--> HTTP response error code mapping, return corresponding error messages to the user based on the error code.
Mio error ---> HTTP status code

## Usage

```go
func Test_errors(t *testing.T) {
 base := errors.New("base")
 fmt.Println(base)
 >>>base
 xbase := xerrors.FromError(base)
 fmt.Println(xbase)
 >>>error: code = 500 reason = message = base metadata = map[] 
 wbase := xerrors.Wrapf(base, "wrap %d", 123)
 fmt.Println(wbase)
 >>>wrap 123: base 
 wxbase := xerrors.FromError(wbase)
 fmt.Println(wxbase)
 >>>error: code = 500 reason = message = wrap 123: base metadata = map[] 
 wwxbase := xerrors.Wrapf(wxbase, "wrap %d", 456)
 fmt.Println(wwxbase)
 >>>wrap 456: error: code = 500 reason = message = wrap 123: base metadata = map[] 
 fmt.Println(xerrors.Is(wbase, base))
 >>>true
 fmt.Println(xerrors.Is(wwxbase, xbase))
 >>>true
}
```

## method of providing

Provide the following methods through the wrap method.

- Is()
- As()
- wrap()
- unwrap()
- ...

```go
package errors

import (
 stderrors "errors"

 pkgerrors "github.com/pkg/errors"
)

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
func Is(err, target error) bool { return stderrors.Is(err, target) }

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// As will panic if target is not a non-nil pointer to either a type that implements
// error, or to any interface type. As returns false if err is nil.
func As(err error, target interface{}) bool { return stderrors.As(err, target) }

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
 return stderrors.Unwrap(err)
}

// Wrap returns an error annotating err with a stack trace at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, msg string) error {
 return pkgerrors.Wrap(err, msg)
}

// Wrap returns an error annotating err with a stack trace at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
 return pkgerrors.Wrapf(err, format, args...)
}
```

## Error with Stack Trace

The `pkg/errors` package provides error with stack trace information.
For programs, errors should include error details such as error type and error code for easy propagation between modules.
For humans, errors should include code information such as relevant call parameters and runtime information for easier troubleshooting.
In summary, errors should hide implementation details from programs while displaying them for diagnosis. The `runtime.Caller()` function provides stack trace information when errors occur.

## Error Propagation

Finding a solution for error propagation can be challenging. Global errors can be loose and easily broken, relying on a gentleman's agreement. A better approach is to translate errors at each service and ensure that each service and error enumeration is unique, as defined in the proto file.  

It is not advisable to blindly propagate these service errors to our clients. When translating errors, we recommend the following:  

- Hide implementation details and confidential information.  
- Adjust the responsibility of the error to the appropriate party. For example, if a server receives an INVALID_ARGUMENT error from another service, it should propagate INTERNAL error to its own caller.  

### Error Aggregation using Libraries

When writing code with retry strategies, aggregating multiple errors into one can be very useful. Utilize goroutines for parallel and serial execution, and consider using libraries like `go.uber.org/multierr`.  

## Error Handling

### Pass the error up to the caller and log it at the top level or in the middle

Avoid logging errors using `log.error`. Instead, return the error directly or wrap it using `errors.wrap`.  
Convert the underlying error to mio error and wrap it before throwing it. The top-level application should log the error.

### Handle the error immediately

If the error is properly handled, there is no need to return it.

```go
if err := planA(); err != nil {
 log. Infof("could't open the foo file, continuing )
 planB ()
}
```

There has been a degradation of service here, which essentially belongs to a diminished level of service. It is more preferable to use "Warning" in this context.

# Log

| Method | Logging Type                                         | Scale  | Advantages                       | Adjustments          | Format  |
| ------ | ---------------------------------------------------- | ------ | -------------------------------- | -------------------- | ------- |
| Log    | Requests, errors, slow logs                          | Medium | Aggregation, error convergence   | Adjust log level     | Dynamic |
| Trace  | Request logs (can be granular to the function level) | Large  | TraceId, invocation relationship | Adjust sampling rate | Fixed   |

## Basic Features

### Log Level

There are only two things we should log:

- Things that developers care about when developing or debugging software.
- Things that users care about when using software.

The log levels are referenced from zap:

- debug
- info
- warn
- error
- panic

#### Warning Level

No one reads warnings because, by definition, nothing is wrong. Maybe there will be a problem in the future, but that sounds like someone else's problem. We try to eliminate warning levels as much as possible. They are either informational messages or errors. Taking inspiration from the design philosophy of Go language, all warnings are treated as errors. Other languages' warnings can be ignored unless they are forced to be treated as errors by the IDE or in the CI/CD process, compelling programmers to eliminate them as much as possible. Similarly, if we want to eventually eliminate a warning, we can log it as an error to make the code author pay attention to it.

Warnings are particularly suitable for library deprecation and degradation scenarios.

#### Fatal Level

After logging the message, call _os.Exit(1)_. This means:

- defer statements in other goroutines will not be executed.
- All kinds of buffers, including log buffers, will not be flushed.
- Temporary files or directories will not be removed.

Do not use "_fatal_" to log, but return the error to the caller. If the error persists until _main.main_, that's the right place to handle any cleanup operations before exiting.

#### Debug Level

Clearly, they are for debugging and informational purposes respectively.

_log.Info_ is simply writing that line to the log output. There should be no option to turn it off because users should only be informed of things relevant to them. If an unhandled error occurs, it will be thrown to _main.main_. _main.main_ is where the program terminates. Insert a "fatal" prefix before the final log message, or write it directly to _os.Stderr_.

_log.Debug_, on the other hand, is something entirely different. It is controlled by developers or support engineers. During development, debug statements should be abundant instead of relying on _trace_ or _debug2_ (we know who we are) levels. The logging package should support fine-grained control to enable or disable debugging and only enable or disable debug statements within the package or finer scope. Look into Google's glog for more information.

### Log Format

1. JSON
2. Text
3. Terminal Output (debug mode),console colorful print

Terminal color display can be achieved using the faith/color library.

```go
// DebugEncodeLevel ...
// / 根据lvl，打印不同的颜色
func DebugEncodeLevel(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
 var colorize = color.RedString
 switch lv {
 case zapcore.DebugLevel:
  colorize = color.BlueString
 case zapcore.InfoLevel:
  colorize = color.GreenString
 case zapcore.WarnLevel:
  colorize = color.YellowString
 case zapcore.ErrorLevel, zap.PanicLevel, zap.DPanicLevel, zap.FatalLevel:
  colorize = color.RedString
 default:
 }
 enc.AppendString(colorize(lv.CapitalString()))
}
```

### Log Categories

Divided into framework-type logs and business-type logs.

- mio.sys (Framework logs)
- default.log (Business logs)

### Dynamically Enable Debug Logs

```go
// set
// IsDebugMode ...
func (logger *Logger) IsDebugMode() bool {
 return logger.config.Debug
}

// use
// Infow ...
func (logger *Logger) Infow(msg string, keysAndValues ...interface{}) {
 if logger.IsDebugMode() {
  msg = normalizeMessage(msg)
 }
 logger.sugar.Infow(msg, keysAndValues...)
}
```

### Dynamically Adjust Log Levels, in conjunction with dynamic config

```go
// AutoLevel ...
// / 动态调整zap等级
func (logger *Logger) AutoLevel(confKey string) {
 conf.OnChange(func(config *conf.Configuration) {
  lvText := strings.ToLower(config.GetString(confKey))
  if lvText != "" {
   logger.Info("update level", String("level", lvText), String("name", logger.config.Name))
   _ = logger.lv.UnmarshalText([]byte(lvText))
  }
 })
}
```

### Debug Log Information

The following six elements, namely configuration name, request URL, request parameters, response data, execution time, and line number, are referred to as the Debug "six-tuple" information.

## Unified Log Fields

Currently, referring to opentracing, and in the future, referring to otel log. The advantage of having a unified log format is that it makes it easier for operations and SRE teams to read. It can be troublesome if the formats are not consistent.
The implementation uses zap's fields:

- lv: Log level
- ts: Timestamp,log Timestamp
- msg: Log message
- aid: Application ID
- iid: Application Instance ID
- tid: Request Trace ID

### log Tracing ID

Print the Span ID.
By default, the framework automatically enables tracing and adds the `trace id` to all server or client `access` logs, so users don't need to worry about it. With our frontend framework, we can display the trace ID when an error occurs. Using the opentrace protocol, the trace is automatically added to the logs. The middleware sets the field `traceid` when loading the log.

### log the File Name and Function Name of the Called Log Method

### implement

```go
// DefaultZapConfig ...
func DefaultZapConfig() *zapcore.EncoderConfig {
 return &zapcore.EncoderConfig{
  TimeKey:        "ts",
  LevelKey:       "lv",
  NameKey:        "logger",
  CallerKey:      "caller",
  MessageKey:     "msg",
  StacktraceKey:  "stack",
  LineEnding:     zapcore.DefaultLineEnding,
  EncodeLevel:    zapcore.LowercaseLevelEncoder,
  EncodeTime:     timeEncoder,
  EncodeDuration: zapcore.SecondsDurationEncoder,
  EncodeCaller:   zapcore.ShortCallerEncoder,
 }
}


// 应用唯一标识符
func FieldAid(value string) Field {
 return String("aid", value)
}

// 模块
func FieldMod(value string) Field {
 value = strings.Replace(value, " ", ".", -1)
 return String("mod", value)
}

// 依赖的实例名称。以mysql为例，"dsn = "root:mio@tcp(127.0.0.1:3306)/mio?charset=utf8"，addr为 "127.0.0.1:3306"
func FieldAddr(value string) Field {
 return String("addr", value)
}

// FieldAddrAny ...
func FieldAddrAny(value interface{}) Field {
 return Any("addr", value)
}

// FieldName ...
func FieldName(value string) Field {
 return String("name", value)
}

// FieldType ...
func FieldType(value string) Field {
 return String("type", value)
}

// FieldCode ...
func FieldCode(value int32) Field {
 return Int32("code", value)
}

// 耗时时间
func FieldCost(value time.Duration) Field {
 return String("cost", fmt.Sprintf("%.3f", float64(value.Round(time.Microsecond))/float64(time.Millisecond)))
}

// FieldKey ...
func FieldKey(value string) Field {
 return String("key", value)
}

// 耗时时间
func FieldKeyAny(value interface{}) Field {
 return Any("key", value)
}

// FieldValue ...
func FieldValue(value string) Field {
 return String("value", value)
}

// FieldValueAny ...
func FieldValueAny(value interface{}) Field {
 return Any("value", value)
}

// FieldErrKind ...
func FieldErrKind(value string) Field {
 return String("errKind", value)
}

// FieldErr ...
func FieldErr(err error) Field {
 return zap.Error(err)
}

// FieldErr ...
func FieldStringErr(err string) Field {
 return String("err", err)
}

// FieldExtMessage ...
func FieldExtMessage(vals ...interface{}) Field {
 return zap.Any("ext", vals)
}

// FieldStack ...
func FieldStack(value []byte) Field {
 return ByteString("stack", value)
}

// FieldMethod ...
func FieldMethod(value string) Field {
 return String("method", value)
}

// FieldEvent ...
func FieldEvent(value string) Field {
 return String("event", value)
}

// FieldTid 设置链路id
func FieldTid(value string) Field {
 return String("tid", value)
}

// FieldCtxTid 设置链路id
// 存在循环依赖问题
// func FieldCtxTid(ctx context.Context) Field {
//  return String("tid", itrace.ExtractTraceID(ctx))
// }

// FieldCustomKeyValue 设置自定义日志
func FieldCustomKeyValue(key string, value string) Field {
 return String(strings.ToLower(key), value)
}

// FieldLogName 设置mio日志的log name，用于stderr区分系统日志和业务日志
func FieldLogName(value string) Field {
 return String("lname", value)
}

// FieldUniformCode uniform code
func FieldUniformCode(value int32) Field {
 return Int32("ucode", value)
}

// FieldDescription ...
func FieldDescription(value string) Field {
 return String("desc", value)
}
```

## rotate

refer to [lumberjack](https://github.com/sevenNt/lumberjack)

## Async and buffer write

With the help of Zap, we have implemented Async and buffer write.

```go
/// 3核心，配置zap并且调用zap.core生成
func newLogger(config *Config) *Logger {
 ///使用zapOptions，配置zap
 zapOptions := make([]zap.Option, 0)
 zapOptions = append(zapOptions, zap.AddStacktrace(zap.DPanicLevel))
 if config.AddCaller {
  zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
 }
 if len(config.Fields) > 0 {
  zapOptions = append(zapOptions, zap.Fields(config.Fields...))
 }
 /// 配置WriteSyncer。debug就stdout,其他模式使用retate写盘
 var ws zapcore.WriteSyncer
 if config.Debug {
  ws = os.Stdout
 } else {
  ws = zapcore.AddSync(newRotate(config))
 }
 /// 是否异步，异步使用buffer,
 if config.Async {
  var close CloseFunc
  ws, close = Buffer(ws, defaultBufferSize, defaultFlushInterval)

  xdefer.Register(close)
 }
 //设置zap的lvl
 lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
 if err := lv.UnmarshalText([]byte(config.Level)); err != nil {
  panic(err)
 }

 // encoderConfig := defaultZapConfig()
 // if config.Debug {
 //  encoderConfig = defaultDebugConfig()
 // }
 /// 配置encoder。debug使用console；其他模式使用JSON
 encoderConfig := *config.EncoderConfig
 core := config.Core
 if core == nil {
  core = zapcore.NewCore(
   func() zapcore.Encoder {
    if config.Debug {
     return zapcore.NewConsoleEncoder(encoderConfig)
    }
    return zapcore.NewJSONEncoder(encoderConfig)
   }(),
   ws,
   lv,
  )
 }
 /// 根据之前的core和option，new一个zap,然后放入Logger返回s
 zapLogger := zap.New(
  core,
  zapOptions...,
 )
 return &Logger{
  desugar: zapLogger,
  lv:      &lv,
  config:  *config,
  sugar:   zapLogger.Sugar(),
 }
}
```

## Improvement

Multiple channels for writing to the local disk.

## Log System

- Production & Collection
- Transmission & Splitting
- Storage & Retrieval

## Comparison of Four Ways to Generate Logs

- DockerEngine Direct Writing
- Business Direct Writing (Recommended for scenarios with a huge volume of logs)
- DaemonSet (Usually used in small to medium-sized clusters)
- Sidecar (Recommended for super large clusters)

## Collection

The most basic solution, which can handle the majority of cases, is to write logs to the local disk. This approach is suitable for multi-language applications since log SDKs may differ among languages, but writing to disk is universally established. Therefore, by writing to the disk and not going through the network or remote systems, only a log agent is needed to collect logs from containers. Filebeat can be used to collect logs from local files.

### In-Container Application Log Collection

Using Docker log, which is based on overlay2, the corresponding log files are directly fetched from the host machine.

### Transmission

Unified transmission platform based on Flume + Kafka.

### Log Splitting Based on LogID

1. General Level
2. Low Level
3. High Level (ERROR)

## Storage - Elasticsearch

Elasticsearch in a multi-cluster architecture:

- Log-level segmentation and high availability.
  Within a single data cluster:
  Master node + Data node (hot/stale) + Client node.

## Retrieval - Kibana

- Daily fixed-time hot-to-cold migration.
- Index creation is done one day in advance, managed based on templates and mapping.
- Retrieval is based on Kibana.

# Metrics

## Interface + Prometheus

There are also some open-source projects to reference, such as go-metrics.

Three types of monitoring interfaces: Counter, Gauge, and Observer.
Wrapping them with Prometheus for easy use.

- Counter: prometheus.CounterVec
- Gauge: prometheus.GaugeVec
- Summary: prometheus.SummaryVec
- Histogram: prometheus.HistogramVec

### Counter

Counter is the simplest counter and provides two methods: Inc and Add. It can only be used for counting increments.
It is commonly used to count the number of errors, request QPS, and cache hits in a service.

```go
type Counter interface {
    With(lvs ...string) Counter
    Inc()
    Add(delta float64)
}
```

### Gauge

Gauge is a metric used to indicate the current state of a service. It records a value that can increase or decrease over time.
Gauge is commonly used to monitor the current CPU usage, memory usage, and other similar metrics of a service. It provides a snapshot of the current state at a specific point in time.

```go
type Gauge interface {
    With(lvs ...string) Gauge
    Set(value float64)
    Add(delta float64)
    Sub(delta float64)
}
```

### Observer:Histogram+Summary

Observer is a more complex monitoring metric that provides additional information compared to the previous two. It can be used to observe and calculate the total value, count, and percentiles.
In Prometheus, Observer corresponds to the implementations of **Histogram** and **Summary**.

- Histogram is used to record the number of occurrences in different buckets or bins. For example, the number of requests in different time ranges. It indicates that the metric is stored in multiple buckets, so Histogram has almost no overhead.
- Summary, on the other hand, records values at different percentiles, calculated based on probabilistic sampling. For example, the 90th and 99th percentile response times. Since additional calculations are required, it incurs some overhead for the service.

## prometheus exemplar

A typical use case is to add trace information to metrics, allowing metrics and tracing to be correlated.  

By adding trace information to metrics, we can establish a relationship between the performance metrics of a service and the distributed traces generated by requests flowing through that service. This correlation allows for better understanding and analysis of system behavior and performance.  

For example, by adding a trace ID as a label to metrics, we can track the latency or error rate of specific requests and analyze their impact on overall system performance. This integration of metrics and tracing provides valuable insights into system performance, facilitates troubleshooting, and enables performance optimization.

> OpenMetrics introduces the ability for scrape targets to add exemplars to certain metrics.  
> Exemplars are references to data outside of the MetricSet. A common use case are IDs of program traces.

The target object exposes metrics and exemplar information through the metrics interface. When Prometheus pulls the data, it will fetch and store them together.

## Dashbroad

To manage and visualize metrics, we can use Grafana.
The goal is to consolidate all metrics into a single dashboard, avoiding the need for users to open multiple sources of information.

In the dashboard, we can organize metrics into different sections based on their nature and relevance:

1. Time-based metrics: Metrics related to time, such as request latency or response time.
2. Server metrics: Metrics related to server performance, such as CPU usage or memory usage.
3. Runtime metrics: Metrics related to the execution environment, like the number of threads or garbage collection statistics.
4. Kernel metrics: Metrics related to the underlying operating system, such as disk I/O or network activity.
5. Client metrics: Metrics from the perspective of the client, which can help identify issues with dependencies or external services. For example, if a client cannot connect to a memcached server, this perspective can help quickly identify and troubleshoot the problem.
6. Business-specific metrics or any other relevant metrics.

we can have consistent metric names while leveraging different labels (such as service name, environment, or machine name) to differentiate between different instances or environments.

Having a unified app ID for microservices provides a way to trace requests across the entire system, including upstream and downstream components, as well as assets. This helps in troubleshooting and identifying the root causes of issues.

## Profiling

- Open the profiling port online; by default, open using `go tool pprof`.
- Use service discovery to find node information and provide convenient ways to quickly visualize the profiling information of processes (such as flame graphs) in a web-based manner.
- Use a watchdog that triggers automatic collection based on signals such as memory and CPU; when an issue is detected, it automatically triggers data collection.

## Metrics Signals

### Google's Four Golden Signals

Google's experience with distributed systems monitoring, Four Golden Signals, refers to four metrics that are generally monitored at the service level: latency, traffic, errors, and saturation.
Google's Google SRE Books book puts forward the four golden indicators of system monitoring

Involving net (tcp, http, mq), cache, db, rpc and other resource types of basic libraries, the first monitoring dimension of the four golden indicators:

- Latency (time consuming, need to distinguish between normal and abnormal)
- Traffic (need to cover the source, i.e., caller)
- Errors (override the error code or HTTP Status Code)
- Saturation (how "full" the service capacity)

### CPU load

- Top 10 apps in the chart `topk(10,sum(rate(process_cpu_seconds_total{}[1m])) by (app))`
- Some app CPU `sum(rate(process_cpu_seconds_total{app="your app name"}[1m]) by (app))`
- `topk(3, max(max_over_time(irate[process_cpu_seconds_total{}[1m]](1d:))*100) by (job))`

### All requests

- All requests `sum(irate(mio_server_handle_total{}\[1m])))`
- Requests for a particular app `sum(irate(mio_server_handle_total{app="your app name"}\[1m]))`
- Chart top 10 apps `topk(10,sum (rate (mio_server_handle_total{}\[1m])) by (app))`
- aggregation app `sum(rate(mio_server_handle_total{}\[1m])) by (app)`

### Server-side metrics

The following are the metrics recorded by the framework, prometheus adds metrics to the collection, such as data from jobs or apps, so we can filter some of the application information data by him

### server-side counters

type There are three types

- http
- unary
- stream

method There are two types

- HTTP method is c.Request.Method+"." +c.Request.URL.Path
- The gRPC method is grpc.UnaryServerInfo.FullMethod

peer Fetching data

- HTTP's peer takes the app from the header header, this is the application name of the peer node
- gRPC's peer takes the app from the header, and this is the application name of the peer node.

code Fetch data

- HTTP code is the HTTP status code.
- The code of gRPC is the message returned by gRPC, only the system error code is recorded, the success of the system error code is OK, the non-system error code is recorded as `biz err`, to prevent Prometheus from exploding errors.

### server histogram

server_handle_seconds

type
method
peer peer node

type There are three types

- http
- unary
- stream

method There are two types

- HTTP method is c.Request.Method+"." +c.Request.URL.Path
- The gRPC method is grpc.UnaryServerInfo.FullMethod

Peer Fetching Data

- HTTP's peer takes the app from the header, which is the application name of the peer node.
- The peer for gRPC takes the app from the header, this is the application name of the peer node

### User-side front-end fe

Systems can be divided into two simple categories

- Resource Provisioning Systems - provide simple resources to the outside world, such as CPU, storage, network bandwidth.
- Service Provisioning System - Provides higher level processing capabilities for business related tasks, such as ticket booking, shopping, etc.

### USE metrics are system metrics that are not appropriate for the application

- CPU, Memory, IO, Network, TCP/IP status, etc., FD (and others), Kernel: Context Switch.
- Runtime: all kinds of GC, Mem internal state, etc.
  For resource-providing systems, there is a simpler and more intuitive USE criterion
- Utilization - Often expressed as a percentage of resource usage.
- Saturation - the degree of saturation or overloading of resources, an overloaded system often means that the system needs a secondary queuing system to complete the relevant tasks. This is related to the Utilisation metric above but measures a different situation, for example, CPU, Utilization is often the percentage of CPU used and Saturation is the length of the current queue of processes waiting to be scheduled for the CPU.
- Errors - This may be the error rate or number of errors in the use of resources, such as packet loss or bit error rate of the network, etc.

### red, service

For service-based systems, this is often measured in terms of RED

- Rate - the ability to fulfil service requests per unit of time
- Errors - Error rate or number of errors: the ratio or number of service errors per unit of time.
- Duration - the average duration of a single service (or the latency at which a user receives a service response).

# Tracing

base on otel

Use middleware for tracing
Usual sampling with 1/1000

## from the front-end to the database

- Before starting the `a` and `b` services, we configure the custom link attributes in the headers/metadata to be automatically collected and appended to the access logs. We can set this configuration using the environment variable `MIO_LOG_EXTRA_KEYS` with the value `X-Mio-Uid`. The framework will parse this environment variable and automatically append the specified custom link attributes in the component logs.
- When the `a` service receives a request from the client, it performs token validation. If the validation passes, it writes the value `9527` for the key `X-Mio-Uid` in the `http.Request.Context()`.
- Then, the `a` service passes the `http.Request.Context()` to the `gRPC` client.
- The `gRPC` client retrieves the value of `X-Mio-Uid=9527` from the `gRPC` context and writes it into the `gRPC` header before making a request to the `b` service.
- The `b` service receives the `X-Mio-Uid=9527` from the `gRPC` header and writes it into the `Context` before passing it to the `MySQL` client.
- The `MySQL` client retrieves the value of `X-Mio-Uid=9527` from the `Context`.
- Finally, each component responds and asynchronously writes the logs, appending the `trace-id` and `x-mio-uid` (the trace ID is automatically appended by the framework and does not require user configuration).

In summary, the `X-Mio-Uid` value passed through the `http.Request.Context()` and `gRPC` headers will be collected and appended to the logs by the framework. Along with the automatically appended trace ID, these attributes provide useful information for tracing and debugging purposes.
![](/img/Pasted%20image%2020230721113955.png)

## Combining with Logs

By using metadata, we can implement the integration of request logs and linkage. Request logs in linkage have more metrics and can provide granularity down to the function level. The most important aspect is having a globally unique `trace id` that can correlate the timing and error conditions across different systems. However, the challenge lies in the fine granularity of the data, resulting in a large amount of data and the need for sampling, which may lead to missing some data. Additionally, it cannot aggregate certain error information for error convergence.  

On the other hand, traditional framework request logs can record error information and enable error convergence, but they lack the excellent features of linkage.  

We can see that both logs and linkage have the common goal of recording request logs, but their logs cannot be shared, which leads to significant resource waste. In practical production environments, we do not necessarily need the fine granularity of linkage down to the function level. Generally, service-level linkage can satisfy our needs for troubleshooting.  

By unifying these two aspects, we can let logs have the aggregation and error convergence features, while also leveraging the advantages of linkage. In this context, we can enhance the `access` logs in the `mio` framework with linkage features, enabling them to record `trace id` for correlating various services and allowing users to customize fields such as `uid` and `orderId` to be logged.  

## Combining with Metrics

By using exemplar in Prometheus, we can integrate with metrics. The foundation of the integration lies in the uniform encapsulation provided by the client, which solves the following problems:  

# client

Provide a unified encapsulation for basic components.

1. Loading configuration  
2. Handling context  
3. Logging  
4. Supporting opentracing  
5. Collecting Prometheus metrics
