package agent

import (
	"reflect"
	"testing"
)

func TestSplitDomains(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single", "example.com", []string{"example.com"}},
		{"multiple", "a.example.com,b.example.com", []string{"a.example.com", "b.example.com"}},
		{"whitespace and blanks", " a.example.com , , b.example.com ,", []string{"a.example.com", "b.example.com"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitDomains(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("splitDomains(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
