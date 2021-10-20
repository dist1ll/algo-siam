package core

import (
	"errors"
	"testing"
)


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