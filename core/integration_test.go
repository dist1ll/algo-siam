// +build integration

package core

import (
	"context"
	"github.com/m2q/aema/core/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIntegration_StoreData(t *testing.T) {
	buffer, _ := CreateAlgorandBufferFromEnv()

	ctx, cancel := context.WithTimeout(context.Background(), client.AlgorandDefaultTimeout)
	info, err := buffer.Client.AccountInformation(buffer.AccountCrypt.Address.String(), ctx)
	cancel()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(info.CreatedApps))
}
