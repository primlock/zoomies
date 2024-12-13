# SamKnows Test Methodology White Paper

Highlights extracted from SamKnows Test Methodology white paper.

#### Architecture

* On startup the application performs a latency test on a list of nearby servers to determine the optimal test server. The server with the *lowest round trip latency* is selected as the candidate for the test.

* The server examines the IP address of the client and uses geo-location databases such as MaxMind to find the location and ISP of the user.

* Test measurement results are submitted back to SamKnows for analytics.

#### Performance Tests

* The download and upload transfers are performed over one or more concurrent TCP connections and measured in bits per second.

* If the client supports WebSockets then these are used as the transport for all measurement traffic.

* During the download test, the client will fetch chunks of data from the server and discard them. During the upload test, the client will generate the payload and send it to the server where it is then discarded.

* The download and upload speed tests operate for a fixed-duration specified in seconds. A max limit on transfer volume may be imposed if data volumes are of a concern.

* Both download and upload tests will dynamically scale the number of parallel TCP connections. 6 are used by default but this may scale up to 32 parallel connections to support the fastest broadband lines (1 Gbps).

* TCP slow start are accounted for and removed from the main test results. The test's "warm-up" period is meant to account for the impact of the TCP slow start and to determine how many parallel connections will be used for the remained of the test.

* The round trip time of the smallest possible packet between the client and the server is recorded in microseconds. Due to the test operating over TCP, it is not possible to capture packet loss.