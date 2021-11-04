package client

import "os"


// Environment variable names for Algorand
const EnvURLNode = "AEMA_URL_NODE"
const EnvAlgodToken = "AEMA_ALGOD_TOKEN"

// EnvPrivateKey is the Base64 encoded private key of our targeted account or application.
const EnvPrivateKey = "AEMA_PRIVATE_KEY"

// GetAlgorandEnvironmentVars returns a config tuple needed to interact with the Algorand node.
//  addr: URL of the Algorand node
//  token: API token for the algod endpoint
//  base64key: Base64-encoded private key of the Algorand application
func GetAlgorandEnvironmentVars() (URL string, token string, base64key string) {
	URL = os.Getenv(EnvURLNode)
	token = os.Getenv(EnvAlgodToken)
	base64key = os.Getenv(EnvPrivateKey)
	return URL, token, base64key
}