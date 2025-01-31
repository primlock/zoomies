package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/primlock/zoomies/api"
	"github.com/primlock/zoomies/internal/logger"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type RemoteServerResponse struct {
	Client  api.Client   `json:"client"`
	Targets []api.Server `json:"targets"`
}

type Candidate struct {
	Server api.Server
	RTT    time.Duration
}

type Parameters struct {
	// The endpoint used to gather testing server information.
	APIEndpointToken string

	// The option to skip the download speed test.
	NoDownload bool

	// The option to skip the upload speed test.
	NoUpload bool

	// Configurations that apply to download, upload and latency tests.
	Config *TestConfig

	// Provide additional information to the user from the logger
	Verbose bool
}

type TestConfig struct {
	Timeout int

	// The amount of time the download and upload test runs for in seconds
	Duration int

	// The number of pings sent to the server in the latency test
	PingCount int

	// The number of concurrent HTTP request being made to download and upload
	ConcurrentRequests int

	// Determines whether the unit prefixes are displayed as decimal (Mbps) or binary (Mibit/s)
	BinaryUnitPrefix bool
}

var (
	ErrUnknownAppToken      = errors.New("invalid token passed as a parameter")
	ErrDurationOutOfBounds  = errors.New("duration must be in the range 3-30 inclusive")
	ErrPingCountOutOfBounds = errors.New("ping must be in the range 1-5 inclusive")
	ErrNoCandidatesToRank   = errors.New("the candidates object supplied was nil")
)

var log = logger.TLog

const (
	CommandName               = "zoomies"
	CommandDescription        = "zoomies is a network speed measurement tool"
	UploadTestPayloadSize     = 25 * 1024 * 1024 // 25 MB
	DefaultTestServerCount    = 5
	DefaultNoDownload         = false
	DefaultNoUpload           = false
	DefaultTimeout            = 30
	DefaultDuration           = 15
	DefaultPingCount          = 3
	DefaultConcurrentRequests = 3
	DefaultChunkSize          = 26214400
	DefaultBinaryUnitPrefix   = false
)

func NewTestConfig() *TestConfig {
	return &TestConfig{
		Timeout:            DefaultTimeout,
		Duration:           DefaultDuration,
		PingCount:          DefaultPingCount,
		ConcurrentRequests: DefaultConcurrentRequests,
		BinaryUnitPrefix:   DefaultBinaryUnitPrefix,
	}
}

func NewParameters() *Parameters {
	return &Parameters{
		NoDownload: DefaultNoDownload,
		NoUpload:   DefaultNoUpload,
		Config:     NewTestConfig(),
		Verbose:    false,
	}
}

func NewCmd() *cobra.Command {
	params := NewParameters()

	cmd := &cobra.Command{
		Use:          CommandName,
		Short:        CommandDescription,
		SilenceUsage: true,
	}

	// Define the user provided params.
	cmd.Flags().StringVarP(&params.APIEndpointToken, "token", "t", "", "user provided api endpoint access token")
	cmd.Flags().BoolVar(&params.NoDownload, "nodownload", params.NoDownload, "skip the download test")
	cmd.Flags().BoolVar(&params.NoUpload, "noupload", params.NoUpload, "skip the upload test")

	cmd.Flags().IntVarP(&params.Config.Duration, "duration", "d", params.Config.Duration, "the length of time the test should run for (3-30 seconds)")
	cmd.Flags().IntVarP(&params.Config.PingCount, "pings", "p", params.Config.PingCount, "the number of pings sent to the server in the latency test (1-5)")
	cmd.Flags().BoolVarP(&params.Config.BinaryUnitPrefix, "binary", "b", params.Config.BinaryUnitPrefix, "display the unit prefixes in binary (Mibit/s) instead of decimal (Mbps)")
	cmd.Flags().BoolVar(&params.Verbose, "verbose", params.Verbose, "provide additional information from the logger")

	// Set the function to execute the logic.
	cmd.RunE = cmdRunE(params)

	return cmd
}

// cmdRunE executes the logic of the command line application.
func cmdRunE(params *Parameters) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := cmdValidateE(params)
		if err != nil {
			return err
		}

		if params.Verbose {
			log.Verbose()
		}

		log.Info(
			"token: %s, nodownload: %v, noupload: %v, duration: %d, binary: %v, verbose: %v\n",
			params.APIEndpointToken,
			params.NoDownload,
			params.NoUpload,
			params.Config.Duration,
			params.Config.BinaryUnitPrefix,
			params.Verbose,
		)

		resp, err := getRemoteServerList(params.APIEndpointToken)
		if err != nil {
			return err
		}

		pterm.DefaultBasicText.Printf("Testing from Origin: %s — %s, %s [%s]\n", resp.Client.ISP, resp.Client.Location.City, resp.Client.Location.Country, resp.Client.IP)

		servers, err := getLowestRTTServers(resp.Targets, DefaultTestServerCount, api.ICMPProbe)
		if err != nil {
			return err
		}

		err = runTestSuite(params, servers)
		if err != nil {
			return err
		}

		return nil
	}
}

