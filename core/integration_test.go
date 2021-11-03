// +build integration

package core

import (
	"context"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/m2q/aema/core/client"
	"github.com/stretchr/testify/assert"
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

// Test if app removal works
func TestIntegration_RemoveAccount(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
}

// Remove application, and see if Manage re-creates the application
func TestIntegration_AccountGetsRestored(t *testing.T) {
	buffer := createBufferAndRemoveApps(t)
	wg, cancel := buffer.SpawnManagingRoutine()

	var info models.Account
	for !client.ValidAccount(info) {
		info, _ = buffer.Client.AccountInformation(buffer.AccountCrypt.Address.String(), context.Background())
		time.Sleep(time.Second)
	}

	cancel()

	if waitTimeout(wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

func TestIntegration_PushData(t *testing.T) {
	buffer, err := CreateAlgorandBufferFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	wg, cancel := buffer.SpawnManagingRoutine()
	return
	err = buffer.PutElements(map[string]string{
		"554213" : "Astralis",
	})
	assert.Nil(t, err)

	cancel()

	if waitTimeout(wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}