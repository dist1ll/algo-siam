package core

import (
	"errors"
	core "github.com/m2q/aema/core/client"
	"testing"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/stretchr/testify/assert"
)

// If HealthCheck and token verification works, expect no errors
func TestAlgorandBuffer_HealthAndTokenPass(t *testing.T) {
	client := core.CreateAlgorandClientMock("", "")
	_, err := NewAlgorandBuffer(client, GeneratePrivateKey64())
	if err != nil {
		t.Errorf("failing health check doesn't return error %s", err)
	}
}

// If the HealthCheck is not working, return error upon buffer creation
func TestAlgorandBuffer_NoHealth(t *testing.T) {
	client := core.CreateAlgorandClientMock("", "")
	client.SetError(true, (*core.AlgorandMock).HealthCheck)
	buffer, err := NewAlgorandBuffer(client, GeneratePrivateKey64())
	if err == nil {
		t.Errorf("failing health check doesn't return error %s", err)
	}
	// buffer should still have created account
	assert.NotEqual(t, models.Account{}, buffer.Account)
}

// If the Token Verification is not working, return error upon buffer creation
func TestAlgorandBuffer_IncorrectToken(t *testing.T) {
	client := core.CreateAlgorandClientMock("", "")
	client.SetError(true, (*core.AlgorandMock).Status)
	buffer, err := NewAlgorandBuffer(client, GeneratePrivateKey64())
	if err == nil {
		t.Errorf("failing token verification doesn't return error %s", err)
	}
	// buffer should still have created account
	assert.NotEqual(t, models.Account{}, buffer.Account)
}

func TestChainAppCreationDeletion(t *testing.T) {
	return
	a, err := NewAlgorandBufferFromEnv()

	//fmt.Println(runtime.FuncForPC(reflect.ValueOf(a.CreateApplication).Pointer()).Name())
	//a.CreateApplication()

	if e := &(NoApplication{}); errors.As(err, &e) {
		t.Logf("no apps registered under %s", e.Account.Address)
		err = a.CreateApplication()
		if err != nil {
			t.Fatalf("Couldn't create application %s", err)
		}
	} else if e := &(TooManyApplications{}); errors.As(err, &e) {
		t.Fatalf("too many applications registered under {%s}", e.Account.Address)
	} else if err != nil {
		t.Fatalf("fatal error %s", err)
	}

	t.Logf("found an app. proceeding to delete.")
	err = a.DeleteApplication(a.AppId)
	if err != nil {
		t.Fatalf("error deleting app %s", err)
	}
}
