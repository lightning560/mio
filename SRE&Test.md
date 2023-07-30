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

# Syllabus

## SRE

- sli,slo
- oncall
- Pre-failure
- Fault happening
- After fault

## Test

- Util test
- Api test
- Integration test
- Stress test
- Full chain stress test

# SRE methodology

- Ensure long-term focus on R&D
- All product incidents should be summarised, whether an alarm is triggered or not.
- Not many companies get alarms right. Alarms do hierarchy, auto-generated work orders and phone calls to engage scriptures.
- Alarm accuracy and recall. google SRE's second book talks about alarms. Get the stats right first, finally do aiops.
  Maximise iteration speed while safeguarding service SLOs.
- Bug budget, release strategy.

## sli

sLI, Service Level Indicator, is really about which metrics we choose to measure our stability.

### Two principles for choosing SLIs

Principle 1: Choose indicators that can identify whether a subject is stable or not. If they are not indicators of the subject itself, or cannot identify the stability of the subject, they should be excluded.

Principle 2: For e-commerce and other business systems with user interfaces, give priority to indicators that are strongly related to user experience or that can be clearly perceived by users.

### Google Methods VALET Selection SLI

Volume, Availability, Latency, Error and Ticket

- Volume is the throughput of the data platform; QPS, TPS and so on.
- Availability
- Latency
- Error: Ticket.
- Tickets: The SLO of Tickets can be imagined as its Chinese meaning: tickets. In a cycle, the number of tickets is fixed, for example, 20 per month, each time the manual intervention, one will be consumed, if consumed, still need to manually intervene, then it is not up to the standard!

## slo

SLO, Service Level Objective, refers to the stability goal we set, such as "a few 9s".
Maximise iteration speed while maintaining service SLOs.

### Bug budget management

### Penalties for failures that don't make it to the SOP

### Multiple development tools

How does a software engineer cope with duplication of effort?
For example, without tools, codereview and so on certainly can't get off the ground.

## oncall

- Services for end users are 5 minutes long. For non-sensitive services it's usually 30 minutes.
- During an oncall, bring a laptop, an internet card, a charger and a phone.
- Multiple channels to receive alerts (not limited to email, no one is available, SMS, automated phone calls. The phone does not answer, call BACKUP, and then do not answer to call the boss)
- Response time is related to the reliability of the business, if the service is 99.99% then there is 13 minutes of unavailability per quarter, so oncall engineers have to respond to production incidents on a minute to minute basis (and more importantly, should be self-healing)
- Once an alarm is received, the engineer must acknowledge (ack), be able to locate the problem in time and try to resolve it, and escalate for support if necessary (should escalate faults, not mess with them themselves).
- Generally, the main oncall person is on duty, the deputy oncall as an auxiliary, usually the team can also be used as each other's deputy oncall, each other on duty, and share the work pressure.
- During oncall duty, on-call engineers must have enough time to deal with emergencies and follow-up work, such as writing incident reports (Mao Jian has issued and SRE has fault templates).
- When faced with a challenge, a person will actively or inactively choose the downlink way to deal with it:
- Rely on intuition, automate, act quickly
  - Rational, focussed, conscious cognitive type activities
- When dealing with loaded systems, the second way of doing things is better, and is likely to produce better results and a more well-planned execution.
- Operating intuitively and responding quickly both seem like useful approaches, but each of these approaches has its own drawbacks. Step-by-step problem solving when there is enough data to support it, while constantly reviewing and validating all current assumptions.
- Oncall can seek external help (do the information synchronisation first, the boss is more worried about getting it right. Being a leader is helping you solve the problem and make it better)
- Clearly defined problem escalation routes
  - Clearly defined steps for dealing with emergencies
  - No finger pointing, no blame culture
- System is too stable and prone to laxity, regular rotations and disaster recovery drills

# Pre-failure preparation

## codeReview

Be sure to have a change log. And add the error rate to the release key point.
Code viewer should be asynchronous. Don't have a bunch of code viewers, it's embarrassing and a waste of time.

## Resource deployment

## A combination of change management and capacity planning

## Capacity planning + traffic forecasting

Much better if it's in the cloud
Demand Forecasting and Capacity Planning: Natural Growth + Unnatural Growth

- There must be an accurate natural growth demand forecasting model, and the demand forecast should exceed the time of resource acquisition.
- There must be prepared statistics on the sources of unnatural growth demand in the planning.
- There must be periodic stress tests to readily reconcile raw system resources with business capacity.

