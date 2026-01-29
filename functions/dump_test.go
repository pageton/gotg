package functions

import (
	"strings"
	"testing"
)

func TestDump(t *testing.T) {
	tests := []struct {
		name string
		val  any
		key  []string
		want string
	}{
		{
			name: "struct no key",
			val:  struct{ Name string }{Name: "test"},
			want: "{\n  \"Name\": \"test\"\n}",
		},
		{
			name: "struct with key",
			val:  struct{ ID int }{ID: 42},
			key:  []string{"MSG"},
			want: "[MSG] {\n  \"ID\": 42\n}",
		},
		{
			name: "map no key",
			val:  map[string]int{"a": 1},
			want: "{\n  \"a\": 1\n}",
		},
		{
			name: "map with key",
			val:  map[string]int{"a": 1},
			key:  []string{"MAP"},
			want: "[MAP] {\n  \"a\": 1\n}",
		},
		{
			name: "empty key string",
			val:  map[string]int{"x": 5},
			key:  []string{""},
			want: "{\n  \"x\": 5\n}",
		},
		{
			name: "nil value",
			val:  nil,
			want: "null",
		},
		{
			name: "nil value with key",
			val:  nil,
			key:  []string{"KEY"},
			want: "[KEY] null",
		},
		{
			name: "string value",
			val:  "hello",
			want: "\"hello\"",
		},
		{
			name: "int value",
			val:  123,
			want: "123",
		},
		{
			name: "emoji key",
			val:  struct{ Out bool }{Out: false},
			key:  []string{"📱 MSG"},
			want: "[📱 MSG] {\n  \"Out\": false\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Dump(tt.val, tt.key...)
			if got != tt.want {
				t.Errorf("Dump() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestDumpIndentation(t *testing.T) {
	val := struct {
		A string
		B int
	}{A: "x", B: 1}
	got := Dump(val)
	if !strings.Contains(got, "  \"A\"") {
		t.Errorf("expected 2-space indent, got:\n%s", got)
	}
}

type mockDumpable struct {
	inner struct{ Value int }
}

func (m *mockDumpable) DumpValue() any {
	return m.inner
}

func TestDumpWithDumpable(t *testing.T) {
	d := &mockDumpable{inner: struct{ Value int }{Value: 99}}

	got := Dump(d, "TEST")
	want := "[TEST] {\n  \"Value\": 99\n}"
	if got != want {
		t.Errorf("Dump(Dumpable) =\n%s\nwant:\n%s", got, want)
	}

	got = Dump(d)
	want = "{\n  \"Value\": 99\n}"
	if got != want {
		t.Errorf("Dump(Dumpable, no key) =\n%s\nwant:\n%s", got, want)
	}
}
