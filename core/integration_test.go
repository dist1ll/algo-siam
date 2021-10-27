// +build integration

package core

import (
	"context"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/m2q/aema/core/client"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// Note: During integration tests, we need to make sure that there's no
// leak of goroutines. Because of this, every test will check for routine
// exits with WaitGroups

func TestIntegration_ValidAccount(t *testing.T) {
	buffer, _ := CreateAlgorandBufferFromEnv()

	ctx, cancel := context.WithTimeout(context.Background(), client.AlgorandDefaultTimeout)
	info, err := buffer.Client.AccountInformation(buffer.AccountCrypt.Address.String(), ctx)
	cancel()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(info.CreatedApps))
}

// Utility function that creates an AlgorandBuffer, and subsequently deletes the application
// so that only one remains.
func createBufferAndRemoveApps(t *testing.T) *AlgorandBuffer {
	buffer, err := CreateAlgorandBufferFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	err = buffer.Client.DeleteApplication(buffer.AccountCrypt, buffer.AppId)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that app has 0 apps
	info, err := buffer.Client.AccountInformation(buffer.AccountCrypt.Address.String(), context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 0, len(info.CreatedApps))

	return buffer
}

// Test if app removal works
func TestIntegration_RemoveAccount(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
}

// Remove application, and see if Manage re-creates the application
func TestIntegration_AccountGetsRestored(t *testing.T) {
	buffer := createBufferAndRemoveApps(t)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer.Manage(ctx, &ManageConfig{})
	}()

	var info models.Account
	for !client.ValidAccount(info) {
		info, _ = buffer.Client.AccountInformation(buffer.AccountCrypt.Address.String(), ctx)
		time.Sleep(time.Second)
	}

	cancel()

	if waitTimeout(&wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}
