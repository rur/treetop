package writers

import (
	"reflect"
	"testing"

	"github.com/rur/treetop/generator"
)

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

func Test_iterateSortedBlocks(t *testing.T) {
	type args struct {
		blocks map[string][]generator.PartialDef
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				map[string][]generator.PartialDef{},
			},
			want:    []string{},
			wantErr: false,
		}, {
			name: "sort three blocks",
			args: args{
				map[string][]generator.PartialDef{
					"C": []generator.PartialDef{generator.PartialDef{Name: "third"}},
					"A": []generator.PartialDef{generator.PartialDef{Name: "first"}},
					"Z": []generator.PartialDef{generator.PartialDef{Name: "last"}},
					"B": []generator.PartialDef{generator.PartialDef{Name: "second"}},
				},
			},
			want:    []string{"A", "B", "C", "Z"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockNames := make([]string, len(tt.args.blocks))
			got, err := iterateSortedBlocks(tt.args.blocks)
			for i := 0; i < len(got); i++ {
				blockNames[i] = got[i].name
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("iterateSortedBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(blockNames, tt.want) {
				t.Errorf("iterateSortedBlocks() = %v, want %v", blockNames, tt.want)
			}
		})
	}
}
