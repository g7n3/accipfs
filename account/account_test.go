package account

import (
	"github.com/glvd/accipfs/config"
	"testing"
)

func TestNewAccount(t *testing.T) {
	config.WorkDir = "D:\\workspace\\pvt"
	if err := config.LoadConfig(); err != nil {
		t.Fatal(err)
		return
	}
	acc, err := NewAccount(config.Global())
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := acc.Save(config.Global()); err != nil {
		t.Fatal(err)
		return
	}

}