package emoji_test

import (
	"testing"

	"github.com/sabafly/sabafly-lib/v2/emoji"
)

func TestEmoji(t *testing.T) {
	type args struct {
		a string
	}
	{
		tests := []struct {
			name string
			args args
			want bool
		}{
			{
				name: "case1",
				args: args{a: "😁"},
				want: true,
			},
			{
				name: "case2",
				args: args{a: "🇯🇵"},
				want: true,
			},
			{
				name: "case3",
				args: args{a: "🧑🏽‍🚀"},
				want: true,
			},
			{
				name: "case4",
				args: args{a: "1"},
				want: false,
			},
			{
				name: "case5",
				args: args{a: "A"},
				want: false,
			},
			{
				name: "case6",
				args: args{a: "<:modify:1082025248330891388>"},
				want: true,
			},
		}
		for _, v := range tests {
			t.Run(v.name, func(t *testing.T) {
				if got := emoji.MatchString(v.args.a); got != v.want {
					t.Errorf("UnicodeEmoji.MatchString() = %v, want %v", got, v.want)
				}
			})
		}
	}
	{
		tests := []struct {
			name string
			args args
			want int
		}{
			{
				name: "case1",
				args: args{a: "😁"},
				want: 1,
			},
			{
				name: "case2",
				args: args{a: `🇯🇵`},
				want: 1,
			},
			{
				name: "case3",
				args: args{a: "🧑🏽‍🚀"},
				want: 1,
			},
			{
				name: "case4",
				args: args{a: "1"},
				want: 0,
			},
			{
				name: "case5",
				args: args{a: "A"},
				want: 0,
			},
			{
				name: "case6",
				args: args{a: "<:modify:1082025248330891388>"},
				want: 1,
			},
		}
		for _, v := range tests {
			t.Run(v.name, func(t *testing.T) {
				if got := emoji.FindAllString(v.args.a); len(got) != v.want {
					t.Errorf("FindAllString() = %#v, len() = %v, want %v", got, len(got), v.want)
				}
			})
		}
	}
}
