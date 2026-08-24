package money

import "testing"

func TestParseMinorUnits(t *testing.T) {
	cases := map[string]int64{
		"1500000.00": 150000000,
		"1.5":        150,
		"0":          0,
		"-100000":    -10000000,
		"123.45":     12345,
	}
	for in, want := range cases {
		got, err := ParseMinorUnits(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("Parse(%q) = %d, want %d", in, got, want)
		}
	}
	for _, bad := range []string{"", "abc", "1.234", "1.2.3", "100000000000000000000"} {
		if _, err := ParseMinorUnits(bad); err == nil {
			t.Fatalf("Parse(%q) expected error", bad)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	for _, v := range []int64{0, 1, 150, 12345, -100, 99999999999} {
		s := FormatMinorUnits(v)
		got, err := ParseMinorUnits(s)
		if err != nil || got != v {
			t.Fatalf("round trip %d via %q = %d err %v", v, s, got, err)
		}
	}
}
