//go:build unit

package siam

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/m2q/algo-siam/client"
	"github.com/stretchr/testify/assert"
)

// If HealthCheck and token verification works, expect no errors
func TestAlgorandBuffer_HealthAndTokenPass(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	_, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
	if err != nil {
		t.Errorf("failing health check doesn't return error %s", err)
	}
}

// If the HealthCheck is not working, return error upon buffer creation
func TestAlgorandBuffer_NoHealth(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.SetError(true, (*client.AlgorandMock).HealthCheck)
	buffer, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
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
	buffer, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
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
	buffer, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
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

// AppChannel should not return anything if the DeleteApplication function
// returns errors.
func TestAlgorandBuffer_DeletionError(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6, 18, 32)
	c.SetError(true, (*client.AlgorandMock).DeleteApplication)

	_, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
	if err == nil {
		t.Fatalf("blocking deleteApp doesn't return error.")
	}
}

func TestAlgorandBuffer_DeleteAppsWhenTooMany(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	c.CreateDummyApps(6, 18, 32)

	// Check if client is made valid.
	assert.False(t, client.ValidAccount(c.Account))
	_, err := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
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
	_, _ = CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
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
	_, _ = CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
	assert.True(t, client.ValidAccount(c.Account))

	// Check if remaining app is the one that was created first
	assert.EqualValues(t, 50, c.Account.CreatedApps[0].CreatedAtRound)
}

func TestAlgorandBuffer_Creation(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")

	assert.False(t, client.ValidAccount(c.Account))
	_, _ = CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
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
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)

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

	if waitTimeout(&wg, time.Millisecond*100) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

// Check if Manage() goroutine respects cancel, when no arguments are put
// into the buffer, and the health check times are very long.
func TestAlgorandBuffer_ManageQuits2(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer.Manage(ctx, &ManageConfig{
			SleepTime:           time.Minute,
			HealthCheckInterval: time.Minute})
	}()

	time.Sleep(time.Millisecond * 10)
	cancel()
	if waitTimeout(&wg, time.Millisecond*100) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

// Test if buffer restores valid account state after adding an application
// AFTER the buffer has been verified and initialized
func TestAlgorandBuffer_AppAddedAfterSetup(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)

	// Add application after setup
	c.AddDummyApps(56)
	assert.False(t, client.ValidAccount(c.Account))

	wg, cancel := buffer.SpawnManagingRoutine(&ManageConfig{})

	// Manage routine should make account valid in less than a second
	now := time.Now()
	for !client.ValidAccount(c.Account) && time.Now().Sub(now) < time.Second {
		time.Sleep(time.Millisecond)
	}
	assert.True(t, client.ValidAccount(c.Account))
	cancel()
	wg.Wait()
}

func TestAlgorandBuffer_GetBuffer(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
	data, err := buffer.GetBuffer(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 0)
}

func TestAlgorandBuffer_PutElements(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
	// store in buffer
	data := map[string]string{
		"2654658": "Astralis",
	}
	wg, cancel := buffer.SpawnManagingRoutine(&ManageConfig{})
	err := putElementsAndWait(buffer, data, time.Second)
	assert.Nil(t, err)
	// confirm buffer size
	d, err := buffer.GetBuffer(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(d), "buffer should have exactly one element")
	cancel()
	wg.Wait()
}

func TestAlgorandBuffer_PutElementsTooBig(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
	// store kv pair that exceeds 128 byte
	data := map[string]string{
		"key": strings.Repeat("x", 128),
	}
	wg, cancel := buffer.SpawnManagingRoutine(&ManageConfig{})
	err := putElementsAndWait(buffer, data, time.Second)
	assert.NotNil(t, err)
	// confirm buffer size
	d, err := buffer.GetBuffer(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 0, len(d), "buffer should be empty, because kv pair exceeds 128 byte total")
	cancel()
	wg.Wait()
}

