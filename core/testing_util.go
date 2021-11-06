package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

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

// createBufferWithData creates an AlgorandBuffer, inserts data with a
// Manage-routine, waits for data to be published to the blockchain, and
// then returns the buffer, as well as the WaitGroup and CancelFunc for the
// managing routine
func fillBufferWithData(a *AlgorandBuffer, m map[string]string) (*sync.WaitGroup, context.CancelFunc, error) {
	wg, cancel := a.SpawnManagingRoutine(nil)
	// Manager-routine as an actual struct with functions.
	err := a.PutElements(m)
	if err != nil {
		return nil, nil, err
	}

	// wait for values to be published to buffer
	data, _ := a.GetBuffer()
	for now := time.Now(); len(data) == 0; {
		time.Sleep(time.Millisecond * 50)
		data, _ = a.GetBuffer()
		if time.Now().Sub(now) > time.Second * 30 {
			break
		}
	}
	if len(data) != len(m) {
		return nil, nil, fmt.Errorf("expected %d data points, got %d", len(m), len(data))
	}
	return wg, cancel, nil
}

// bufferLengthWithin returns nil if the given buffer reached a given buffer
// length within a given time frame. Otherwise returns error. This is a blocking
// call
func bufferLengthWithin(a *AlgorandBuffer, l int, t time.Duration) error {
	now := time.Now()
	for  time.Now().Sub(now) < t {
		data, err := a.GetBuffer()
		if err != nil {
			return err
		}
		if len(data) == l {
			return nil
		}
	}
	return errors.New("time limit exceeded. buffer doesn't have correct length")
}

// waitTimeout waits for the sync.WaitGroup until a timeout.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
    c := make(chan struct{})
    go func() {
        defer close(c)
        wg.Wait()
    }()
    select {
    case <-c:
        return false // completed normally
    case <-time.After(timeout):
        return true // timed out
    }
}
