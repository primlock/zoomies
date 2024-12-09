# Zoomies

Use the fast.com web service to measure user download and upload speeds through the CLI.

This [blog](https://netflixtechblog.com/building-fast-com-4857fe0f8adb) from Netflix describes the Fast.com project and some of it's engineering.

This [blog](https://about.netflix.com/en/news/fast-com-now-measures-latency-and-upload-speed)  talks about including upload speeds and latency as well.

### TODO's

[*] Implement the code to get a token from the web page
[*] Define a structure for the JSON object representing the client and the target.
[*] Remove the unicode insertion in the JSON object and test to see if we can download the speedtest file from our returned JSON object target link.
    - \u0026 is being inserted. That is the unicode character for &.
    - Replacing the \u0026 allows us to download the file.
[*] Extract the query components from the target URL 
[*] Insert the extracted query components into the struct
[ ] Create a Test Configuration structure that can be reused for upload and download
    [ ] Implement the config
[ ] Do the upload and download tests concurrently
    [ ] Research and implement context
[*] Implement a CLI library to run the speed tests
    [*] Define options for the users to configure before running the test
[ ] Provide your own custom list of servers

### Targets Object

Breaking down the returned value from the HTTP Get request to api.fast.com/netflix/speedtest/v2.
https://ipv4-a999-xyz929-verizon-isp.1.oca.nflxvideo.net/speedtest?c=us\u0026n=888\u0026v=16\u0026e=328913283\u0026t=ZYEBuBdEbWRQVKCWsaWHcgUBpJSXyqHKhDDswk

* Server: ipv4-a999-xyz929-verizon-isp.1.oca.nflxvideo.net/
* Speedtest File: speedtest/
* Client Parameters: 
    * c: us
    * n: 888
    * v: 16
    * e: 328913283
    * t: ZYEBuBdEbWRQVKCWsaWHcgUBpJSXyqHKhDDswk

### Notes

* We have 2 options when making a GET request to a target link in the JSON object:
    * After ../speedtest we can specify a range (../speedtest/range/0-2048) to download a portion (the range of bytes we specified) of the speedtest file.
    * If we do not specify a range, we will get the entire file.

* The JSON object is using the unicode sequence \u0026 instead of the character & when our client requests the JSON object.

### Designing the Speeds Tests

#### General Guidelines

* Consistency: Use the same data size for download (GET) and upload (POST) tests across servers
* Warm-Up Requests: Perform a few "warm-up" requests (e.g., 1-2) to allow for caching and to stabilize initial fluctuations.
* Sample Size: Use multiple requests (e.g., 3-5) to the same server and calculate the average speed. More requests reduce variability but also increase test time.
* Concurrency: Test sequentially or with a limited number of concurrent requests to avoid overwhelming your network or server.

#### Download Recommendations

* Data Size: Use larger files (e.g., 10-50 MB) to minimize the impact of connection setup time (TCP handshake, DNS lookup).
    * speedtest file is 26 MB.
* Number of Requests: 3-5 requests per server. You can increase this if you notice significant variability in results.
* Timeouts: Use a reasonable timeout (e.g., 30 seconds) to handle slow servers gracefully.
* Protocol Efficiency: Consider HTTP/2 or HTTP/3, which improve performance over long distances.

#### When to stop testing?

* We could approach this in a couple of ways.
    1. We tell the program how many times to attempt to download a file from the server and average the speed.
    2. We tell the program to make as many requests as you can in a limited amount of time and then we average the speed.