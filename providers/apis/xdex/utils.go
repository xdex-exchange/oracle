package xdex

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/skip-mev/slinky/oracle/config"
)

const (
	// Name is the name of the Xdex provider.
	Name = "xdex_api"

	// URL is the base URL of the Xdex API. This includes the base and quote
	// currency pairs that need to be inserted into the URL. This URL should be utilized
	// by Non-US users.
	URL = "https://abvote.orderstory.xyz/api/v1/oracle/price?symbols=%s%s%s"

	Quotation    = "%22"
	Separator    = ","
	LeftBracket  = "%5B"
	RightBracket = "%5D"
)

var DefaultNonUSAPIConfig = config.APIConfig{
	Name:             Name,
	Atomic:           true,
	Enabled:          true,
	Timeout:          3000 * time.Millisecond,
	Interval:         750 * time.Millisecond,
	ReconnectTimeout: 2000 * time.Millisecond,
	MaxQueries:       1,
	Endpoints:        []config.Endpoint{{URL: URL}},
}

type (
	Response []Data

	Data struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
)

func Decode(resp *http.Response) (Response, error) {
	var result Response
	err := json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}
