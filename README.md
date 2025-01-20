# zoomies

Zoomies is a command line tool that enables performance monitoring of your latency, upload and download speeds through the [fast.com](https://fast.com/) web service. Checkout `docs/papers` for the two inspirational papers that helped provide context for ISP speed tests making this project possible!

### Building the Application

Clone the repository to your local directory:
```
git clone https://github.com/primlock/zoomies.git
``` 

Build the binary:

```
go build -o zoomies main.go
```

Run the binary with `./zoomies`, passing any of the available parameters.

### Passing Parameters

Zoomies supports a growing list of parameters that you can pass to the application to modify the test structure and results.

```
Usage:
  zoomies [flags]

Flags:
  -b, --binary         display the unit prefixes in binary (Mibit/s) instead of decimal (Mbps)
  -d, --duration int   the length of time the test should run for (3-30 seconds) (default 15)
  -h, --help           help for zoomies
      --nodownload     skip the download test
      --noupload       skip the upload test
  -p, --pings int      the number of pings sent to the server in the latency test (1-5) (default 3)
  -t, --token string   user provided api endpoint access token
      --verbose        provide additional information from the logger
```

### Contributions

If you would like to contribute to the project or see an issue you would like to fix PR's are welcome!