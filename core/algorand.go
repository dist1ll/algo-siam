package core

import "os"

// Environment variable names for Algorand
const envURLNode = "AEMA_URL_NODE"
const envAlgodToken = "AEMA_ALGOD_TOKEN"
const envPrivateKey = "AEMA_PRIVATE_KEY"

// Returns three strings:
// 1) addr: URL of the node to connect to
// 2) token: API token to query algod over HTTP
// 3) pk: Private key of our account of interest
func GetAlgorandConfig() (addr string, token string, pk string) {
	addr =  os.Getenv(envURLNode)
	token = os.Getenv(envAlgodToken)
	pk = os.Getenv(envPrivateKey)
	return addr, token, pk
}
