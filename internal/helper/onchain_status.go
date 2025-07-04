package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const healthCheckerTimeInterval = 5 * time.Second

type OnchainStatus struct {
	ChainID             string
	BlockHeight         int64
	EarliestBlockHeight int64
}

func GetOnChainStatus(endpoints []string, protocolType string) OnchainStatus {
	results := make(chan OnchainStatus, len(endpoints))
	client := &http.Client{Timeout: healthCheckerTimeInterval}
	var resultStatus OnchainStatus
	var wg sync.WaitGroup

	for _, url := range endpoints {
		// Add the number of endpoints to the WaitGroup counter
		wg.Add(1)

		switch protocolType {
		case "cosmos":
			go func(url string) {
				defer wg.Done()
				getStatusForCosmos(client, url, results)
			}(url)
		case "ethereum":
			go func(url string) {
				defer wg.Done()
				getStatusForEthereum(client, url, results)
			}(url)
		case "gnoland":
			go func(url string) {
				defer wg.Done()
				getStatusForGnoland(client, url, results)
			}(url)
		}
	}

	// Close the results channel once all goroutines are done
	wg.Wait()
	close(results)

	// Listen for results from the channel
	for status := range results {
		if status.ChainID != "" {
			resultStatus = status
			break
		}
	}

	return resultStatus
}

func getStatusForCosmos(client *http.Client, url string, resultChan chan<- OnchainStatus) {
	var checkPath string = "/status"

	resp, err := client.Get((url + checkPath))
	if err != nil {
		resultChan <- OnchainStatus{}
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		resultChan <- OnchainStatus{}
		return
	}

	chainID, blockHeight, earliestBlockHeight, err := CosmosStatusParser(bodyBytes)
	if err != nil {
		resultChan <- OnchainStatus{}
		return
	}

	resultChan <- OnchainStatus{
		ChainID:             chainID,
		BlockHeight:         blockHeight,
		EarliestBlockHeight: earliestBlockHeight,
	}
}

func CosmosStatusParser(resp []byte) (
	/* on-chain id */ string,
	/* on-chain height*/ int64,
	/* node earliest block height */ int64,
	/* unexpected error */ error,
) {
	type StatusResponse struct {
		NodeInfo struct {
			Network string `json:"network"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHeight   string    `json:"latest_block_height"`
			LatestBlockTime     time.Time `json:"latest_block_time"`
			EarliestBlockHeight string    `json:"earliest_block_height"`
			CatchingUp          bool      `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo map[string]any `json:"validator_info"`
	}

	{
		type CosmosV34Status struct {
			JsonRPC string         `json:"jsonrpc"`
			ID      int            `json:"id"`
			Result  StatusResponse `json:"result"`
		}

		var status CosmosV34Status
		if err := json.Unmarshal(resp, &status); err == nil && status.JsonRPC != "" {
			blockHeight, err := strconv.ParseInt(status.Result.SyncInfo.LatestBlockHeight, 10, 64)
			if err != nil {
				return "", 0, 0, err
			}
			ealiestBlockHeight, err := strconv.ParseInt(status.Result.SyncInfo.EarliestBlockHeight, 10, 64)
			if err != nil {
				return "", 0, 0, err
			}
			return status.Result.NodeInfo.Network, blockHeight, ealiestBlockHeight, nil
		}
	}

	{
		type CosmosV37Status StatusResponse

		var status CosmosV37Status
		if err := json.Unmarshal(resp, &status); err == nil {
			blockHeight, err := strconv.ParseInt(status.SyncInfo.LatestBlockHeight, 10, 64)
			if err != nil {
				return "", 0, 0, err
			}
			ealiestBlockHeight, err := strconv.ParseInt(status.SyncInfo.EarliestBlockHeight, 10, 64)
			if err != nil {
				return "", 0, 0, err
			}
			return status.NodeInfo.Network, blockHeight, ealiestBlockHeight, nil
		}
	}

	return "", 0, 0, errors.New("unrecognized response format for cosmos status")
}

// TODO: not implemented
func getStatusForEthereum(client *http.Client, url string, resultChan chan<- OnchainStatus) {
	var checkPayload string = `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	resp, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(checkPayload)))
	if err != nil {
		resultChan <- OnchainStatus{}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		resultChan <- OnchainStatus{}
	} else {
		resultChan <- OnchainStatus{}
	}
}

func getStatusForGnoland(client *http.Client, url string, resultChan chan<- OnchainStatus) {
	var checkPath string = "/status"

	resp, err := client.Get((url + checkPath))
	if err != nil {
		resultChan <- OnchainStatus{}
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		resultChan <- OnchainStatus{}
		return
	}

	type GnolandStatus struct {
		JsonRPC string `json:"jsonrpc"`
		Result  struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"node_info"`
			SyncInfo struct {
				LatestBlockHeight string    `json:"latest_block_height"`
				LatestBlockTime   time.Time `json:"latest_block_time"`
				CatchingUp        bool      `json:"catching_up"`
			} `json:"sync_info"`
		} `json:"result"`
	}

	var status GnolandStatus
	if err := json.Unmarshal(bodyBytes, &status); err == nil && status.JsonRPC != "" {
		blockHeight, err := strconv.ParseInt(status.Result.SyncInfo.LatestBlockHeight, 10, 64)
		if err != nil {
			resultChan <- OnchainStatus{}
			return
		}

		resultChan <- OnchainStatus{
			ChainID:             status.Result.NodeInfo.Network,
			BlockHeight:         blockHeight,
			EarliestBlockHeight: blockHeight,
		}
		return
	}

	resultChan <- OnchainStatus{}
}
