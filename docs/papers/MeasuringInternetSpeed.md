# Measuring Internet Speed - Feamster & Livingood

Highlights extracted from the Feamster and Livingood paper on Measuring Internet Speed.

### Introduction

In the past, some speed testing tools made the assumption that the most constrained link (bottleneck) was the ISP last mile access network. The performance bottleneck has often shifted from the ISP access network to a user's device, wifi network, network interconnections or speed testing infrastructure.

There can be many factors that influence the results of a test which include:
- Age of the device
- Interconnect capacity
- Test server capacity
- User access link activity during the test

### Performance Metrics

* When people talk about internet *speed*, they are generally talking about **throughput** (both downstream and upstream). This is the amount of data that can be transferred between two network endpoints over a limited amout of time. The end-to-end performance is typically measured with a collection of metrics, namely throughput, latency and packet loss.

* Throughput is *not* a constant. It changes from minute to minute based on many factors including what other users are doing on the internet.

* **Jitter** is the variation between two latency measurements. Large jitter measurements are problematic.

* **Packet loss rate** is typically computer as the number of lost packets divided by the number of packets transmitted. High packet loss rates generally correspond to worse performance, some packet loss is normal because a TCP sender typically uses packet loss as the *feedback signal* to determine the best transmission rate. Certain network design choices such as increasing buffer sizes can reduce packet loss but at the expence of latency, leading to a condition known as *buffer bloat*.

* A speed test sends traffic that traverses many network links including the wifi link inside the home, the link from the ISP device to the ISP network and the many network level hops between the ISP and the speed test server. The throughput measurement that results from such a test reflects the capacity of the *most constrained* link, referred to as the *bottleneck* link.

Most speed tests use TCP with multiple parallel TCP connections. Any TCP based test should:
- Be long enough to measure steady-state transfer
- Recognize that TCP transmission rates naturally vary over time
- Use multiple TCP connections

* TCP has a **slow start phase** where the transmission rate is far lower than the network capacity. Including this data in a throughput calculation will result in a measurement that is less than the actual available network capacity. If the speed test is too short, the test will tend to *underestimate* throughput. TCP throughput continually varies because the sender may implement **additive-increase/multiplicative-decrease (AIMD)** which is a feedback control algorithm designed to increase the transmission rate (window size), probing for usable bandwidth, until loss occurs.

### Limitations of Existing Tests

* As network access links have become faster, the network bottleneck has moved from the ISP access link to elsewhere on the network. The bottleneck may have moved to any number of places, from the home wireless network to the user's device itself.

### User Related Considerations

* Speed tests that are run over a home wireless connection often reflect a measurement of the user's home wireless connection. because the wifi network is usually the lowest capacity link between the user and the test server.

* Users continue to user older network devices as their hardware that do not support higher speeds. If a user has a 100 Mbps ethernet card which is connected to a 1 Gbps internet connection, the users speed will never exceed 100 Mbps due to the hardware limitation.

### Test Infrastructure Considerations

* The throughput of an internet speed test will depend on the distance between the client and the server endpoints which is measured by a packet's **round trip time (RTT)**. TCP throughput is inversely proportionate to RTT between two endpoints (client and the server). In order to obtain the highest level of throughput, having a geographically close server is key but other factors such as concurrent server activity can make an impact.

* Using as many parallel TCP connections as possible is a huge advantage in a speed test as the goal is to move as much data as possible in the limited amount of time to reach the available link capacity. 

* The length of the test and the amount of data transferred also significatly affects the results. A TCP sender doesn't immediately begin sending traffic at full capacity. It begins the TCP slow start until the sending rate reaches a pre-configured threashold value at which point AIMD begins. If the test is too short, the results will include much of the TCP slow start and only a small percentage approaching the available capacity.

* Estimating the throughput of the link is not as simple as dividing the amount of data transferred by the total time elapsed over the course of the transfer. A more accurate estimate of the transfer rate would instead measure the transfer during steady-state AIMD, without the TCP slow start.

### The Future of Speed Testing

* Measure to multiple destinations. It may make sense to perform active speed test measurements to multiple destinations simultaneously, to mitigate the possibility that any single destination or end-to-end network path becomes the network bottleneck.