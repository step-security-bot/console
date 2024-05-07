package devices

import (
	"testing"
)

type test struct {
	name string
	mock func(*MockRepository)
	err  error
}

func TestSendPowerAction(t *testing.T) {
	t.Parallel()

	tests := []test{}
}
