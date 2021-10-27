// +build unit

package core

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/m2q/aema/core/client"
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

// AppChannel should not return anything if the DeleteApplication function
// returns errors.
func TestAlgorandBuffer_DeletionError(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6, 18, 32)
	c.SetError(true, (*client.AlgorandMock).DeleteApplication)

	_, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err == nil {
		t.Fatalf("blocking deleteApp doesn't return error.")
	}
}

func TestAlgorandBuffer_DeleteAppsWhenTooMany(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6, 18, 32)

	// Check if client is made valid.
	assert.False(t, client.ValidAccount(c.Account))
	_, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	assert.True(t, client.ValidAccount(c.Account))

	assert.Nil(t, err)
}

func TestAlgorandBuffer_DeletePartial(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyAppsWithSchema(models.ApplicationStateSchema{}, 6, 18, 32)

	// Set one application to have correct schema
	g, l := client.GenerateSchemasModel()
	c.Account.CreatedApps[0].Params = models.ApplicationParams{GlobalStateSchema: g, LocalStateSchema: l}

	// Check if client is made valid.
	assert.False(t, client.ValidAccount(c.Account))
	_, _ = CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	assert.True(t, client.ValidAccount(c.Account))
}

// Given several applications with the right schema, delete the one that has
// been created most recently
func TestAlgorandBuffer_DeleteNewest(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6, 18, 32)
	c.Account.CreatedApps[0].CreatedAtRound = 200
	c.Account.CreatedApps[1].CreatedAtRound = 50
	c.Account.CreatedApps[2].CreatedAtRound = 150

	// Check if client is made valid.
	assert.False(t, client.ValidAccount(c.Account))
	_, _ = CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	assert.True(t, client.ValidAccount(c.Account))

	// Check if remaining app is the one that was created first
	assert.EqualValues(t, 50, c.Account.CreatedApps[0].CreatedAtRound)
}

func TestAlgorandBuffer_Creation(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")

	assert.False(t, client.ValidAccount(c.Account))
	_, _ = CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	assert.True(t, client.ValidAccount(c.Account))
}

// Check if Manage() goroutine quits, when we cancel the provided context
// Whenever the Manage() routine receives errors from the node or application,
// it may fall asleep for a certain amount of time (to not ddos a server that
// might have some problems).
// The *_ManageQuits* tests make sure that the cancel() call is respected in
// every situation.
func TestAlgorandBuffer_ManageQuits(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())

	c.SetError(true, (*client.AlgorandMock).HealthCheck)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer.Manage(ctx, &ManageConfig{SleepTime: time.Minute})
	}()

	time.Sleep(time.Millisecond * 10)

	cancel()

	if waitTimeout(&wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

// Check if Manage() goroutine respects cancel, when no arguments are put
// into the buffer, and the health check times are very long.
func TestAlgorandBuffer_ManageQuits2(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer.Manage(ctx, &ManageConfig{
			SleepTime: time.Minute,
			HealthCheckInterval: time.Minute})
	}()

	time.Sleep(time.Millisecond * 10)
	cancel()
	if waitTimeout(&wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

// Test if buffer restores valid account state after adding an application
// AFTER the buffer has been verified and initialized
func TestAlgorandBuffer_AppAddedAfterSetup(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())

	// Add application after setup
	c.AddDummyApps(56)
	assert.False(t, client.ValidAccount(c.Account))

	go buffer.Manage(context.Background(), &ManageConfig{})

	// Manage() should make account valid in less than a second
	now := time.Now()
	for !client.ValidAccount(c.Account) && time.Now().Sub(now) < time.Second {
		time.Sleep(time.Millisecond)
	}
	assert.True(t, client.ValidAccount(c.Account))
}

func TestAlgorandBuffer_GetBuffer(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())
	data, err := buffer.GetBuffer()
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 0)
}

func TestAlgorandBuffer_PutElements(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64())

	go buffer.Manage(context.Background(), &ManageConfig{})

	// store in buffer
	values := map[string]string { "2654658" : "Astralis" }
	buffer.PutElements(values)
	d, _ := buffer.GetBuffer()

	// buffer should be non-zero within a second
	now := time.Now()
	for ; len(d) == 0 && time.Now().Sub(now) < time.Second; d, _ = buffer.GetBuffer() {
		time.Sleep(time.Millisecond)
	}
	d, _ = buffer.GetBuffer()
	assert.Equal(t, 1, len(d), "buffer should have exactly one element")
}