package password

import "testing"

func TestHashVerify(t *testing.T) {
	h, err := Hash("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := Verify("correct horse battery staple", h)
	if err != nil || !ok {
		t.Fatalf("expected match, ok=%v err=%v", ok, err)
	}
	ok, _ = Verify("wrong password", h)
	if ok {
		t.Fatal("wrong password should not match")
	}
}

func TestVerifyBadHash(t *testing.T) {
	if _, err := Verify("x", "not-a-hash"); err == nil {
		t.Fatal("expected error for malformed hash")
	}
}

func TestHashesUnique(t *testing.T) {
	a, _ := Hash("same")
	b, _ := Hash("same")
	if a == b {
		t.Fatal("expected salt randomization to produce distinct hashes")
	}
}
