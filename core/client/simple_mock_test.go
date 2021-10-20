package core

import (
	"context"
	"testing"
)

func TestAlgorandMock(t *testing.T) {
	client := CreateAlgorandClientSimpleMock("", "")
	client.SetError(true, (*AlgorandMock).AccountInformation, (*AlgorandMock).GetApplicationByID)

	if _, err := client.AccountInformation("", context.Background()); err == nil {
		t.Errorf("expected error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err == nil {
		t.Errorf("expected error")
	}

	client.SetError(false, (*AlgorandMock).AccountInformation)

	if _, err := client.AccountInformation("", context.Background()); err != nil {
		t.Errorf("expected no error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err == nil {
		t.Errorf("expected error")
	}

	client.ClearFunctionErrors()

	if _, err := client.AccountInformation("", context.Background()); err != nil {
		t.Errorf("expected no error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err != nil {
		t.Errorf("expected no error")
	}
}
