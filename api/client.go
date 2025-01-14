package api

type Client struct {
	IP       string `json:"ip"`
	ASN      string `json:"asn"`
	ISP      string `json:"isp"`
	Location struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"location"`
}
