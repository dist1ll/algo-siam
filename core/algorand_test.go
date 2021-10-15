package core

import "testing"

func TestEnvironmentVariablesExist(t *testing.T) {
	addr, token, key := GetAlgorandConfig()

	if addr == "" {
		t.Errorf("Env var %s not set or empty!", envURLNode)
	}
	if token == "" {
		t.Errorf("Env var %s not set or empty!", envAlgodToken)
	}
	if key == "" {
		t.Errorf("Env var %s not set or empty!", envPrivateKey)
	}
}
