package util

import (
	"encoding/binary"
	"reflect"
	"testing"
)

var byte8 = []byte{0, 0, 0, 1, 0, 0, 0, 0}
var byte16 = []byte{0, 0, 0, 0, 1, 0, 0, 0}
var byte254 = []byte{0, 1, 1, 1, 1, 1, 1, 1}
var byte511 = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1}

func TestByteToBits(t *testing.T) {
	type args struct {
		d byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "8",
			args: args{d: 8},
			want: byte8,
		},
		{
			name: "16",
			args: args{d: 16},
			want: byte16,
		},
		{
			name: "254",
			args: args{d: 254},
			want: byte254,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ByteToBits(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByteToBits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitsToByte(t *testing.T) {
	type args struct {
		buf []byte
	}
	tests := []struct {
		name  string
		args  args
		wantU byte
	}{
		{
			name:  "8",
			args:  args{buf: byte8},
			wantU: 8,
		},
		{
			name:  "16",
			args:  args{buf: byte16},
			wantU: 16,
		},
		{
			name:  "254",
			args:  args{buf: byte254},
			wantU: 254,
		},
		{
			name:  "255",
			args:  args{buf: byte511},
			wantU: 255,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotU := BitsToByte(tt.args.buf); gotU != tt.wantU {
				t.Errorf("BitsToByte() = %v, want %v", gotU, tt.wantU)
			}
		})
	}
}

var big12816 = []byte{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
var big40 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0}
var little12816 = []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0}
var little40 = []byte{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func TestTwoByteToBits(t *testing.T) {
	type args struct {
		data  []byte
		order binary.ByteOrder
	}
	tests := []struct {
		name    string
		args    args
		wantDst []byte
	}{
		{
			name: "128/16",
			args: args{
				data:  []byte{128, 16},
				order: binary.BigEndian,
			},
			wantDst: big12816,
		},
		{
			name: "4/0",
			args: args{
				data:  []byte{4, 0},
				order: binary.BigEndian,
			},
			wantDst: big40,
		},
		{
			name: "128/16",
			args: args{
				data:  []byte{128, 16},
				order: binary.LittleEndian,
			},
			wantDst: little12816,
		},
		{
			name: "4/0",
			args: args{
				data:  []byte{4, 0},
				order: binary.LittleEndian,
			},
			wantDst: little40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDst := TwoByteToBits(tt.args.data, tt.args.order); !reflect.DeepEqual(gotDst, tt.wantDst) {
				t.Errorf("TwoByteToBits() = %v, want %v", gotDst, tt.wantDst)
			}
		})
	}
}
