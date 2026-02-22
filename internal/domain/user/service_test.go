package user

import "testing"

func TestGenerateRandomAlnum(t *testing.T) {
	alnum, err := generateRandomAlnum(6)
	if err != nil {
		t.Fatalf("generateRandomAlnum failed: %v", err)
	}
	t.Logf("generateRandomAlnum: %s", alnum)
}
