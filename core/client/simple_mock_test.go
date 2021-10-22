package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAlgorandMock_DummyApps tests if created dummy apps have the correct ID
func TestAlgorandMock_DummyApps(t *testing.T) {
	c := CreateAlgorandClientMock("", "")
	c.CreateDummyApps(2, 5, 8)
	assert.Equal(t, 3, len(c.Account.CreatedApps))
	assert.EqualValues(t, 2, c.Account.CreatedApps[0].Id)
	assert.EqualValues(t, 8, c.Account.CreatedApps[2].Id)

	for _, val := range c.Account.CreatedApps {
		assert.EqualValues(t, globalBytes, val.Params.GlobalStateSchema.NumByteSlice)
		assert.EqualValues(t, globalInts, val.Params.GlobalStateSchema.NumUint)
		assert.EqualValues(t, localBytes, val.Params.LocalStateSchema.NumByteSlice)
		assert.EqualValues(t, localInts, val.Params.LocalStateSchema.NumUint)
	}
}

func TestAlgorandMock_ErrorFunctions(t *testing.T) {
	client := CreateAlgorandClientMock("", "")

	client.SetError(true, (*AlgorandMock).AccountInformation, (*AlgorandMock).GetApplicationByID)

	if _, err := client.AccountInformation("", context.Background()); err == nil {
		t.Errorf("expected error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err == nil {
		t.Errorf("expected error")
	}

	client.SetError(false, (*AlgorandMock).AccountInformation)

	if _, err := client.AccountInformation("", context.Background()); err != nil {
		t.Errorf("expected no error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err == nil {
		t.Errorf("expected error")
	}

	client.ClearFunctionErrors()

	if _, err := client.AccountInformation("", context.Background()); err != nil {
		t.Errorf("expected no error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err != nil {
		t.Errorf("expected no error")
	}
}
