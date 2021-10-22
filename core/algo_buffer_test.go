package core

import (
	"github.com/m2q/aema/core/client"
	"testing"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/stretchr/testify/assert"
)

// If HealthCheck and token verification works, expect no errors
func TestAlgorandBuffer_HealthAndTokenPass(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	_, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err != nil {
		t.Errorf("failing health check doesn't return error %s", err)
	}
}

// If the HealthCheck is not working, return error upon buffer creation
func TestAlgorandBuffer_NoHealth(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.SetError(true, (*client.AlgorandMock).HealthCheck)
	buffer, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err == nil {
		t.Errorf("failing health check doesn't return error %s", err)
	}
	// buffer should still have created account
	assert.NotEqual(t, models.Account{}, buffer.AccountCrypt)
}

// If the Token Verification is not working, return error upon buffer creation
func TestAlgorandBuffer_IncorrectToken(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.SetError(true, (*client.AlgorandMock).Status)
	buffer, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err == nil {
		t.Errorf("failing token verification doesn't return error %s", err)
	}
	// buffer should still have created account
	assert.NotEqual(t, models.Account{}, buffer.AccountCrypt)
}

// Without calling buffer's Manage() function, storing on the buffer is invalid
// and should result in a panic
func TestAlgorandBuffer_RequireManagement(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())

	shouldPanic := func() {
		buffer.StoreBuffer(make(map[string]string, 3))
	}
	assert.Panics(t, shouldPanic)
}

func TestAlgorandBuffer_DeleteAppsWhenTooMany(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6, 18, 32)
	buffer, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err != nil {
		t.Error(err)
	}

	go buffer.Manage()

	return
	acc, _ := c.AccountInformation("", nil)
	for len(acc.CreatedApps) != 1 && client.FulfillsSchema(acc.CreatedApps[0]) {
		acc, _ = c.AccountInformation("", nil)
	}
}