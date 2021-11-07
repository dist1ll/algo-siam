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

// fillBufferWithData fills an empty(!) AlgorandBuffer with data. It spawns a
// Manage-routine, waits for data to be published to the blockchain, and
// then returns the WaitGroup and CancelFunc for the managing routine
func fillBufferWithData(a *AlgorandBuffer, m map[string]string) (*sync.WaitGroup, context.CancelFunc, error) {
	wg, cancel := a.SpawnManagingRoutine(nil)
	// Manager-routine as an actual struct with functions.
	err := a.PutElements(m)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	err = bufferLengthWithin(a, len(m), time.Second * 30)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return wg, cancel, nil
}

// bufferLengthWithin returns nil if the given buffer reached a given buffer
// length within a given time frame. Otherwise returns error. This is a blocking
// call
func bufferLengthWithin(a *AlgorandBuffer, l int, t time.Duration) error {
	now := time.Now()
	for  time.Now().Sub(now) < t {
		time.Sleep(time.Millisecond * 50)
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

// bufferEqualsWithin returns nil if the 'buffer[key] = expected' within a given duration.
// Use this to check if a value got correctly inserted or updated into the AlgorandBuffer.
func bufferEqualsWithin(a *AlgorandBuffer, key string, expected string, t time.Duration) error {
	now := time.Now()
	data, err := a.GetBuffer()
	for  time.Now().Sub(now) < t {
		data, err = a.GetBuffer()
		if err != nil {
			return err
		}
		if _, ok := data[key]; !ok {
			continue
		} else if data[key] == expected {
			return nil
		}
	}
	return fmt.Errorf("time limit exceeded. buffer['%s']='%s', but expected '%s'", key, data[key], expected)
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
