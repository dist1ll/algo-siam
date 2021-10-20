package core

import (
	"errors"
	"github.com/m2q/aema/core/client"
	"testing"
	//"github.com/stretchr/testify"
)


// If the HealthCheck is not working, return error upon buffer creation
func TestAlgorandBuffer_NoHealth(t *testing.T) {
	client := core.CreateAlgorandClientMock("", "")
	client.SetError(false, (*core.AlgorandMock).HealthCheck)

	_, err := NewAlgorandBuffer(client, GeneratePrivateKey64())
	if err != nil {
		t.Errorf("failing health check doesn't return error %s", err)
	}

	// buffer should still have created account

}


func TestChainAppCreationDeletion(t *testing.T) {
	a, err := NewAlgorandBufferFromEnv()

	//fmt.Println(runtime.FuncForPC(reflect.ValueOf(a.CreateApplication).Pointer()).Name())
	//a.CreateApplication()
	return
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