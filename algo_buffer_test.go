//go:build unit

package siam

import (
	"context"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/m2q/algo-siam/client"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

// If HealthCheck and token verification works, expect no errors
func TestAlgorandBuffer_HealthAndTokenPass(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	_, err := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err != nil {
		t.Errorf("failing health check doesn't return error %s", err)
	}
}

// If the HealthCheck is not working, return error upon buffer creation
func TestAlgorandBuffer_NoHealth(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.SetError(true, (*client.AlgorandMock).HealthCheck)
	buffer, err := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
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
	buffer, err := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err == nil {
		t.Errorf("failing token verification doesn't return error %s", err)
	}
	// buffer should still have created account
	assert.NotEqual(t, models.Account{}, buffer.AccountCrypt)
}

// If the target account is valid, correctly funded and has a valid application,
// then after the buffer has been initialized, it should have valid fields
func TestAlgorandBuffer_CorrectBufferWhenValid(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6)
	buffer, err := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err != nil {
		t.Fatal(err)
	}

	info, err := buffer.Client.AccountInformation(buffer.AccountCrypt.Address.String(), context.Background())
	if err != nil {
		t.Fatalf("eror getting account info %s", err)
	}
	assert.Equal(t, 1, len(info.CreatedApps))
	assert.EqualValues(t, 6, buffer.AppId)
}

func TestAlgorandBuffer_DeletionError(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6, 18, 32)
	c.SetError(true, (*client.AlgorandMock).DeleteApplication)

	_, err := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	if err == nil {
		t.Fatalf("blocking deleteApp doesn't return error.")
	}
}

func TestAlgorandBuffer_DeleteAppsWhenTooMany(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6, 18, 32)

	// Check if client is made valid.
	assert.False(t, client.ValidAccount(c.Account))
	_, err := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
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
	_, _ = NewAlgorandBuffer(c, client.GeneratePrivateKey64())
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
	_, _ = NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	assert.True(t, client.ValidAccount(c.Account))

	// Check if remaining app is the one that was created first
	assert.EqualValues(t, 50, c.Account.CreatedApps[0].CreatedAtRound)
}

func TestAlgorandBuffer_Creation(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")

	assert.False(t, client.ValidAccount(c.Account))
	_, _ = NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	assert.True(t, client.ValidAccount(c.Account))
}

func TestAlgorandBuffer_GetBuffer(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	data, err := buffer.GetBuffer(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 0)
}

func TestAlgorandBuffer_PutElements(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	// store in buffer
	data := map[string]string{
		"2654658": "Astralis",
	}
	err := buffer.PutElements(context.Background(), data)
	assert.Nil(t, err)

	// confirm buffer size
	d, err := buffer.GetBuffer(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(d), "buffer should have exactly one element")
}

func TestAlgorandBuffer_PutElementsTooBig(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	// store kv pair that exceeds 128 byte
	data := map[string]string{
		"key": strings.Repeat("x", 128),
	}
	err := buffer.PutElements(context.Background(), data)
	assert.NotNil(t, err)
	// confirm buffer size
	d, err := buffer.GetBuffer(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 0, len(d), "buffer should be empty, because kv pair exceeds 128 byte total")
}

func TestAlgorandBuffer_TooMany(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := NewAlgorandBuffer(c, client.GeneratePrivateKey64())
	// Put Maximum Data
	data := make(map[string]string, client.GlobalBytes)
	for i := 0; i < client.GlobalBytes; i++ {
		data[strconv.Itoa(i)] = ""
	}
	err := buffer.PutElements(context.Background(), data)
	assert.Nil(t, err)

	err = buffer.PutElements(context.Background(), map[string]string{"x": "y"})
	assert.Nil(t, err)

	// confirm buffer size
	d, err := buffer.GetBuffer(context.Background())
	assert.Nil(t, err)
	_, exists := d["x"]
	assert.False(t, exists, "buffer should not have 'x' element")
}

func TestAlgorandBuffer_Contains(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := NewAlgorandBuffer(c, client.GeneratePrivateKey64())

	// store in buffer
	data := map[string]string{
		"x": "y",
	}
	err := buffer.PutElements(context.Background(), data)
	assert.Nil(t, err)

	b, err := buffer.Contains(context.Background(), map[string]string{"x": "y"})
	assert.True(t, b)
	assert.Nil(t, err)

	b, err = buffer.Contains(context.Background(), map[string]string{})
	assert.False(t, b)
	assert.Nil(t, err)
}

func TestAlgorandBuffer_DeleteElements(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := NewAlgorandBuffer(c, client.GeneratePrivateKey64())

	// store in buffer
	data := map[string]string{
		"1000": "Astralis",
		"1001": "Vitality",
		"1002": "Gambit",
		"1003": "OG",
	}
	err := buffer.PutElements(context.Background(), data)
	assert.Nil(t, err)

	err = buffer.DeleteElements(context.Background(), "1001")
	assert.Nil(t, err)

	d, _ := buffer.GetBuffer(context.Background())
	assert.EqualValues(t, len(d), 3)

	// Make sure that key=1001 doesn't exist
	b, err := buffer.GetBuffer(context.Background())
	_, ok := b["1001"]
	assert.False(t, ok)
}

func TestAlgorandBuffer_AchieveDesiredState(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := NewAlgorandBuffer(c, client.GeneratePrivateKey64())

	// Put Maximum Data
	data := make(map[string]string, client.GlobalBytes)
	for i := 0; i < client.GlobalBytes; i++ {
		data[strconv.Itoa(i)] = ""
	}
	assert.Nil(t, buffer.PutElements(context.Background(), data))
	assert.Nil(t, buffer.AchieveDesiredState(context.Background(), map[string]string{}))

	d, _ := buffer.GetBuffer(context.Background())
	assert.EqualValues(t, 0, len(d))

	// Same keys, different values
	data = make(map[string]string, client.GlobalBytes)
	for i := 0; i < client.GlobalBytes-1; i++ {
		data[strconv.Itoa(i)] = "val"
	}
	assert.Nil(t, buffer.AchieveDesiredState(context.Background(), data))
	d, _ = buffer.GetBuffer(context.Background())
	assert.Equal(t, "val", d["0"])
	assert.Equal(t, "", d[strconv.Itoa(client.GlobalBytes-1)])
}
