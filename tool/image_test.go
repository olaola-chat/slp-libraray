package tool

import "testing"

func TestAppendCdnHost(t *testing.T) {
	type args struct {
		ul string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "1",
			args: args{
				ul: "/aa",
			},
			want: "https://image.caihongxq.com/aa",
		},
		{
			name: "2",
			args: args{
				ul: "aa",
			},
			want: "https://image.caihongxq.com/aa",
		},
		{
			name: "3",
			args: args{
				ul: "http://xxx",
			},
			want: "http://xxx",
		},
		{
			name: "4",
			args: args{
				ul: "https:/xxx",
			},
			want: "https:/xxx",
		},
		{
			name: "5",
			args: args{
				ul: "/static/room/lecture_room_banner.png",
			},
			want: "https://image.caihongxq.com/static/room/lecture_room_banner.png",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Img.AppendCdnHost(tt.args.ul); got != tt.want {
				t.Errorf("AppendCdnHost() = %v, want %v", got, tt.want)
			}
		})
	}
}
