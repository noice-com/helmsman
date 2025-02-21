package app

import (
	"testing"
)

func Test_specFromYAML(t *testing.T) {
	type args struct {
		file string
		s    *StateFiles
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test case 1 -- Valid YAML",
			args: args{
				file: "../../examples/example-spec.yaml",
				s:    new(StateFiles),
			},
			want: true,
		}, {
			name: "test case 2 -- Invalid Yaml",
			args: args{
				file: "../../tests/Invalid_example_spec.yaml",
				s:    new(StateFiles),
			},
			want: false,
		},
	}

	teardownTestCase, err := setupStateFileTestCase(t)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTestCase(t)
	for _, tt := range tests {
		// os.Args = append(os.Args, "-f ../../examples/example.yaml")
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.s.specFromYAML(tt.args.file)
			if err != nil {
				t.Log(err)
			}

			got := err == nil
			if got != tt.want {
				t.Errorf("specFromYaml() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_specFileSort(t *testing.T) {
	type args struct {
		files fileOptionArray
	}
	tests := []struct {
		name string
		args args
		want [3]int
	}{
		{
			name: "test case 1 -- Files sorted by priority",
			args: args{
				files: fileOptionArray(
					[]fileOption{
						fileOption{"third.yaml", 0},
						fileOption{"first.yaml", -20},
						fileOption{"second.yaml", -10},
					}),
			},
			want: [3]int{-20, -10, 0},
		},
	}

	teardownTestCase, err := setupStateFileTestCase(t)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTestCase(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.files.sort()

			got := [3]int{}
			for i, f := range tt.args.files {
				got[i] = f.priority
			}
			if got != tt.want {
				t.Errorf("files from spec file are not sorted by priority = %v, want %v", got, tt.want)
			}
		})
	}
}
