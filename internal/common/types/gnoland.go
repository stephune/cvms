package types

import (
	"encoding/base64"
	"fmt"
	"time"
)

type GnolandABCIResponse struct {
	Result struct {
		Response struct {
			ResponseBase struct {
				Data string `json:"data"`
			} `json:"ResponseBase"`
		} `json:"response"`
	} `json:"result"`
}

var GnolandBFTValidatorQueryPathWithHeight = func(height int64) string {
	return fmt.Sprintf("/validators?height=%d", height)
}

var GnolandBFTValidatorQueryPath = func() string {
	return "/validators"
}

type GnolandBFTValidatorResponse struct {
	Result GnolandBFTValidators `json:"result"`
}

// response of cosmos-sdk based chain validators
type GnolandBFTValidators struct {
	BlockHeight string                `json:"block_height"`
	Validators  []GnolandBFTValidator `json:"validators"`
}

type GnolandBFTValidator struct {
	Address string `json:"address"`
	Pubkey  struct {
		Type  string `json:"@type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	VotingPower      string `json:"voting_power"`
	ProposerPriority string `json:"proposer_priority"`
}

type GnolandBlockSummary struct {
	BlockHeight          int64
	BlockTimeStamp       time.Time
	BlockProposerAddress string
	// Txs                   []Tx
	LastCommitBlockHeight int64
	BlockSignatures       []GnolandSignature
	GnolandBFTValidators  []GnolandBFTValidator
}

// response type for v34 cosmos block
type GnolandBlockResponse struct {
	JsonRPC string       `json:"jsonrpc"`
	Result  GnolandBlock `json:"result"`
}

// query path for cosmos block by height
var GnolandBlockQueryPath = func(height int64) string {
	return fmt.Sprintf("/block?height=%d", height)
}

// response of cosmos-sdk based chain block
type GnolandBlock struct {
	BlockMeta struct {
		BlockID interface{} `json:"-"`
		Header  interface{} `json:"-"`
	} `json:"block_meta"`
	Block struct {
		Header struct {
			ChainID         string    `json:"chain_id"`
			Height          string    `json:"height"`
			Time            time.Time `json:"time"`
			ProposerAddress string    `json:"proposer_address"`
		} `json:"header"`
		Data struct {
			Txs []Tx `json:"txs"`
		} `json:"data"`
		Evidence   interface{} `json:"-"`
		LastCommit struct {
			PreCommits []GnolandPrecommit `json:"precommits"`
		} `json:"last_commit"`
	} `json:"block"`
}

type GnolandPrecommit struct {
	Type             int
	Height           string      `json:"height"`
	Round            string      `json:"round"`
	BlockID          interface{} `json:"-"`
	Timestamp        time.Time   `json:"timestamp"`
	ValidatorAddress string      `json:"validator_address"`
	ValidatorIndex   string      `json:"validator_index"`
	Signature        string      `json:"signature"`
}

type GnolandSignature struct {
	Timestamp        time.Time
	ValidatorAddress string
	ValidatorIndex   string
	Signature        string
}

var GnolandSysValidatorQueryPath = func() string {
	data := "gno.land/r/sys/validators/v2.GetValidators()"
	base64EncodedData := base64.StdEncoding.EncodeToString([]byte(data))
	return fmt.Sprintf(`/abci_query?path="vm/qeval"&data="%s"`, base64EncodedData)
}

type GnolandSysValidatorsQueryResponse struct {
	Validators []CosmosStakingValidator `json:"validators"`
	Pagination struct {
		// NextKey interface{} `json:"-"`
		Total string `json:"total"`
	} `json:"pagination"`
}

type GnolandSysValidator struct {
	OperatorAddress string          `json:"operator_address"`
	ConsensusPubkey ConsensusPubkey `json:"consensus_pubkey"`
	Description     struct {
		Moniker string `json:"moniker"`
	} `json:"description"`
	Tokens string `json:"tokens"`
	Status string `json:"status"`
}

// Validator defines the structure to hold parsed data for a single validator.
// It corresponds to the gno.land/p/sys/validators.Validator struct.
type GnolandValidator struct {
	Address     string // bech32 address of the validator.
	PubKey      string // bech32 representation of the public key.
	VotingPower uint64 // The voting power of the validator.
}

var GnolandValidatorInfoQueryPath = func(validatorAddress string) string {
	data := fmt.Sprintf(`gno.land/r/gnoland/valopers.GetByAddr("%s")`, validatorAddress)
	base64EncodedData := base64.StdEncoding.EncodeToString([]byte(data))
	return fmt.Sprintf(`/abci_query?path="vm/qeval"&data="%s"`, base64EncodedData)
}

// Valoper defines the structure to hold the parsed data for a gno.land Valoper.
// It is designed to match the fields in the provided input string.
type GnolandValidatorInfo struct {
	Name        string // e.g., "devx-val-2"
	Description string // e.g., "NT DevX Val2"
	Address     string // The .uverse.address of the valoper
	PubKey      string // The public key string
	KeepRunning bool   // The boolean status
	// AuthorizableRaw string // The raw string representation of the authorizable pointer
}
