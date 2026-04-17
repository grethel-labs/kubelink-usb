package utils

import "testing"

func TestParseUSBIdentifiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		id         string
		wantVendor string
		wantProd   string
	}{
		{
			name:       "valid vendor and product",
			id:         "1234:abcd",
			wantVendor: "1234",
			wantProd:   "abcd",
		},
		{
			name:       "missing separator",
			id:         "1234abcd",
			wantVendor: "",
			wantProd:   "",
		},
		{
			name:       "too many segments",
			id:         "12:34:56",
			wantVendor: "",
			wantProd:   "",
		},
		{
			name:       "empty input",
			id:         "",
			wantVendor: "",
			wantProd:   "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotVendor, gotProd := ParseUSBIdentifiers(tc.id)
			if gotVendor != tc.wantVendor || gotProd != tc.wantProd {
				t.Fatalf("ParseUSBIdentifiers(%q)=(%q,%q) want (%q,%q)", tc.id, gotVendor, gotProd, tc.wantVendor, tc.wantProd)
			}
		})
	}
}
