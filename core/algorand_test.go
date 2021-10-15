package core

import (
	"testing"
)

func TestEnvironmentVariablesExist(t *testing.T) {
	addr, token, key := GetAlgorandEnvironmentVars()

	if addr == "" {
		t.Errorf("Env var %s not set or empty!", envURLNode)
	}
	if token == "" {
		t.Errorf("Env var %s not set or empty!", envAlgodToken)
	}
	if len(key) == 0 {
		t.Errorf("Env var %s not set, empty or incorrect! Pay attention to base64 encoding", envPrivateKey)
	}
}
