package core

import (
	"context"
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

// putElementsAndWait fills a buffer with data and waits for the data to be written
// to the AlgorandBuffer, until a given timeout t. This is a blocking call.
func putElementsAndWait(a *AlgorandBuffer, m map[string]string, t time.Duration) error {
	err := a.PutElements(m)
	if err != nil {
		return err
	}
	err = bufferLengthWithin(a, len(m), t)
	if err != nil {
		return err
	}
	return nil
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
	return fmt.Errorf("time limit exceeded. buffer length expected %d", l)
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
