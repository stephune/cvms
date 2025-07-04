package indexer

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/config"

	// axelar specific
	aavindexer "github.com/cosmostation/cvms/internal/packages/axelar/amplifier-verifier/indexer"

	// babylon specific
	btclcindexer "github.com/cosmostation/cvms/internal/packages/babylon/btc-lightclient/indexer"
	bcindexer "github.com/cosmostation/cvms/internal/packages/babylon/checkpoint/indexer"
	bcsindexer "github.com/cosmostation/cvms/internal/packages/babylon/covenant-committee/indexer"
	bfpindexer "github.com/cosmostation/cvms/internal/packages/babylon/finality-provider/indexer"

	// gnoland specific
	gvtindexer "github.com/cosmostation/cvms/internal/packages/gnoland/voteindexer"

	// cosmos native
	bdaindexer "github.com/cosmostation/cvms/internal/packages/block-data-analytics/indexer"
	veindexer "github.com/cosmostation/cvms/internal/packages/consensus/veindexer/indexer"
	voteindexer "github.com/cosmostation/cvms/internal/packages/consensus/voteindexer/indexer"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

func selectPackage(
	m common.Mode, f promauto.Factory, l *logrus.Logger,
	idb *common.IndexerDB, mainnet bool, chainID, chainName, pkg, protocolType string,
	isConsumer bool,
	cc config.ChainConfig, monikers []string,
) error {
	// Add validation logic on each provided URL
	validAPIs := make([]string, 0)
	validRPCs := make([]string, 0)

	for _, node := range cc.Nodes {
		if helper.ValidateURL(node.RPC) {
			validRPCs = append(validRPCs, node.RPC)
		}

		if helper.ValidateURL(node.API) {
			validAPIs = append(validAPIs, node.API)
		}
	}

	providerRPCs := make([]string, 0)
	providerAPIs := make([]string, 0)
	if isConsumer {
		for _, node := range cc.ProviderNodes {
			if helper.ValidateURL(node.RPC) {
				providerRPCs = append(providerRPCs, node.RPC)
			}
			if helper.ValidateURL(node.API) {
				providerAPIs = append(providerAPIs, node.API)
			}
		}
	}

	switch {
	case pkg == "voteindexer":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		if isConsumer {
			providerEndpoints := common.Endpoints{RPCs: providerRPCs, CheckRPC: true, APIs: providerAPIs, CheckAPI: true}
			p.SetAddtionalEndpoints(providerEndpoints)
			p.SetConsumer()
		}
		voteindexer, err := voteindexer.NewVoteIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return voteindexer.Start()
	case pkg == "veindexer":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		if isConsumer {
			providerEndpoints := common.Endpoints{RPCs: providerRPCs, CheckRPC: true, APIs: providerAPIs, CheckAPI: true}
			p.SetAddtionalEndpoints(providerEndpoints)
			p.SetConsumer()
		}
		veindexer, err := veindexer.NewVEIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return veindexer.Start()
	case pkg == "babylon-checkpoint":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		if isConsumer {
			providerEndpoints := common.Endpoints{RPCs: providerRPCs, CheckRPC: true, APIs: providerAPIs, CheckAPI: true}
			p.SetAddtionalEndpoints(providerEndpoints)
			p.SetConsumer()
		}
		bcindexer, err := bcindexer.NewCheckpointIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return bcindexer.Start()
	case pkg == "babylon-finality-provider-indexer":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		if isConsumer {
			providerEndpoints := common.Endpoints{RPCs: providerRPCs, CheckRPC: true, APIs: providerAPIs, CheckAPI: true}
			p.SetAddtionalEndpoints(providerEndpoints)
			p.SetConsumer()
		}
		fpindexer, err := bfpindexer.NewFinalityProviderIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return fpindexer.Start()
	case pkg == "babylon-btc-lightclient":
		if m == common.VALIDATOR {
			l.Infof("%s don't need to enable this package. CVMS will ignore %s package", m, pkg)
			return nil
		}
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		btclcindexer, err := btclcindexer.NewBTCLightClientIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return btclcindexer.Start()
	case pkg == "babylon-covenant-committee":
		if m == common.VALIDATOR {
			l.Infof("%s don't need to enable this package. CVMS will ignore %s package", m, pkg)
			return nil
		}
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		csindexer, err := bcsindexer.NewCovenantSignatureIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return csindexer.Start()
	case pkg == "axelar-amplifier-verifier":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		aavindexer, err := aavindexer.NewAxelarAmplifierVerifierIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return aavindexer.Start()
	case pkg == "block-data-analytics":
		if m == common.VALIDATOR {
			l.Infof("%s don't need to enable this package. CVMS will ignore %s package", m, pkg)
			return nil
		}
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		bdaindexer, err := bdaindexer.NewBDAIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return bdaindexer.Start()
	case pkg == "gnoland-voteindexer":
		if m == common.VALIDATOR {
			l.Infof("%s don't need to enable this package. CVMS will ignore %s package", m, pkg)
			return nil
		}
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		gvtindexer, err := gvtindexer.NewVoteIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return gvtindexer.Start()
	}

	return common.ErrUnSupportedPackage
}
