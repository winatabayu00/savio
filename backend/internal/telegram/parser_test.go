package telegram

import "testing"

func TestParseExpense(t *testing.T) {
	cases := []struct {
		in     string
		desc   string
		minor  int64
		wantOK bool
	}{
		{"chocolate hazelnut dutch 24000", "chocolate hazelnut dutch", 2400000, true},
		{"grab food 24.500", "grab food", 2450000, true},
		{"grab food 24,500", "grab food", 2450000, true},
		{"pulsa 50rb", "pulsa", 5000000, true},
		{"token listrik 100k", "", 0, false},
		{"kopi 2.5jt", "kopi", 250000000, true},
		{"gojek 45 ribu", "gojek", 4500000, true},
		{"2 botol susu 50000", "2 botol susu", 5000000, true},
		{"chocolate 1.000.000", "chocolate", 100000000, true},
		{"macet", "", 0, false},
		{"", "", 0, false},
		{"2 botol susu", "", 0, false},
	}
	for _, c := range cases {
		got, gotAmt, ok := parseExpense(c.in)
		if !c.wantOK {
			if ok {
				t.Errorf("parseExpense(%q) should fail, got (%q, %d/100)", c.in, got, gotAmt)
			}
			continue
		}
		if !ok {
			t.Errorf("parseExpense(%q) failed, want (%q, %d/100)", c.in, c.desc, c.minor)
			continue
		}
		if got != c.desc || gotAmt != c.minor {
			t.Errorf("parseExpense(%q) = (%q, %d/100) want (%q, %d/100)", c.in, got, gotAmt, c.desc, c.minor)
		}
	}
}
