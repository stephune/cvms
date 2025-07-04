package api

import (
	"fmt"
	"os"
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/parser"
	"github.com/cosmostation/cvms/internal/common/types"
	"github.com/stretchr/testify/assert"
)

func TestGetGnolandBFTValidators(t *testing.T) {
	rpcEndpoint := os.Getenv("TEST_BLOCK_GNOLAND_HOST_ADDRESS")
	commonApp := common.NewCommonApp(p)
	commonApp.SetRPCEndPoint(rpcEndpoint)

	lh, _, err := GetStatus(commonApp.CommonClient)
	assert.NoError(t, err)

	bftVals, err := GetGnolandBFTValidators(commonApp.CommonClient, lh)
	assert.NoError(t, err)

	for _, bftVal := range bftVals {
		t.Logf("bft val address: %v", bftVal.Address)
		t.Logf("bft val pubkey type: %v", bftVal.Pubkey.Type)
		t.Logf("bft val pubkey value: %v", bftVal.Pubkey.Value)
		break
	}
}

func TestGetGnolandBlock(t *testing.T) {
	rpcEndpoint := os.Getenv("TEST_BLOCK_GNOLAND_HOST_ADDRESS")
	commonApp := common.NewCommonApp(p)
	commonApp.SetRPCEndPoint(rpcEndpoint)

	lh, _, err := GetStatus(commonApp.CommonClient)
	assert.NoError(t, err)

	blockHeight, _, blockProposerAddress, _, lastCommitBlockHeight, gnolandBlockSigs, err := GetGnolandBlock(commonApp.CommonClient, lh)
	assert.NoError(t, err)
	assert.Equal(t, lh, blockHeight)
	assert.NotEmpty(t, blockProposerAddress)
	assert.Equal(t, lh-1, lastCommitBlockHeight)

	// If parsing is successful, print the results.
	fmt.Printf("✅ Successfully parsed %d validators.\n\n", len(gnolandBlockSigs))
	for i, sig := range gnolandBlockSigs {
		fmt.Printf("--- Validator %d ---\n", i+1)
		fmt.Printf("Address:     %s\n", sig.ValidatorAddress)
		fmt.Printf("Index:     %s\n", sig.ValidatorIndex)
		fmt.Printf("Sig:      %s\n", sig.Signature)
	}
}

func TestGetGnolandSysValidators(t *testing.T) {
	rpcEndpoint := os.Getenv("TEST_BLOCK_GNOLAND_HOST_ADDRESS")
	commonApp := common.NewCommonApp(p)
	commonApp.SetRPCEndPoint(rpcEndpoint)

	gnolandSysValidators, err := GetGnolandSysValidators(commonApp.CommonClient)
	assert.NoError(t, err)

	// If parsing is successful, print the results.
	fmt.Printf("✅ Successfully parsed %d validators.\n\n", len(gnolandSysValidators))
	for i, v := range gnolandSysValidators {
		fmt.Printf("--- Validator %d ---\n", i+1)
		fmt.Printf("Moniker:     %s\n", v.Name)
		fmt.Printf("Address:     %s\n", v.Address)
		fmt.Printf("PubKey:      %s\n", v.PubKey)
	}
}

