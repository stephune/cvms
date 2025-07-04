package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	indexermodel "github.com/cosmostation/cvms/internal/common/indexer/model"
	"github.com/cosmostation/cvms/internal/common/parser"
	"github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/pkg/errors"
)

// query cosmos validators on each a new block
func GetGnolandBFTValidators(c common.CommonClient, height ...int64) ([]types.GnolandBFTValidator, error) {
	// init context
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	// create requester
	requester := c.RPCClient.R().SetContext(ctx)

	totalValidators := make([]types.GnolandBFTValidator, 0)
	var queryPath string

	if len(height) > 0 {
		queryPath = types.GnolandBFTValidatorQueryPathWithHeight(height[0])
	} else {
		queryPath = types.GnolandBFTValidatorQueryPath()
	}

	resp, err := requester.Get(queryPath)
	if err != nil {
		return nil, errors.Errorf("rpc call is failed from %s: %s", resp.Request.URL, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("stanage status code from %s: [%d]", resp.Request.URL, resp.StatusCode())
	}

	validators, err := parser.GnolandBFTValidatorParser(resp.Body())
	if err != nil {
		return nil, errors.Wrapf(err, "got data, but failed to parse the data")
	}

	totalValidators = append(totalValidators, validators...)

	c.Debugf("found all gnoland bft validators %d, who matched each staking validator", len(totalValidators))
	return totalValidators, nil
}

//nolint:dupl // gnoland block is similar logic with cosmos chains, but in the future, the logic will be changed via very different way
func GetGnolandBlock(c common.CommonClient, height int64) (
	/* block height */ int64,
	/* block timestamp */ time.Time,
	/* block proposer addrss */ string,
	/* block txs */ []types.Tx,
	/* last commit block height*/ int64,
	/* block validators signatures */ []types.GnolandSignature,
	error,
) {
	// init context
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	// create requester
	requester := c.RPCClient.R().SetContext(ctx)

	resp, err := requester.Get(types.GnolandBlockQueryPath(height))
	if err != nil {
		return 0, time.Time{}, "", nil, 0, nil, errors.Errorf("rpc call is failed from %s: %s", resp.Request.URL, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return 0, time.Time{}, "", nil, 0, nil, errors.Errorf("stanage status code from %s: [%d]", resp.Request.URL, resp.StatusCode())
	}

	blockHeight, blockTimeStamp, blockProposerAddress, blockTxs, lastCommitBlockHeight, blockSignatures, err := parser.GnolandBlockParser(resp.Body())
	if err != nil {
		return 0, time.Time{}, "", nil, 0, nil, errors.Wrapf(err, "got data, but failed to parse the data")
	}

	return blockHeight, blockTimeStamp, blockProposerAddress, blockTxs, lastCommitBlockHeight, blockSignatures, nil
}

// https://gno.land/r/sys/validators/v2$help#func-GetValidators
func GetGnolandSysValidators(c common.CommonClient) ([]types.GnolandValidatorInfo, error) {
	// init context
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	// create requester
	requester := c.RPCClient.R().SetContext(ctx)

	// get on-chain validators in staking module
	resp, err := requester.Get(types.GnolandSysValidatorQueryPath())
	if err != nil {
		return nil, errors.Wrap(err, "failed in api")
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("got %d code from %s", resp.StatusCode(), resp.Request.URL)
	}

	decoded, err := parser.GnolandABCIParser(resp.Body())
	if err != nil {
		return nil, errors.Wrap(err, "failed in api")
	}

	gnolandValidators, err := parser.GnolandValidatorsParser(decoded)
	if err != nil {
		return nil, errors.Wrap(err, "failed in api")
	}

	// init
	gnolandValidatorInfos := make([]types.GnolandValidatorInfo, 0)

	// collect validator info in go-routine
	ch := make(chan helper.Result)
	var wg sync.WaitGroup

	for _, gnoVal := range gnolandValidators {
		wg.Add(1)
		val := gnoVal

		go func(ch chan helper.Result) {
			defer helper.HandleOutOfNilResponse(c.Entry)
			defer wg.Done()

			validatorInfo, err := GetGnolandValidatorInfo(c, val)
			if err != nil {
				c.Errorf("failed to get %s info, %s", val.Address, err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}
			ch <- helper.Result{
				Item:    validatorInfo,
				Success: true,
			}
		}(ch)
		time.Sleep(10 * time.Millisecond)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	errorCount := 0
	for r := range ch {
		if r.Success {
			item := r.Item.(types.GnolandValidatorInfo)
			gnolandValidatorInfos = append(gnolandValidatorInfos, item)
			continue
		}
		errorCount++
	}

	if errorCount > 0 {
		return nil, errors.New("failed to get gnoland validator infos in go-routine")
	}

	// logging total validators count
	c.Debugf("total gnoland sys validators: %d", len(gnolandValidatorInfos))
	return gnolandValidatorInfos, nil
}

func GetGnolandValidatorInfo(c common.CommonClient, gnolandVal types.GnolandValidator) (types.GnolandValidatorInfo, error) {
	// init context
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	// create requester
	requester := c.RPCClient.R().SetContext(ctx)

	// get on-chain validators in staking module
	resp, err := requester.Get(types.GnolandValidatorInfoQueryPath(gnolandVal.Address))
	if err != nil {
		return types.GnolandValidatorInfo{}, errors.Wrap(err, "failed in api")
	}
	if resp.StatusCode() != http.StatusOK {
		return types.GnolandValidatorInfo{}, errors.Errorf("got %d code from %s", resp.StatusCode(), resp.Request.URL)
	}
	// fmt.Printf("resp: %s", resp.Body())

	decoded, err := parser.GnolandABCIParser(resp.Body())
	if err != nil {
		return types.GnolandValidatorInfo{}, errors.Wrap(err, "failed in api")
	}

	// TODO: temporary handling
	if decoded == "" {
		return types.GnolandValidatorInfo{
			Name:        gnolandVal.Address,
			Description: "",
			PubKey:      gnolandVal.PubKey,
			Address:     gnolandVal.Address,
			KeepRunning: true,
		}, nil
	}
	gnolandValiatorInfo, err := parser.GnolandValidatorInfoParser(decoded)
	if err != nil {
		return types.GnolandValidatorInfo{}, errors.Wrap(err, "failed in api")
	}

	return gnolandValiatorInfo, nil
}

// NOTE: gnoland validator address and tendermint address are same
func MakeGnolandValidatorInfoList(app common.CommonApp, chainInfoID int64, newValidatorAddressMap map[string]bool) ([]indexermodel.ValidatorInfo, error) {
	// 1. init
	gnolandValidatorInfoMap := make(map[string]types.StakingValidatorMetaInfo)

	// 2. get gnovm sys validator info
	gnolandSysValidators, err := GetGnolandSysValidators(app.CommonClient)
	if err != nil {
		return nil, errors.Cause(err)
	}

	// 3. provide info into inited map
	for _, gnovali := range gnolandSysValidators {
		gnolandValidatorInfoMap[gnovali.Address] = types.StakingValidatorMetaInfo{
			Moniker:         gnovali.Name,
			OperatorAddress: gnovali.Address,
		}
	}

	// make new validator info list to insert into indexer db
	newValidatorInfoList := make([]indexermodel.ValidatorInfo, 0)
	for newValidatorAddress := range newValidatorAddressMap {
		newValidatorInfoList = append(
			newValidatorInfoList,
			indexermodel.ValidatorInfo{
				ChainInfoID:     chainInfoID,
				HexAddress:      newValidatorAddress,
				OperatorAddress: gnolandValidatorInfoMap[newValidatorAddress].OperatorAddress,
				Moniker:         gnolandValidatorInfoMap[newValidatorAddress].Moniker,
			})
	}
	return newValidatorInfoList, nil
}