### Efficiency and Performance

- Continuous optimisation of resource utilisation reduces the total cost of ownership of the system.
- Adequate capacity is deployed and maintained according to a pre-determined latency target.

### Intervene when machine costs are high

## Avoid single points, redundancy

At least n+2

### Network redundancy

For example, cdn.

## Change management

70% of production incidents are triggered by changes

- Use a progressive release mechanism
- Detect problems quickly and accurately.
- When problems are detected, changes are rolled back safely and quickly.

## Failure planning playbook best practices

After the release, there is a sop and check list, you can't just look at the error, you have to look at the warning, don't peak the release.

## Code change, release management, change management

70% of problems are caused by changes, restoring usable code is not always a bad thing (rollback first, then check.). Ask first if there has been a recent release if there is a problem)
Change management: 70% of production incidents are triggered from changes

- Use an incremental release mechanism
- Detect problems quickly and accurately
- Safely and quickly roll back changes when problems are detected

## chaos

- Any dependency can fail, do chaos monkey testing, inject failure testing. google internal dirt, sla index exceeded, make a little failure does not matter

## Ops security

Key management, vault.
Firewalls are used to isolate the environment.
Fortresses.
Scripts to regularly scan for vulnerabilities, cve checker.

## Dependency management

Dependency Analysis
Critical modules do not depend on non-critical modules

## Failure rehearsal chaos

## Full-link stress testing

Extreme stress testing + Failure walkthrough.
With the automated stress testing platform, sla has improved a lot, especially in the big promotion.

## Daily inspection

I've been working on the moring checklist for a while now, and I've been checking every morning.
A failure at 8 o'clock took a long time to recover because all the staff were on their way to work.
One person from each line of business, one person arrives at the company one hour early every day to observe the system metrics and find all kinds of problems. 1000 problems were found in half a year.

Periodic cluster health check

## Self-healing

- Overload avoidance.
- Overload protection, traffic scheduling, etc. At its core, offloading traffic
- Elegant Degradation:
- Lossy service to avoid core link dependency failures. Always rehearse.
- Retry Retreat.
- The fallback algorithm, freezing time, API retry detail control policy. response can bring back some policy, faster than the config.
- Timeout control.
- Intra-process + inter-service timeout control. Only the whole link is valid .

Multi-live offsite
client flow limiting, based on health, inflight, latency, and so on.

Remove ddos nodes, use gslb.

## observer monitoring system

alert, ticket, logging, tracing, metrics

### Alert

The second book of Google Sre talks about alerting.
Not many companies do alarms well. Alerts are hierarchical and automatically generate work orders and phone calls.
Accuracy and recall of alerts.

Get the stats right first, then do aiops.

## golang framework HA module

retry,exponential backoff+jitter,retry tag.
timeout, in-process+grpc cross-process, full-link.
Distributed flow-limiting quota,max-min fairness, e.g. QuotaServer failure, consider downgrading to local policy or even release;

rate limiting + overload protection, based on sliding window computing + cpu, using queues for management, codel controlled delay algorithms, discard part of the flow
degradation, touting scheme
Meltdown,google sre,max(0, (requests- K\*accepts) / (requests + 1))

probes,livenessProbe,readinessProbe
Defensive programming
Isolation + node grouping
Load balancing power of two choices

# Fault happening

## Core metrics for fault response

MTTF + MTTR

## Stop loss before repair

- Restart

- Rollback

- Eliminate harmful traffic. If it's an app it's global traffic, use fusion for a server, gateway for an interface

- Have a preplanned playbook for the best way to do this.

### Manual scaling or auto scaling

## Effective troubleshooting

- Generic troubleshooting process + sufficient knowledge of the system in which the failure occurred (warning phenomena, analysis of sub-causes).
- The troubleshooting process is iterative and uses "assume - rule out" (keep the site, reboot all, keep one, e.g. take off traffic.).
- When an alarm is received, find out how serious the problem is.
- For large problems, immediately declare an all-hands meeting.
  - The first reaction of the majority of people is to immediately start the troubleshooting process, trying to find the root cause of the problem as soon as possible, the correct approach is: do everything possible to get the system back in service and stop the damage
  - When locating a problem quickly: Save the problem site, e.g., logs, monitoring, etc.
