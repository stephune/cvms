package types

// NOTE: SupportedChainTypes = []string{"cosmos", "ethereum", "aptos", "sui", "avalanche", "celestia", "polkadot", "aleo"}
var (
	SupportedChainTypes = []string{"cosmos", "ethereum", "celestia", "gnoland"}
)

type CommonBlock struct {
	LastBlockHeight    float64
	LastBlockTimeStamp float64
}
