package generator

import (
	"testing"
)

func Test_ValidIdentifier(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		"basic test",
		args{"this is the end"},
		"thisIsTheEnd",
	}, {
		"case sensitive",
		args{"thisIsTheEnd"},
		"thisIsTheEnd",
	}, {
		"dash characters",
		args{"this-is-the-end"},
		"thisIsTheEnd",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidIdentifier(tt.args.name); got != tt.want {
				t.Errorf("ValidIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}