// cmdValidateE validates the parameters the users passes on the command line.
func cmdValidateE(params *Parameters) error {
	if params.Config.Duration < 3 || params.Config.Duration > 30 {
		return ErrDurationOutOfBounds
	}

	if params.Config.PingCount < 1 || params.Config.PingCount > 5 {
		return ErrPingCountOutOfBounds
	}

	return nil
}

// getRemoteServerList gets a list of servers from a remote URL.
func getRemoteServerList(token string) (*RemoteServerResponse, error) {
	// Dynamically retrieve the endpoint token
	if token == "" {
		log.Warn("no token found in provided params; getting api endpoint token\n")
		t, err := api.GetAPIEndpointToken()
		if err != nil {
			return nil, err
		}

		token = t
	}

	// Query the remote for the JSON list of the nearest servers
	resp, err := http.Get(fmt.Sprintf("%s?token=%s&https=true", api.FastSpeedTestServerURL, token))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		log.Error("GET request made to %s?token=%s&https=true returned status code %d\n", api.FastSpeedTestServerURL, token, resp.StatusCode)
		return nil, ErrUnknownAppToken
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert the remote response into a JSON object.
	var remote RemoteServerResponse
	err = json.Unmarshal(body, &remote)
	if err != nil {
		return nil, err
	}

	return &remote, nil
}

// getLowestRTTServers determines the testing servers by evaluating the lowest round-trip times (RTT).
// The number of servers returned is limited by 'count' and the type of probe is determined by 'pf'.
func getLowestRTTServers(candidates []api.Server, count int, pf api.ProbeFunc) ([]api.Server, error) {
	if len(candidates) == 0 {
		return []api.Server{}, ErrNoCandidatesToRank
	} else if len(candidates) < count {
		log.Warn("number of candidates was less than the count parameter\n")
		count = len(candidates)
	}

	// Get the RTT of each server and store it in our Candidate struct for sorting.
	s := make([]Candidate, 0, len(candidates))
	for i := 0; i < len(candidates); i++ {
		rtt, err := pf(candidates[i], 1)
		if err != nil {
			return []api.Server{}, err
		}

		s = append(s, Candidate{Server: candidates[i], RTT: rtt})
		log.Info("server in %s, %s reported a ping of %s\n", candidates[i].Location.City, candidates[i].Location.Country, (rtt.Round(time.Millisecond)))
	}

	// Sort by RTT (ascending).
	sort.Slice(s, func(i, j int) bool {
		return s[i].RTT < s[j].RTT
	})

	// Hold only the top N lowest RTT servers.
	servers := make([]api.Server, count)
	for i := 0; i < count; i++ {
		servers[i] = s[i].Server
	}

	return servers, nil
}

func runTestSuite(params *Parameters, servers []api.Server) error {
	for i, s := range servers {
		ip, err := s.GetIPv4()
		if err != nil {
			return err
		}

		pterm.DefaultBasicText.Printf("Testing Server: %s, %s [%s]\n", s.Location.City, s.Location.Country, ip)

		runLatencyTest(s, params.Config.PingCount)

		if params.NoDownload {
			pterm.DefaultBasicText.Printf(" %s  Download test is disabled\n", pterm.ThemeDefault.Checkmark.Unchecked)
		} else {
			if err := runDownloadTest(s, params.Config.ConcurrentRequests, params.Config.Duration, DefaultChunkSize, params.Config.BinaryUnitPrefix); err != nil {
				return err
			}
		}

		if params.NoUpload {
			pterm.DefaultBasicText.Printf(" %s  Upload test is disabled\n", pterm.ThemeDefault.Checkmark.Unchecked)
		} else {
			if err := runUploadTest(s, params.Config.ConcurrentRequests, params.Config.Duration, params.Config.BinaryUnitPrefix); err != nil {
				return err
			}
		}

		if i < len(servers)-1 {
			pterm.DefaultBasicText.Printf("\n")
		}

		// Test only the first server for right now. Intention is to test nearest 3 at the same time and record the results.
		if true {
			break
		}
	}

	return nil
}

// runLatencyTest performs the latency test that measures server ping.
func runLatencyTest(server api.Server, pings int) error {
	err := server.Latency(pings)
	if err != nil {
		return err
	}

	return nil
}

// runDownloadTest performs the download speed test that measures the download rate in Mbps.
func runDownloadTest(server api.Server, requests, duration int, chunk int64, binary bool) error {
	err := server.SetChunkSize(chunk)
	if err != nil {
		return fmt.Errorf("failed to append chunk size: %s", err)
	}

	err = server.Download(requests, chunk, time.Duration(duration)*time.Second, binary)
	if err != nil {
		return err
	}

	return nil
}

// runDownloadTest performs the upload speed test that generates a payload to send to the server
// and measures it's upload rate in Mbps.
func runUploadTest(server api.Server, requests, duration int, binary bool) error {
	payload, err := api.GeneratePayload(UploadTestPayloadSize)
	if err != nil {
		return err
	}

	err = server.Upload(requests, time.Duration(duration)*time.Second, payload, binary)
	if err != nil {
		return err
	}

	return nil
}
