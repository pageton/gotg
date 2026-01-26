package functions

import (
	"testing"
	"time"
)

func TestGetOpt(t *testing.T) {
	tests := []struct {
		name     string
		opts     []int
		want     int
		wantOk   bool
	}{
		{
			name:   "no options",
			opts:   nil,
			want:   0,
			wantOk: false,
		},
		{
			name:   "no options - empty slice",
			opts:   []int{},
			want:   0,
			wantOk: false,
		},
		{
			name:   "valid option",
			opts:   []int{42},
			want:   42,
			wantOk: true,
		},
		{
			name:   "zero value option",
			opts:   []int{0},
			want:   0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetOpt(tt.opts...)
			if got != tt.want {
				t.Errorf("GetOpt() = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetOpt() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}

	// Test panic on too many options
	t.Run("panic on too many options", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("GetOpt() should panic with >1 option")
			}
		}()
		GetOpt([]int{1, 2}...)
	})
}

func TestGetOptDef(t *testing.T) {
	tests := []struct {
		name  string
		def   int
		opts  []int
		want  int
	}{
		{
			name: "no options - returns default",
			def:  42,
			opts: nil,
			want: 42,
		},
		{
			name: "no options - empty slice",
			def:  42,
			opts: []int{},
			want: 42,
		},
		{
			name: "valid option - returns option",
			def:  42,
			opts: []int{100},
			want: 100,
		},
		{
			name: "zero value option - returns default",
			def:  42,
			opts: []int{0},
			want: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOptDef(tt.def, tt.opts...); got != tt.want {
				t.Errorf("GetOptDef() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test panic on too many options
	t.Run("panic on too many options", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("GetOptDef() should panic with >1 option")
			}
		}()
		GetOptDef(42, []int{1, 2}...)
	})
}

func TestGetOptString(t *testing.T) {
	tests := []struct {
		name   string
		opts   []string
		want   string
		wantOk bool
	}{
		{
			name:   "no options",
			opts:   nil,
			want:   "",
			wantOk: false,
		},
		{
			name:   "valid string",
			opts:   []string{"hello"},
			want:   "hello",
			wantOk: true,
		},
		{
			name:   "empty string",
			opts:   []string{""},
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetOpt(tt.opts...)
			if got != tt.want {
				t.Errorf("GetOpt() = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetOpt() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestGetOptDefDuration(t *testing.T) {
	defaultTimeout := 30 * time.Second

	tests := []struct {
		name string
		opts []time.Duration
		want time.Duration
	}{
		{
			name: "no options - returns default",
			opts: nil,
			want: defaultTimeout,
		},
		{
			name: "valid duration - returns option",
			opts: []time.Duration{60 * time.Second},
			want: 60 * time.Second,
		},
		{
			name: "zero duration - returns default",
			opts: []time.Duration{0},
			want: defaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOptDef(defaultTimeout, tt.opts...); got != tt.want {
				t.Errorf("GetOptDef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOptBool(t *testing.T) {
	tests := []struct {
		name   string
		opts   []bool
		want   bool
		wantOk bool
	}{
		{
			name:   "no options",
			opts:   nil,
			want:   false,
			wantOk: false,
		},
		{
			name:   "true",
			opts:   []bool{true},
			want:   true,
			wantOk: true,
		},
		{
			name:   "false",
			opts:   []bool{false},
			want:   false,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetOpt(tt.opts...)
			if got != tt.want {
				t.Errorf("GetOpt() = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetOpt() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

// Test pointer types (nil-safe)
type Config struct {
	Timeout time.Duration
	Enabled bool
}

func TestGetOptPointer(t *testing.T) {
	cfg := &Config{Timeout: 10 * time.Second}

	tests := []struct {
		name   string
		opts   []*Config
		want   *Config
		wantOk bool
	}{
		{
			name:   "no options",
			opts:   nil,
			want:   nil,
			wantOk: false,
		},
		{
			name:   "valid pointer",
			opts:   []*Config{cfg},
			want:   cfg,
			wantOk: true,
		},
		{
			name:   "nil pointer",
			opts:   []*Config{nil},
			want:   nil,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetOpt(tt.opts...)
			if got != tt.want {
				t.Errorf("GetOpt() = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetOpt() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
