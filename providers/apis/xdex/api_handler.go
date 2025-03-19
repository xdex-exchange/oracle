package xdex

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	providertypes "github.com/skip-mev/slinky/providers/types"

	"github.com/skip-mev/slinky/pkg/math"

	"github.com/skip-mev/slinky/oracle/config"
	"github.com/skip-mev/slinky/oracle/types"
)

var _ types.PriceAPIDataHandler = (*APIHandler)(nil)

type APIHandler struct {
	api config.APIConfig
	cache types.ProviderTickers
}

func NewAPIHandler(
	api config.APIConfig,
) (types.PriceAPIDataHandler, error) {
	if api.Name != Name {
		return nil, fmt.Errorf("expected api config name %s, got %s", Name, api.Name)
	}

	if !api.Enabled {
		return nil, fmt.Errorf("api config for %s is not enabled", Name)
	}

	if err := api.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid api config for %s: %w", Name, err)
	}

	return &APIHandler{
		api:   api,
		cache: types.NewProviderTickers(),
	}, nil
}

func (h *APIHandler) CreateURL(
	tickers []types.ProviderTicker,
) (string, error) {
	var tickerStrings string
	for _, ticker := range tickers {
		tickerStrings += fmt.Sprintf("%s%s%s%s", Quotation, ticker.GetOffChainTicker(), Quotation, Separator)
		h.cache.Add(ticker)
	}

	if len(tickerStrings) == 0 {
		return "", fmt.Errorf("empty url created. invalid or no ticker were provided")
	}

	return fmt.Sprintf(
		h.api.Endpoints[0].URL,
		LeftBracket,
		strings.TrimSuffix(tickerStrings, Separator),
		RightBracket,
	), nil
}

func (h *APIHandler) ParseResponse(
	tickers []types.ProviderTicker,
	resp *http.Response,
) types.PriceResponse {
	result, err := Decode(resp)
	if err != nil {
		return types.NewPriceResponseWithErr(
			tickers,
			providertypes.NewErrorWithCode(err, providertypes.ErrorFailedToDecode),
		)
	}

	var (
		resolved   = make(types.ResolvedPrices)
		unresolved = make(types.UnResolvedPrices)
	)

	for _, data := range result {
		// Filter out the responses that are not expected.
		ticker, ok := h.cache.FromOffChainTicker(data.Symbol)
		if !ok {
			continue
		}

		price, err := math.Float64StringToBigFloat(data.Price)
		if err != nil {
			wErr := fmt.Errorf("failed to convert price %s to big.Float: %w", data.Price, err)
			unresolved[ticker] = providertypes.UnresolvedResult{
				ErrorWithCode: providertypes.NewErrorWithCode(wErr, providertypes.ErrorFailedToParsePrice),
			}
			continue
		}

		resolved[ticker] = types.NewPriceResult(price, time.Now().UTC())
	}

	// Add currency pairs that received no response to the unresolved map.
	for _, ticker := range tickers {
		_, resolvedOk := resolved[ticker]
		_, unresolvedOk := unresolved[ticker]

		if !resolvedOk && !unresolvedOk {
			unresolved[ticker] = providertypes.UnresolvedResult{
				ErrorWithCode: providertypes.NewErrorWithCode(fmt.Errorf("no response"), providertypes.ErrorNoResponse),
			}
		}
	}

	return types.NewPriceResponse(resolved, unresolved)
}
