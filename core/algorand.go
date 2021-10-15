package core

import (
	"os"
)

// Environment variable names for Algorand
const envURLNode = "AEMA_URL_NODE"
const envAlgodToken = "AEMA_ALGOD_TOKEN"

// envPrivateKey Base64 encoded private key of our targeted account/application.
const envPrivateKey = "AEMA_PRIVATE_KEY"

// GetAlgorandEnvironmentVars returns a config tuple needed to interact with the Algorand node
// addr: URL of the node to connect to
// token: API token to query algod over HTTP
// base64key: Base64-encoded private key of our account of interest
func GetAlgorandEnvironmentVars() (addr string, token string, base64key string) {
	addr =  os.Getenv(envURLNode)
	token = os.Getenv(envAlgodToken)
	base64key = os.Getenv(envPrivateKey)
	return addr, token, base64key
}
