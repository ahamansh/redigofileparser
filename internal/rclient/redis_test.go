package rclient

import "testing"

func TestGetRedisClient(t *testing.T) {

	rds, err := GetRedisClient()

	if rds != nil {
		t.Error("rds nil is expected")
	}

	if err == nil {
		t.Error("Error is expected")
	}

}
