package util

import "testing"

func TestContain(t *testing.T) {
	type args struct {
		a string
		b []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "空值判断",
			args: args{a: "", b: []string{"DO1", "DO2"}},
			want: false,
		},
		{
			name: "包含1个元素",
			args: args{a: "DO1", b: []string{"DO1", "DO2"}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contain(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Contain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	type args struct {
		a []string
		b []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "空值判断",
			args: args{a: []string{}, b: []string{"DO1", "DO2"}},
			want: false,
		},
		{
			name: "不相等判断1",
			args: args{a: []string{"DO3"}, b: []string{"DO1", "DO2"}},
			want: false,
		},
		{
			name: "不相等判断2",
			args: args{a: []string{"DO1"}, b: []string{"DO1", "DO2"}},
			want: false,
		},
		{
			name: "不相等判断3",
			args: args{a: []string{"DO2", "DO2"}, b: []string{"DO1", "DO2"}},
			want: false,
		},
		{
			name: "乱序相等判断",
			args: args{a: []string{"DO2", "DO1"}, b: []string{"DO1", "DO2"}},
			want: true,
		},
		{
			name: "相等判断",
			args: args{a: []string{"DO1", "DO2"}, b: []string{"DO1", "DO2"}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
