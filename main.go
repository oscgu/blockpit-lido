package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Event struct {
	APR                    string `json:"apr,omitempty"`
	Block                  string `json:"block,omitempty"`
	BlockTime              string `json:"blockTime,omitempty"`
	ID                     string `json:"id,omitempty"`
	LogIndex               string `json:"logIndex,omitempty"`
	TotalPooledEtherAfter  string `json:"totalPooledEtherAfter,omitempty"`
	TotalPooledEtherBefore string `json:"totalPooledEtherBefore,omitempty"`
	TotalSharesAfter       string `json:"totalSharesafter,omitempty"`
	TotalSharesBefore      string `json:"totalSharesBefore,omitempty"`
	EpochDays              string `json:"epochDays,omitempty"`
	EpochFullDays          string `json:"epochFullDays,omitempty"`
	Type                   string `json:"type,omitempty"`
	ReportShares           string `json:"reportShares,omitempty"`
	Balance                string `json:"balance,omitempty"`
	Rewards                string `json:"rewards,omitempty"`
	Change                 string `json:"change,omitempty"`
	CurrencyChange         string `json:"currencyChange,omitempty"`
}

type Totals struct {
	EthRewards      string `json:"ethRewards,omitempty"`
	CurrencyRewards string `json:"currencyRewards,omitempty"`
}

type StEthCurrencyPrice struct {
	Eth float64 `json:"eth,omitempty"`
	Usd float64 `json:"usd,omitempty"`
}

type LidoAPIResponse struct {
	Totals             Totals             `json:"totals,omitempty"`
	AverageAPR         string             `json:"averageApr,omitempty"`
	Events             []Event            `json:"events,omitempty"`
	StEthCurrencyPrice StEthCurrencyPrice `json:"stEthCurrencyPrice,omitempty"`
	EthToStEthRatio    float64            `json:"ethToStEthRatio,omitempty"`
	TotalItems         uint64             `json:"totalItems,omitempty"`
}

var headers = []string{
	"Date (UTC)",
	"Integration Name",
	"Label",
	"Outgoing Asset",
	"Outgoing Amount",
	"Incoming Asset",
	"Incoming Amount",
	"Fee Asset (optional)",
	"Fee Amount (optional)",
	"Comment (optional)",
	"Trx. ID (optional)",
}

func main() {
	address := flag.String(
		"address",
		"",
		"address of the wallet you wish to generate a blockpit compatible csv for",
	)
	currency := flag.String("currency", "USD", "")
	archiveRate := flag.String("archive-rate", "false", "")
	onlyRewards := flag.String("only-rewards", "true", "")
	lidoApiUrl := flag.String("lido-api-url", "https://stake.lido.fi/api/rewards", "")
	year := flag.String("year", "2023", "The tax year you want to get a csv for")
	out := flag.String("out", "", "--out <filename> -- defaults to '<year>-report.csv'")
	integration := flag.String("integration", "Lido stETH", "")
	label := flag.String("label", "Staking", "")
	asset := flag.String("asset", "stETH", "")

	flag.Parse()

	evHist, err := getEventHist(
		context.Background(),
		*lidoApiUrl,
		*address,
		*currency,
		*archiveRate,
		*onlyRewards,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := writeCSV(evHist.Events, *year, *out, *integration, *asset, *label); err != nil {
		log.Fatal(err)
	}
	log.Println("done")
}

func getEventHist(
	ctx context.Context,
	lidoApiUrl, address, currency, archiveRate, onlyRewards string,
) (*LidoAPIResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lidoApiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	q := req.URL.Query()
	q.Add("address", address)
	q.Add("currency", currency)
	q.Add("archiveRate", archiveRate)
	q.Add("onlyRewards", onlyRewards)
	req.URL.RawQuery = q.Encode()

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var lidoResponse LidoAPIResponse
	if err := json.Unmarshal(b, &lidoResponse); err != nil {
		return nil, fmt.Errorf("unmarshalling response: %w", err)
	}

	return &lidoResponse, nil
}

func writeCSV(events []Event, year, out, integration, asset, label string) error {
	if out == "" {
		out = year + "-report.csv"
	}
	log.Printf("writing data to %s", out)

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("opening csv file: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("writing csv headers: %w", err)
	}

	for _, e := range events {
		t, err := strconv.ParseInt(e.BlockTime, 10, 64)
		if err != nil {
			return fmt.Errorf("converting %s to int: %w", e.BlockTime, err)
		}

		evDate := time.Unix(t, 0).UTC()
		if year != evDate.Format("2006") {
			continue
		}

		evDateStr := evDate.Format("02-01-2006 15:04:05")
		bigReward, ok := new(big.Int).SetString(e.Rewards, 10)
		if !ok {
			log.Println("could not set rewards")
		}

		bigRewardFloat := new(big.Float).SetInt(bigReward)
		decimalsInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(18)), nil)
		decimalsFloat := new(big.Float).SetInt(decimalsInt)

		reward := new(big.Float).Quo(bigRewardFloat, decimalsFloat).Text(byte('f'), 18)

		if err := writer.Write([]string{
			evDateStr,
			integration,
			label,
			"",
			"",
			asset,
			reward,
			"",
			"",
			"Lido stETH staking reward",
			"",
		}); err != nil {
			return fmt.Errorf("writing row: %w", err)
		}
	}

	return nil
}
