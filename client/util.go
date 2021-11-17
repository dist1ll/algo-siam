package client

import "os"

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

// GetAlgorandEnvironmentVars returns a config tuple needed to interact with the Algorand node.
//  addr:      URL of the Algorand node
//  token:     API token for the algod endpoint
//  base64key: Base64-encoded private key of the Algorand application
// You can use these to initialize a client.CreateAlgorandClientWrapper.
func GetAlgorandEnvironmentVars() (URL string, token string, base64key string) {
	URL = os.Getenv(EnvURLNode)
	token = os.Getenv(EnvAlgodToken)
	base64key = os.Getenv(EnvPrivateKey)
	return URL, token, base64key
}

// HasEnvironmentVars returns true if all configuration environment variables are set.
// These environment variables store configuration for the AlgorandClient.
func HasEnvironmentVars() bool {
	_, existsURL := os.LookupEnv(EnvURLNode)
	_, existsToken := os.LookupEnv(EnvAlgodToken)
	_, existsKey := os.LookupEnv(EnvPrivateKey)
	return existsURL && existsToken && existsKey
}
