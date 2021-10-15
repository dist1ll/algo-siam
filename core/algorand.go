package core

import (
	"encoding/base64"
	"os"
)

// Environment variable names for Algorand
const envURLNode = "AEMA_URL_NODE"
const envAlgodToken = "AEMA_ALGOD_TOKEN"

// envPrivateKey Base64 encoded private key of our targeted account/application.
const envPrivateKey = "AEMA_PRIVATE_KEY"

// GetAlgorandConfig returns a config tuple needed to interact with the Algorand node
// addr: URL of the node to connect to
// token: API token to query algod over HTTP
// pk: Private key of our account of interest
func GetAlgorandConfig() (addr string, token string, pk []byte) {
	addr =  os.Getenv(envURLNode)
	token = os.Getenv(envAlgodToken)
	base64key := os.Getenv(envPrivateKey)

	pk, _ = base64.StdEncoding.DecodeString(base64key)

	return addr, token, pk
}
