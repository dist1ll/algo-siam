package client

import (
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"os"
	"strings"
)

// EnvURLNode is the environment variable name for the URL of the remote Algorand instance.
// The value should be an IP address with port number (e.g. `191.158.1.12:1225`) or a domain
// like `https://testnet.algoexplorerapi.io`
const EnvURLNode = "SIAM_URL_NODE"

// EnvAlgodToken is the environment variable name for the API token of the Algorand node.
// The algod service requires an API token to use most of the API calls. If you're running
// the node yourself, look for an algod.token file (or refer to Algorand official node docs).
const EnvAlgodToken = "SIAM_ALGOD_TOKEN"

// EnvPrivateKey is the environment variable name of the Base64 encoded private key of our
// target account. The 'target account' is the one used to create and maintain the Algorand
// application that stores our buffer data.
const EnvPrivateKey = "SIAM_PRIVATE_KEY"

// EnvHeadersNode is the environment variable name of the custom headers that are sent
// to the Algorand algod-node. This is useful if you're calling custom Node providers
// like PureStake that define their own header and API key. Use the following syntax:
// "header1:value1&header2:value2"
const EnvHeadersNode = "SIAM_HEADERS_NODE"

// GetAlgorandEnvironmentVars returns a config tuple needed to interact with the Algorand node.
//  addr:      URL of the Algorand node
//  token:     API token for the algod endpoint
//  base64key: Base64-encoded private key of the Algorand application
// You can use these to initialize a client.CreateAlgorandClientWrapper.
func GetAlgorandEnvironmentVars() (URL string, token string, base64key string, headers []*common.Header) {
	URL = os.Getenv(EnvURLNode)
	token = os.Getenv(EnvAlgodToken)
	base64key = os.Getenv(EnvPrivateKey)
	rawHeader := os.Getenv(EnvHeadersNode)
	for _, s := range strings.Split(rawHeader, "&") {
		kv := strings.Split(s, ":")
		if len(kv) != 2 {
			continue
		}
		headers = append(headers, &common.Header{Key: kv[0], Value: kv[1]})
	}
	return URL, token, base64key, headers
}

// HasEnvironmentVars returns true if the necessary configuration environment variables are set.
// These environment variables store configuration for the AlgorandClient.
// Key and URL MUST exist. Either token or headers need to be set. Returns false if both are missing.
func HasEnvironmentVars() bool {
	_, existsURL := os.LookupEnv(EnvURLNode)
	_, existsToken := os.LookupEnv(EnvAlgodToken)
	_, existsHeaders := os.LookupEnv(EnvHeadersNode)
	_, existsKey := os.LookupEnv(EnvPrivateKey)
	return existsURL && existsKey && (existsToken || existsHeaders)
}
