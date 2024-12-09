package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/primlock/zoomies/api"
	"github.com/spf13/cobra"
)

type RemoteServerResponse struct {
	Targets []api.Server `json:"targets"`
}

type Options struct {
	APIEndpointToken string
	URLCount         int
	HTTPSEnabled     bool
	RunDownloadTest  bool
	RunUploadTest    bool
	Config           TestConfig

	// Json Server File
	// list the servers being tested
	// test with only the fastest server
}

type TestConfig struct {
	TimeoutSeconds   int
	DownloadRequests int
	UploadRequests   int
}

// Possibly move this into a local scope: RunE
var opts = &Options{
	URLCount:        5,
	HTTPSEnabled:    true,
	RunDownloadTest: true,
	RunUploadTest:   true,
	Config: TestConfig{
		TimeoutSeconds:   30,
		DownloadRequests: 5, // How many GET requests to make to the server
		UploadRequests:   5, // How many POST requests to make to the server
	},
}

var (
	ErrURLCountOutOfBounds = errors.New("count must be in the range 1-5 inclusive")
	ErrUnknownAppToken     = errors.New("invalid token passed as a parameter")
)

const (
	UploadTestPayloadSize = 25 * 1024 * 1024 // 25 MB
)

var cmd = &cobra.Command{
	Use:   "zoomies",
	Short: "zoomies is a network speed measurement tool",
	RunE: func(cmd *cobra.Command, args []string) error {

		if opts.URLCount < 1 || opts.URLCount > 5 {
			return ErrURLCountOutOfBounds
		}

		// Gather the required server information
		servers, err := getRemoteServerList(opts.APIEndpointToken, opts.HTTPSEnabled, opts.URLCount)
		if err != nil {
			return err
		}

		// Keep for debug
		// cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// 	fmt.Printf("Flag '%s': %s\n", flag.Name, flag.Value)
		// })

		// Run the tests
		err = runTestSuite(servers)
		if err != nil {
			return err
		}

		return nil
	},
}

// Get a list of servers from a remote URL.
func getRemoteServerList(token string, https bool, URLCount int) ([]api.Server, error) {
	// Dynamically retrieve the endpoint token
	if token == "" {
		t, err := api.GetAPIEndpointToken()
		if err != nil {
			return nil, err
		}

		token = t
	}

	// Query the remote for the JSON list of the nearest servers
	resp, err := http.Get(fmt.Sprintf("%s?token=%s&https=%v&urlCount=%d", api.FastSpeedTestServerURL, token, https, URLCount))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrUnknownAppToken
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert the remote response into a JSON object
	var remote RemoteServerResponse
	err = json.Unmarshal(body, &remote)
	if err != nil {
		return nil, err
	}

	return remote.Targets, nil
}

// TODO: Get a list of servers from a local file: func getLocalServerList() ([]api.Server, error)

func runTestSuite(servers []api.Server) error {
	if opts.RunDownloadTest {
		if err := runDownloadTest(servers); err != nil {
			return err
		}
	} else {
		fmt.Println("Download test is disabled")
	}

	if opts.RunUploadTest {
		if err := runUploadTest(servers); err != nil {
			return err
		}
	} else {
		fmt.Println("Upload test is disabled")
	}

	return nil
}

func runDownloadTest(servers []api.Server) error {
	for _, s := range servers {
		ip, err := s.GetIPv4()
		if err != nil {
			return err
		}

		// TODO: Turn this into a debug output
		fmt.Printf("Download testing server: %s", ip)

		res, err := s.Download(opts.Config.TimeoutSeconds, opts.Config.DownloadRequests)
		if err != nil {
			return err
		}

		// Display the result of the test
		api.DisplayTestResults(res)
	}

	return nil
}

func runUploadTest(servers []api.Server) error {
	payload, err := api.GeneratePayload(UploadTestPayloadSize)
	if err != nil {
		return err
	}

	for _, s := range servers {
		ip, err := s.GetIPv4()
		if err != nil {
			return err
		}

		// TODO: Turn this into a debug output
		fmt.Printf("Upload testing server: %s", ip)

		res, err := s.Upload(opts.Config.TimeoutSeconds, opts.Config.UploadRequests, payload)
		if err != nil {
			return err
		}

		// Display the results of the test
		api.DisplayTestResults(res)
	}

	return nil
}

func init() {
	cmd.Flags().StringVarP(&opts.APIEndpointToken, "token", "t", "", "user provided api endpoint access token")
	cmd.Flags().BoolVar(&opts.HTTPSEnabled, "https", opts.HTTPSEnabled, "enable https")
	cmd.Flags().IntVarP(&opts.URLCount, "count", "c", opts.URLCount, "number of URLs to test (1-5)")
	cmd.Flags().BoolVar(&opts.RunDownloadTest, "download", opts.RunDownloadTest, "perform the download test")
	cmd.Flags().BoolVar(&opts.RunUploadTest, "upload", opts.RunUploadTest, "perform the upload test")
	// TODO: implement flags for TestConfig
}

func Execute() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
