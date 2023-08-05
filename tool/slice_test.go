package tool

import (
	"reflect"
	"testing"
)

func TestUniqueSlice(t *testing.T) {
	type args struct {
		s1 []uint32
		s2 []uint32
	}
	tests := []struct {
		name string
		args args
		want []uint32
	}{
		{
			name: "没有重复",
			args: args{
				s1: []uint32{1111, 222, 444},
				s2: []uint32{55, 66, 77},
			},
			want: []uint32{1111, 222, 444, 55, 66, 77},
		},
		{
			name: "s2重复",
			args: args{
				s1: []uint32{1111, 222, 444},
				s2: []uint32{55, 66, 222, 77, 444},
			},
			want: []uint32{1111, 222, 444, 55, 66, 77},
		},
		{
			name: "s1重复",
			args: args{
				s1: []uint32{1111, 222, 444, 55},
				s2: []uint32{55, 66, 77},
			},
			want: []uint32{1111, 222, 444, 55, 66, 77},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slice.UniqueSlice(tt.args.s1, tt.args.s2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UniqueSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
