package generator

import (
	"testing"
)

func Test_validIdentifier(t *testing.T) {
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
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validIdentifier(tt.args.name); got != tt.want {
				t.Errorf("validIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_uniqueIdentifiers_new(t *testing.T) {
	ref := make(chan map[string]bool, 1)
	ref <- map[string]bool{
		"test": true,
	}

	type fields struct {
		ref chan map[string]bool
	}
	type args struct {
		name      string
		qualifier string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{{
		"basic test",
		fields{ref: ref},
		args{"test", "Handler"},
		"testHandler",
	}, {
		"snakecase test",
		fields{ref: ref},
		args{"testIsInTheHeart!!", "Handler"},
		"testIsInTheHeart",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &uniqueIdentifiers{
				ref: tt.fields.ref,
			}
			if got := u.new(tt.args.name, tt.args.qualifier); got != tt.want {
				t.Errorf("uniqueIdentifiers.new() = %v, want %v", got, tt.want)
			}
		})
	}
}
