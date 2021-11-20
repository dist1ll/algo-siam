package siam

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// PartitionMap partitions a given map into partitions with a given size.
func PartitionMap(data map[string]string, size int) []map[string]string {
	if size >= len(data) {
		return []map[string]string{data}
	}

	partitions := make([]map[string]string, 0)
	currentP := make(map[string]string)
	for k, v := range data {
		if len(currentP) == size {
			partitions = append(partitions, currentP)
			currentP = make(map[string]string)
		}
		currentP[k] = v
	}
	if len(currentP) > 0 {
		partitions = append(partitions, currentP)
	}
	return partitions
}

// Utility function that creates an AlgorandBuffer, and subsequently deletes the application
// so that only one remains.
func createBufferAndRemoveApps(t *testing.T) *AlgorandBuffer {
	buffer, err := CreateAlgorandBufferFromEnv(nil)
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

// mapContainsMap returns true if every element of a map 'sub' is contained in
// a map 'super'
func mapContainsMap(super map[string]string, sub map[string]string) bool {
	if len(sub) == 0 || len(super) == 0 {
		return false
	}
	if len(super) < len(sub) {
		return false
	}
	for key, subVal := range sub {
		if superVal, ok := super[key]; ok {
			if superVal != subVal {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// bufferLengthWithin returns nil if the given buffer reached a given buffer
// length within a given time frame. Otherwise returns error. This is a blocking
// call
func bufferLengthWithin(a *AlgorandBuffer, l int, t time.Duration) error {
	now := time.Now()
	for time.Now().Sub(now) < t {
		time.Sleep(time.Millisecond * 50)
		ctx, cancel := context.WithTimeout(context.Background(), t-time.Now().Sub(now))
		data, err := a.GetBuffer(ctx)
		cancel()
		if err != nil {
			return err
		}
		if len(data) == l {
			return nil
		}
	}
	return fmt.Errorf("time limit exceeded. buffer length expected %d", l)
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
