package main

import "testing"

func TestVersionLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"0.1.0", "0.1.1", true},
		{"0.1.0", "0.2.0", true},
		{"0.1.0", "1.0.0", true},
		{"0.1.0", "0.1.0", false},
		{"0.1.1", "0.1.0", false},
		// 预发布 < 正式版
		{"0.1.0-rc.6", "0.1.0", true},
		{"0.1.0", "0.1.0-rc.6", false},
		// 同核心，rc 数字比较
		{"0.1.0-rc.5", "0.1.0-rc.6", true},
		{"0.1.0-rc.10", "0.1.0-rc.6", false},
		// alpha < beta < rc
		{"0.1.0-alpha.1", "0.1.0-beta.1", true},
		{"0.1.0-beta.1", "0.1.0-rc.1", true},
		// 核心不同时预发布不比较
		{"0.1.0-rc.9", "0.2.0-alpha.1", true},
	}
	for _, c := range cases {
		if got := versionLess(c.a, c.b); got != c.want {
			t.Errorf("versionLess(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestNormalizeVersion(t *testing.T) {
	cases := map[string]string{
		"0.1.0-rc.6":     "0.1.0-rc.6",
		" 0.1.0-rc.6\n":  "0.1.0-rc.6",
		"v0.1.0":         "0.1.0",
		"dsh 0.1.0-rc.6": "0.1.0-rc.6",
		"0.1.0-rc.6\n\n": "0.1.0-rc.6",
	}
	for in, want := range cases {
		if got := normalizeVersion(in); got != want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", in, got, want)
		}
	}
}
