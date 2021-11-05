package core

import (
    "context"
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
func createBufferWithData(t *testing.T) (*AlgorandBuffer, *sync.WaitGroup, context.CancelFunc) {
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)
	wg, cancel := buffer.SpawnManagingRoutine()
	// Manager-routine as an actual struct with functions.
	err = buffer.PutElements(map[string]string{
		"1000" : "Astralis",
		"1001" : "Vitality",
		"1002" : "Gambit",
		"1003" : "Na'Vi",
		"1004" : "Furia",
		"1005" : "G2",
	})
	assert.Nil(t, err)

	// wait for values to be published to buffer
	data, _ := buffer.GetBuffer()
	for now := time.Now(); len(data) == 0; {
		time.Sleep(time.Millisecond * 200)
		data, _ = buffer.GetBuffer()
		if time.Now().Sub(now) > time.Second * 30 {
			break
		}
	}
	assert.EqualValues(t, 6, len(data))
	return buffer, wg, cancel
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
