package writers

import "testing"

func TestSanitizeName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{{
		name: "test passes value through",
		args: args{"thisIsTheEnd"},
		want: "thisIsTheEnd",
	}, {
		name: "deal with dashes",
		args: args{"this-is-the-end"},
		want: "thisIsTheEnd",
	}, {
		name: "test case from error encountered",
		args: args{"co-list"},
		want: "coList",
	}, {
		name:    "spaces not allowed",
		args:    args{"co list"},
		want:    "Invalid name 'co list'",
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeName(tt.args.name)
			if err != nil {
				if tt.wantErr {
					if err.Error() != tt.want {
						t.Errorf("SanitizeName() = %v, want %v", err.Error(), tt.want)
					}
				} else {
					t.Errorf("SanitizeName() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if got != tt.want {
				t.Errorf("SanitizeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
