package db

import (
	"testing"

	"github.com/hrygo/gosms/bootstrap"
)

func TestInitDB(t *testing.T) {
	InitDB(bootstrap.ConfigYml, "AuthClient.Mongo")
}