func TestGnolandValidatorsParser(t *testing.T) {
	// The raw string data to be parsed.
	input := `(slice[(struct{("g12yvv8pl5s20suxyd30g7ychqenamtlhctgfu90" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zplgfkp8609ghdh20w6newh40f9tz7ussw2zylq23ca0tjda3csztm242ft" string),(1 uint64)} gno.land/p/sys/validators.Validator),(struct{("g13762rd7y8s7jcc6uc4lyxv269hguchhpyzaamt" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq5ndww8w6qrxgfdeastcx2lsuuk5r8w9jckkgevylq6duw59d54n935fq2" string),(1 uint64)} gno.land/p/sys/validators.Validator),(struct{("g14cppfre9hsvu6p4scttuyu7mj082lfwxl7hvz9" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqe87d7lc0c4l4yaa8a94fucfre8882n8556l9z5220zjaaaqj7k5cl5sud" string),(1 uint64)} gno.land/p/sys/validators.Validator),(struct{("g1927k3s7q9ujla04r5zy7q5m3gl84wsrart6663" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq762adl0tcvdn54d6nzqf68r9wrywn7zj87v92mk3qpr436mevpvc63wsz" string),(1 uint64)} gno.land/p/sys/validators.Validator),(struct{("g1mxguhd5zacar64txhfm0v7hhtph5wur5hx86vs" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqma223maxmnw4f42kfqppvgyn8dr8wu7mhtdm6lcq64303a3vlln8xdmms" string),(1 uint64)} gno.land/p/sys/validators.Validator),(struct{("g1p3lyk676gludkk6hqceem58c6xgnpsld45s4v9" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpf65yj5xh8y9qux89skvve77w7hytfcfey92zlvx56ruugqvk9eepk73fg" string),(1 uint64)} gno.land/p/sys/validators.Validator),(struct{("g1t9ctfa468hn6czff8kazw08crazehcxaqa2uaa" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpsq650w975vqsf6ajj5x4wdzfnrh64kmw7sljqz7wts6k0p6l36d0huls3" string),(1 uint64)} gno.land/p/sys/validators.Validator),(struct{("g1wsa9j6nel8ltt6q2lmf78585ymyfh5nsvhaxa3" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqma223maxmnw4f42kfqppvgyn8dr8wu7mhtdm6lcq64303a3vlln8xdmms" string),(1 uint64)} gno.land/p/sys/validators.Validator)] []gno.land/p/sys/validators.Validator)`

	//  / Call the parsing function and handle any potential errors.
	validators, err := parser.GnolandValidatorsParser(input)
	assert.NoError(t, err)

	// If parsing is successful, print the results.
	fmt.Printf("✅ Successfully parsed %d validators.\n\n", len(validators))
	for i, v := range validators {
		fmt.Printf("--- Validator %d ---\n", i+1)
		fmt.Printf("Address:     %s\n", v.Address)
		fmt.Printf("PubKey:      %s\n", v.PubKey)
		fmt.Printf("VotingPower: %d\n\n", v.VotingPower)
	}
}

func TestGnolandValidatorInfoParser(t *testing.T) {
	// The raw string data to be parsed.
	input := `(struct{("devx-val-2" string),("NT DevX Val2" string),("g1t9ctfa468hn6czff8kazw08crazehcxaqa2uaa" .uverse.address),("gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpsq650w975vqsf6ajj5x4wdzfnrh64kmw7sljqz7wts6k0p6l36d0huls3" string),(true bool),(&<nil> *gno.land/p/demo/ownable/exts/authorizable.Authorizable)} gno.land/r/gnoland/valopers.Valoper)`

	//  / Call the parsing function and handle any potential errors.
	valoper, err := parser.GnolandValidatorInfoParser(input)
	assert.NoError(t, err)

	fmt.Println("✅ Successfully parsed Valoper data (auth part excluded).")
	fmt.Println("-----------------------------------------------------")
	fmt.Printf("Name:        %s\n", valoper.Name)
	fmt.Printf("Description: %s\n", valoper.Description)
	fmt.Printf("Address:     %s\n", valoper.Address)
	fmt.Printf("Public Key:  %s\n", valoper.PubKey)
	fmt.Printf("Is Active:   %t\n", valoper.KeepRunning)
	fmt.Println("-----------------------------------------------------")
}

func TestGnolandValidatorInfo(t *testing.T) {
	rpcEndpoint := os.Getenv("TEST_BLOCK_GNOLAND_HOST_ADDRESS")
	commonApp := common.NewCommonApp(p)
	commonApp.SetRPCEndPoint(rpcEndpoint)

	valAddr := types.GnolandValidator{
		Address:     "g13762rd7y8s7jcc6uc4lyxv269hguchhpyzaamt",
		PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq5ndww8w6qrxgfdeastcx2lsuuk5r8w9jckkgevylq6duw59d54n935fq2",
		VotingPower: 1,
	}
	_, err := GetGnolandValidatorInfo(commonApp.CommonClient, valAddr)
	assert.NoError(t, err)
}
