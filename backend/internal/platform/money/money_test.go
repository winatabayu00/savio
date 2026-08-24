package money

import "testing"

func TestParseDecimalToMinorUnits(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"1500000.00", 150000000, false},
		{"100", 10000, false},
		{"0.01", 1, false},
		{"1.1", 110, false},
		{"-100000", 0, true},
		{"+100", 0, true},
		{"0", 0, true},
		{"1.234", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, c := range cases {
		got, err := ParseDecimalToMinorUnits(c.in)
		if c.wantErr != (err != nil) {
			t.Errorf("%q: wantErr=%v got %v", c.in, c.wantErr, err)
			continue
		}
		if !c.wantErr && got != c.want {
			t.Errorf("%q: got %d want %d", c.in, got, c.want)
		}
	}
}

func TestMinorUnitsToDecimalString(t *testing.T) {
	cases := []struct{ in int64; want string }{
		{150000000, "1500000.00"},
		{1, "0.01"},
		{12345, "123.45"},
		{-500, "-5.00"},
		{0, "0.00"},
	}
	for _, c := range cases {
		if got := MinorUnitsToDecimalString(c.in); got != c.want {
			t.Errorf("MinorUnitsToDecimal(%d)=%s want %s", c.in, got, c.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	want := int64(987654321)
	s := MinorUnitsToDecimalString(want)
	got, err := ParseDecimalToMinorUnits(s)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("round trip %d -> %s -> %d", want, s, got)
	}
}