func TestAlgorandBuffer_TooMany(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
	// Put Maximum Data
	data := make(map[string]string, client.GlobalBytes)
	for i := 0; i < client.GlobalBytes; i++ {
		data[strconv.Itoa(i)] = ""
	}
	wg, cancel := buffer.SpawnManagingRoutine(&ManageConfig{})
	err := putElementsAndWait(buffer, data, time.Second)
	assert.Nil(t, err)

	err = putElementsAndWait(buffer, map[string]string{"x": "y"}, time.Millisecond*100)
	assert.NotNil(t, err)

	// confirm buffer size
	d, err := buffer.GetBuffer(context.Background())
	assert.Nil(t, err)
	_, exists := d["x"]
	assert.False(t, exists, "buffer should not have 'x' element")
	cancel()
	wg.Wait()
}

func TestAlgorandBuffer_ContainsWithin(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)
	// store in buffer
	data := map[string]string{
		"x": "y",
	}
	wg, cancel := buffer.SpawnManagingRoutine(&ManageConfig{})
	err := putElementsAndWait(buffer, data, time.Second*2)
	assert.Nil(t, err)
	// Contains should return true, because x is inside the buffer
	assert.True(t, buffer.ContainsWithin(map[string]string{"x": "y"}, time.Second))
	// Contains should return false on an empty map
	assert.False(t, buffer.ContainsWithin(map[string]string{}, time.Second))
	cancel()
	wg.Wait()
}

func TestAlgorandBuffer_DeleteElements(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)

	// store in buffer
	data := map[string]string{
		"1000": "Astralis",
		"1001": "Vitality",
		"1002": "Gambit",
		"1003": "OG",
	}
	wg, cancel := buffer.SpawnManagingRoutine(&ManageConfig{})
	err := putElementsAndWait(buffer, data, time.Second)
	assert.Nil(t, err)

	err = buffer.DeleteElements("1001")
	assert.Nil(t, err)
	d, _ := buffer.GetBuffer(context.Background())
	fmt.Println(len(d))
	// We expect 3 items
	assert.Nil(t, bufferLengthWithin(buffer, 3, time.Second*5))

	// Make sure that key=1001 doesn't exist
	b, err := buffer.GetBuffer(context.Background())
	_, ok := b["1001"]
	assert.False(t, ok)

	cancel()
	wg.Wait()
}

// Assuming we provide >16 delete arguments AND afterwards immediately provide
// >16 store arguments, test that the Manage routine executes ALL delete arguments
// first, before ever executing the store arguments
func TestAlgorandBuffer_DeletePriority(t *testing.T) {
	c := client.CreateAlgorandClientMock("", "")
	buffer, _ := CreateAlgorandBuffer(c, client.GeneratePrivateKey64(), nil)

	// Fill Buffer first
	data := make(map[string]string, client.GlobalBytes)
	for i := 0; i < client.GlobalBytes; i++ {
		data[strconv.Itoa(i)] = ""
	}
	wg, cancel := buffer.SpawnManagingRoutine(&ManageConfig{})
	err := putElementsAndWait(buffer, data, time.Second)
	assert.Nil(t, err)

	// Now create delete args and store args
	del := make([]string, 2*client.MaxArgs)
	for i := 0; i < client.MaxArgs*2; i++ {
		t := i % client.MaxArgs
		del[i] = strconv.Itoa(t)
	}

	put := make(map[string]string, client.MaxArgs)
	for i := 0; i < client.MaxArgs; i++ {
		put[strconv.Itoa(i)] = ""
	}

	// Put delete and store args
	err = buffer.DeleteElements(del...)
	assert.Nil(t, err)
	err = buffer.PutElements(put)
	assert.Nil(t, err)

	// We expect full buffer
	assert.Nil(t, bufferLengthWithin(buffer, client.GlobalBytes, time.Second*1))

	cancel()
	wg.Wait()
}