- A monitoring system records system-wide monitoring metrics, and a good dashboard makes it easy to quickly pinpoint problems, such as Moni.
- Logs are another invaluable tool. Logs record information about each operation and the corresponding system state, allowing you to understand what the entire component is doing at a certain moment, such as Billions.
- Link tracing tools, such as Dapper
- Debug clients, so you know exactly what information the component is returning when it receives a request (well in advance).
- One last modification: a system that works fine until some external factor comes along.
- A configuration file change, a change in user traffic, checking for recent modifications to the system may be helpful in finding the root cause of the problem.

### Chained failures are prioritised

1. optimal availability of releases;
2. a reliable apm tracing system;
3. Reliable and unified metrics dashboard (we have all microservices and all languages, we have unified and converged the monitoring metrics, so that O&M can diagnose multiple microservices across multiple microservices and the dashboard is consistent;

## Emergency response

- Don't panic, you're not alone, the whole team is involved!
- If you feel like you're struggling, get more people involved.
  - - Notify other departments within the organisation of the current situation
- Conduct regular disaster management and emergency response drills (7am drill every morning to see how everyone handles it on a daily basis)
- Always test rollback mechanisms first in large tests (often go live without a rollback plan)
- Emergency response should allow others to get clear and timely updates on the state of affairs (customer service, PR, leadership should all know)
- If you can't think of a solution, then demand help on a larger scale. Find more team members and ask for more help, but do it fast.

## Emergency Incident Management

- Reduce emergencies without process management
- Don't focus too much on technical issues
  - Miscommunication (recovering without talking about it)
  - do not ask for help
- Incident master, transactional team, spokesperson (when there is an incident, an incident master steps in immediately)
- when to announce the incident to the public
- Does a second team need to be brought in to help deal with the problem?
  - Is the incident affecting end users? (If it is affecting users, immediately colleague various departments)
  - Is the problem still unresolved after an hour of focused analysis?
- Prioritise: contain the impact and restore service while preserving the site for root cause investigation.
- Prepare in advance: Prepare a process in advance with all incident participants.
- Trust: Trust everyone involved in the incident, assign responsibilities and let them take the initiative (keep emotions in check).
- Reflection: Pay attention to your emotional and mental state during the incident. If you find yourself starting to panic or become overwhelmed by the stress, you should ask for more help!
- Consider alternatives: periodically revisit the situation and reassess whether the current work should continue or whether something more important or urgent needs to be done (if it turns out that it cannot be resolved).

# After a failure

All product incidents should be summarised, with or without triggered alarms.

## Failure Summary

- After an emergency, set aside some time to write an incident report.
- There is no better learning material than the records of past incidents. Publish and maintain failure reports.
  - Be honest and meticulous in your documentation and always look for ways to tactically as well as strategically avoid the incident.
  - Ensure that you and others actually do the things that are summarised in the incident.

### Use a fault template

- Can this be measured?
- Is it an alarm?
- How to avoid it
- How to recover in less time.
- How to avoid recurrence.
- Can the system be designed to avoid this problem.

Eventually write todo and then periodically backtrack on this issue

## The review process

### Purpose: Learning from Failure

- Learn from the past, not repeat it
- Avoid blame, provide constructive feedback
- Visible downtime or degradation of quality of service to a certain level.
- Any type of data loss
- Incidents that require manual intervention by on-call engineers (including rollbacks, switching user traffic, etc.)
- Problems that take longer than a certain amount of time to resolve
- Monitoring issues (signalling that the problem was detected manually rather than by the alarm system)

### The three golden questions of failure review

First question: What were the causes of the failure?
Second question: What did we do and what can we do to ensure that a similar failure will not occur next time?
Third question: What could we have done at that time to restore the business in a shorter time if we had done something?

### Failure review meeting

Firstly, it will be done in conjunction with Timeline, that is, according to MTTI,MTTK,MTTF and MTTV to do a classification first;
Then for the long time-consuming links, we will discuss the improvement measures over and over again;
Finally, set the responsible person and time point, and then continuously follow up the implementation status.

The role of the meeting facilitator is generally the same as the CL (Communication Lead) we mentioned in the previous section, which is called technical support in our company.
In order to get to the ground, it is important to define who should have the primary responsibility for improvement. Note that I don't use the term "who should take the main responsibility", but rather who should take the main responsibility for improvement, i.e., who should implement the improvement measures.

### Failure Determination Principles

There are a number of principles for determining faults, the most important of which are the following three

1. Robustness principle. This principle means that each component should have certain self-healing capabilities, such as master-standby, clustering, flow-limiting, degradation and retry, etc. The principle of robustness means that each component should have certain self-healing capabilities. For example, in a state where B is dependent on A, the dependent party A has a problem but is able to recover quickly, while the dependent party B is not able to recover quickly, leading to the spread of faults. In this case, it is the relying party B that bears the main responsibility, not the relying party A. 2.
2. Third party is not responsible by default. If third-party services are used, such as public cloud services, including IaaS, PaaS, CDN, video, etc., our principle is that the third party is not responsible by default.
3. Segmented Determination Principle.

This principle is mainly used in more complex scenarios. The fundamental starting point of these principles is to abandon the idea that there is only one root cause.

# test

## test pyramid

70 small 20 medium 10 large tests

Four ways: unit tests, interface tests, integration tests, performance tests.
Some unit tests are run locally via docker-compose, and unit tests are run on committed code using GitLab CI.

Interface tests use the set of test cases in the interface platform described above.

There are two main types of performance tests.
One is benchmark using GitLab CI.
The other type is the full-link pressure test using the platform tools.

Integration testing is currently not done well enough, before the GitLab ci to pull images, through the dind (Docker in Docker) to run the entire process, but there is no topology map, so you need to manually configure yaml, very cumbersome, is currently being combined with the configuration of the Centre's dependency topology map, ready to use jekins to complete the integration test.

## golang module management internal package

1. GOPATH
2. vendor
3. go model, 1.11 onwards, 1.13 on by default
   goproxy+goprivate in-house best practices

# unitTest

Test case management platform

## unitTest Basic Requirements

- Fast
- Consistent environment
- Arbitrary order
- Parallelism

### AIR principle

AIR, short for Automatic, Indepenndent, Repleatable:

- Automatic: automation, unit testing should be automatic execution, automatic verification, automatic results, if you need to manually check (such as the results of the output to the console) of the unit test, not a good unit test;
- Indepenndent: independence, unit testing should be able to run independently, no dependency between test cases and the order of execution, the use of cases within the external resources are not dependent;
- Repleatable: repeatable, unit tests should be able to be executed repeatedly, and the results are stable and reliable every time;

## The traditional way

- Interface Oriented Programming
- Dependency Injection, Control Inversion
- Using Mock

## business server unitTest

Use go's official: Subtests + mock to complete the entire unit test.

- /api
    Ideal for integration testing, testing APIs directly, using API testing frameworks (e.g. yapi), and maintaining a large number of business test cases.
- /data
    docker compose emulates the underlying infrastructure as it is, so you can remove the _infra_ abstraction layer.
- /biz
    Depends on _repo_, rpc client, and uses mock to simulate the implementation of _interface_ for business unit testing.
- /service
    Dependent on _biz_ implementation, build _biz_ implementation class to pass in for unit testing.

Develop _feature_ based on _git branch_, perform unittest locally, then submit _gitlab merge request_ for _CI_ unit testing, build based on _feature branch_, complete functional testing, then merge _master_, perform integration testing, and perform regression testing after going live. Regression testing.
Without integration tests, it's difficult to trust the end-to-end operation of a web service.

http server test Using "net/http/httptest"

```go
type HTTPMock struct {
 *gin.Engine
}

func NewHTTPMock() *HTTPMock {
 gin.SetMode(gin.ReleaseMode)
 return &HTTPMock{
  gin.New(),
 }
}

func (m *HTTPMock) MockPostByStruct(uri string, param interface{}) (*http.Response, error) {
 postByte, err := json.Marshal(param)
 if err != nil {
  return nil, err
 }
 return m.MockPost(uri, postByte)
}

func (m *HTTPMock) MockPost(uri string, param []byte) (*http.Response, error) {
 // 构造post请求
 req := httptest.NewRequest("POST", uri, bytes.NewReader(param))
 req.Header.Set("Content-Type", "application/json")

 // 初始化响应
 w := httptest.NewRecorder()

 // 调用相应handler接口
 m.ServeHTTP(w, req)

 // 提取响应
 result := w.Result()
 defer result.Body.Close()
 return result, nil
}

func (m *HTTPMock) MockGet(uri string) *http.Response {
 // 构造get请求
 req := httptest.NewRequest("GET", uri, nil)
 // 初始化响应
 w := httptest.NewRecorder()

 // 调用相应的handler接口
 m.ServeHTTP(w, req)

 // 提取响应
 result := w.Result()
 defer result.Body.Close()
 return result
}
```

## library

[github.com/stretchr/testify](https://github.com/stretchr/testify) assert
`github.com/cweill/gotests` table test
`github.com/axw/gocov`
`github.com/smartystreets/goconvey/` runtime test、assert

## code gen

Generating an interface using strct2interface
Using mockery to generate unit tests based on interfaces
[Examples - mockery (vektra.github.io)](https://vektra.github.io/mockery/examples/)

## mock

[GitHub - golang/mock: GoMock is a mocking framework for the Go programming language. (kgithub.com)](https://kgithub.com/golang/mock)
To achieve these two goals, we set two rules:

- All external http requests are mocked.
- mysql, memcache, and redis are served directly, and each test case maintains its own test dataset.

### External http mock

This is based on [jarcoal/httpmock](https://github.com/jarcoal/httpmock "A")

### grpc mock

[GitHub - bilibili-base/powermock: Support for gRPC/HTTP protocol, feature-rich Mock Server implementation. (kgithub.com)](<https://kgithub>. com/bilibili-base/powermock)

### infra mock:docker compose

Using docker-compose.yml
By building `MySQL` in `CI`, and defining internal `install` and `initialize` criteria, we were able to easily build the dependent data sources together, making it easy to do testing.

Based on docker-compost to achieve cross-language dependency management solution to solve mysql, redis (because these infra mock is very complex) grpc mock drop test/ directory contains db.sql compose.yaml

# api test

Use postman, apifox tool or yapi tool
[YApi Interface Management Platform (smart-xwork.cn)](http://yapi.smart-xwork.cn/doc/index.html)

# integration test

Using ginkgo
github.com/onsi/ginkgo/v2 v2.6.1
[onsi/ginkgo: A Modern Testing Framework for Go (github.com)](https://github.com/onsi/ginkgo)
github.com/onsi/gomega v1.24.2

# Full chain stress test

## refer to Uber and some SaaS multitenancy

uber multi tenancy
<https://eng.uber.com/multitenancy-microservice-architecture/>

uber Real-Time Dynamic Subsetting
[Better Load Balancing: Real-Time Dynamic Subsetting | Uber Blog](https://www.uber.com/blog/better-load-balancing-real-time-dynamic-subsetting/)
[Efficient and Reliable Compute Cluster Management at Scale | Uber Blog](https://www.uber.com/blog/compute-cluster-management/)

## problem:Unable to emulate

Allowing multiple systems to coexist is one of the most effective ways to make mircoserver stable and modular Integration testing requires multiple systems to coexist. Integration testing requires multiple systems to coexist. Many companies have built multiple environments, and it's hard to guarantee that nothing will go wrong.
Parallel testing requires staging over the environment. Test environments can't be stress-tested and can't be emulated.

### Multi-tenancy Essence

Cross-service delivery of requests with context, data-isolated traffic routing scheme.
Use service discovery to register tenant information and register as a specific tenant.

## Modification system

In kafka, through the meta data or different topic. all have to do transformation, the amount of change is very large.
log, the most important thing is that the log needs redis, db need to shadow database schema consistent mq do topic bank SMS and other third-party mock tuning
The effect of a set of k8s countless sets of environment

### Solution: Traffic Coloring

a-z in only b service changed, only traffic routing b. traffic routing, add a colour field to the registry. The default is to take the colour of null Isolation: the need for reliable services
Grey scale testing has a cost, multi-tenancy has almost no cost.

### example:a to b

There are 2 parts in the LB of a, one for fat1 and one for fat2, and one request for a special field in the header.
A request for a special field in the header. b to c?
It's very cool after you're done. The test account is also isolated.
Need to do pressure testing of the shadow library.
The cost of grey scale testing is costly and affects 1/N users. Where N is the number of nodes.

Bind context to inbound requests (e.g., http header), pass context in in-process, and pass metadata across services (e.g., opentracing baggage item). In this architecture, each infrastructure component understands the tenant information and is able to isolate the traffic based on the tenant's routing, and at the same time, in our platform, we can allow the use of different types of applications running on different platforms. Our platform allows more control over running different microservices, such as metrics and logging.

Typical base components in a microservice architecture are logging, metrics, storage, message queues, caching, and configuration. Segregating data based on tenant information requires separate handling of the base components.
