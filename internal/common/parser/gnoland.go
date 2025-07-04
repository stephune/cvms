package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/cosmostation/cvms/internal/common/types"
)

func GnolandBFTValidatorParser(resp []byte) ([]types.GnolandBFTValidator, error) {
	var result types.GnolandBFTValidatorResponse
	err := json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}
	return result.Result.Validators, nil
}

func GnolandBlockParser(resp []byte) (
	/* block height */ int64,
	/* block timestamp */ time.Time,
	/* block proposer addrss */ string,
	/* txs in the block */ []types.Tx,
	/* last comit block height*/ int64,
	/* block validators signatures */ []types.GnolandSignature,
	error,
) {
	var result types.GnolandBlockResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, time.Time{}, "", nil, 0, nil, err
	}
	heightString, blockTimestamp := result.Result.Block.Header.Height, result.Result.Block.Header.Time

	//  result.Result.Block.LastCommit.Height

	blockHeight, err := strconv.ParseInt(heightString, 10, 64)
	if err != nil {
		return 0, time.Time{}, "", nil, 0, nil, err
	}

	var lastCommitHeightString string
	sigs := make([]types.GnolandSignature, 0)
	for idx, pc := range result.Result.Block.LastCommit.PreCommits {
		if idx == 0 {
			lastCommitHeightString = pc.Height
		}

		sigs = append(sigs, types.GnolandSignature{
			Timestamp:        pc.Timestamp,
			ValidatorAddress: pc.ValidatorAddress,
			ValidatorIndex:   pc.ValidatorIndex,
			Signature:        pc.Signature,
		})
	}

	lastCommitBlockHeight, err := strconv.ParseInt(lastCommitHeightString, 10, 64)
	if err != nil {
		return 0, time.Time{}, "", nil, 0, nil, err
	}

	txs := result.Result.Block.Data.Txs
	proposerAddress := result.Result.Block.Header.ProposerAddress
	return blockHeight, blockTimestamp, proposerAddress, txs, lastCommitBlockHeight, sigs, nil
}

func GnolandABCIParser(resp []byte) (string, error) {
	var result types.GnolandABCIResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	decodedData, err := base64.StdEncoding.DecodeString(result.Result.Response.ResponseBase.Data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", decodedData), nil
}

// ParseValidators takes a raw string input from gno.land and parses it into a slice of Validator structs.
// It returns the slice of validators and an error if parsing fails.
func GnolandValidatorsParser(input string) ([]types.GnolandValidator, error) {
	// 1. Define the regular expression to capture the fields for each validator.
	// The pattern uses capturing groups () for Address, PubKey, and VotingPower.
	re := regexp.MustCompile(`\("([^"]+)" \.uverse\.address\),\("([^"]+)" string\),\((\d+) uint64\)`)

	// 2. Find all occurrences that match the regular expression.
	// FindAllStringSubmatch returns a slice where each element contains the full match and its captured subgroups.
	matches := re.FindAllStringSubmatch(input, -1)

	// If no matches are found, the input might be malformed or empty.
	if len(matches) == 0 {
		return nil, fmt.Errorf("no validators found in the input string")
	}

	// 3. Create a slice to store the parsed results.
	// Pre-allocating the slice with make improves performance slightly.
	validators := make([]types.GnolandValidator, 0, len(matches))

	// 4. Loop through each match to extract data and create Validator structs.
	for _, match := range matches {
		// match[0] is the full matched string, e.g., ("g12y...")
		// match[1] is the first captured group (Address)
		// match[2] is the second captured group (PubKey)
		// match[3] is the third captured group (VotingPower)

		// Convert the VotingPower string to a uint64 integer.
		votingPower, err := strconv.ParseUint(match[3], 10, 64)
		if err != nil {
			// If conversion fails, return an error, indicating a problem with the data.
			return nil, fmt.Errorf("error converting voting power '%s': %w", match[3], err)
		}

		// 5. Append the newly created Validator struct to the slice.
		validators = append(validators, types.GnolandValidator{
			Address:     match[1],
			PubKey:      match[2],
			VotingPower: votingPower,
		})
	}

	// 6. Return the populated slice and a nil error to indicate success.
	return validators, nil
}

// ParseValoper now parses the input string without the final Authorizable field.
func GnolandValidatorInfoParser(input string) (types.GnolandValidatorInfo, error) {
	// 1. The regular expression has been modified.
	// The final comma after the boolean capture group has been removed.
	re := regexp.MustCompile(
		`struct{\("([^"]+)" string\),` + // 1: Name
			`\("([^"]+)" string\),` + // 2: Description
			`\("([^"]+)" \.uverse\.address\),` + // 3: Address
			`\("([^"]+)" string\),` + // 4: PubKey
			`\((true|false) bool\)` + // 5: IsActive (boolean) - NO TRAILING COMMA
			`.*`, // Match the rest of the string without capturing it.
	)

	// 2. Find the first match and its subgroups.
	matches := re.FindStringSubmatch(input)

	// We now expect 6 elements: 1 for the full match + 5 for the captured groups.
	if len(matches) != 6 {
		return types.GnolandValidatorInfo{}, fmt.Errorf("invalid input format: expected 5 fields, found %d", len(matches)-1)
	}

	// 3. Convert the captured boolean string to a bool type.
	isActive, err := strconv.ParseBool(matches[5])
	if err != nil {
		return types.GnolandValidatorInfo{}, fmt.Errorf("could not parse boolean value: %w", err)
	}

	// 4. Create the Valoper struct with the extracted data.
	valoper := types.GnolandValidatorInfo{
		Name:        matches[1],
		Description: matches[2],
		Address:     matches[3],
		PubKey:      matches[4],
		KeepRunning: isActive,
	}

	// 5. Return the populated struct.
	return valoper, nil
}
