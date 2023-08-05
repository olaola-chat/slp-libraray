package tool

import (
	"testing"
)

const (
	china = "中华人民共和国"
	raw   = `\u4e2d\u534e\u4eba\u6c11\u5171\u548c\u56fd`
)

func TestEscapeUnicode(t *testing.T) {
	result := Str.EscapeUnicode(china)
	t.Log("Test EscapeUnicode", result == raw)
}

func TestUnescapeUnicode(t *testing.T) {
	result, _ := Str.UnescapeUnicode([]byte(raw))
	t.Log("Test UnescapeUnicode", string(result) == china)
}

func Test_str_CleanString(t *testing.T) {
	type args struct {
		word string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "(梦) | 梦",
			args: args{word: "(梦) | 梦"},
			want: "梦梦",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &str{}
			if got := st.CleanString(tt.args.word); got != tt.want {
				t.Errorf("CleanString() = %v, want %v", got, tt.want)
			}
		})
	}
}
