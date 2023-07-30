package metadata

import (
	"context"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		mds []map[string]string
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "hello",
			args: args{[]map[string]string{{"hello": "mio"}, {"hello2": "go-mio"}}},
			want: Metadata{"hello": "mio", "hello2": "go-mio"},
		},
		{
			name: "hi",
			args: args{[]map[string]string{{"hi": "mio"}, {"hi2": "go-mio"}}},
			want: Metadata{"hi": "mio", "hi2": "go-mio"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.mds...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetadata_Get(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		m    Metadata
		args args
		want string
	}{
		{
			name: "mio",
			m:    Metadata{"mio": "value", "env": "dev"},
			args: args{key: "mio"},
			want: "value",
		},
		{
			name: "env",
			m:    Metadata{"mio": "value", "env": "dev"},
			args: args{key: "env"},
			want: "dev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Get(tt.args.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetadata_Set(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		m    Metadata
		args args
		want Metadata
	}{
		{
			name: "mio",
			m:    Metadata{},
			args: args{key: "hello", value: "mio"},
			want: Metadata{"hello": "mio"},
		},
		{
			name: "env",
			m:    Metadata{"hello": "mio"},
			args: args{key: "env", value: "pro"},
			want: Metadata{"hello": "mio", "env": "pro"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.Set(tt.args.key, tt.args.value)
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("Set() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestClientContext(t *testing.T) {
	type args struct {
		ctx context.Context
		md  Metadata
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "mio",
			args: args{context.Background(), Metadata{"hello": "mio", "mio": "https://go-mio.dev"}},
		},
		{
			name: "hello",
			args: args{context.Background(), Metadata{"hello": "mio", "hello2": "https://go-mio.dev"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewClientContext(tt.args.ctx, tt.args.md)
			m, ok := FromClientContext(ctx)
			if !ok {
				t.Errorf("FromClientContext() = %v, want %v", ok, true)
			}

			if !reflect.DeepEqual(m, tt.args.md) {
				t.Errorf("meta = %v, want %v", m, tt.args.md)
			}
		})
	}
}

func TestServerContext(t *testing.T) {
	type args struct {
		ctx context.Context
		md  Metadata
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "mio",
			args: args{context.Background(), Metadata{"hello": "mio", "mio": "https://go-mio.dev"}},
		},
		{
			name: "hello",
			args: args{context.Background(), Metadata{"hello": "mio", "hello2": "https://go-mio.dev"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewServerContext(tt.args.ctx, tt.args.md)
			m, ok := FromServerContext(ctx)
			if !ok {
				t.Errorf("FromServerContext() = %v, want %v", ok, true)
			}

			if !reflect.DeepEqual(m, tt.args.md) {
				t.Errorf("meta = %v, want %v", m, tt.args.md)
			}
		})
	}
}

func TestAppendToClientContext(t *testing.T) {
	type args struct {
		md Metadata
		kv []string
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "mio",
			args: args{Metadata{}, []string{"hello", "mio", "env", "dev"}},
			want: Metadata{"hello": "mio", "env": "dev"},
		},
		{
			name: "hello",
			args: args{Metadata{"hi": "https://go-mio.dev/"}, []string{"hello", "mio", "env", "dev"}},
			want: Metadata{"hello": "mio", "env": "dev", "hi": "https://go-mio.dev/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewClientContext(context.Background(), tt.args.md)
			ctx = AppendToClientContext(ctx, tt.args.kv...)
			md, ok := FromClientContext(ctx)
			if !ok {
				t.Errorf("FromServerContext() = %v, want %v", ok, true)
			}
			if !reflect.DeepEqual(md, tt.want) {
				t.Errorf("metadata = %v, want %v", md, tt.want)
			}
		})
	}
}

func TestMergeToClientContext(t *testing.T) {
	type args struct {
		md       Metadata
		appendMd Metadata
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "mio",
			args: args{Metadata{}, Metadata{"hello": "mio", "env": "dev"}},
			want: Metadata{"hello": "mio", "env": "dev"},
		},
		{
			name: "hello",
			args: args{Metadata{"hi": "https://go-mio.dev/"}, Metadata{"hello": "mio", "env": "dev"}},
			want: Metadata{"hello": "mio", "env": "dev", "hi": "https://go-mio.dev/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewClientContext(context.Background(), tt.args.md)
			ctx = MergeToClientContext(ctx, tt.args.appendMd)
			md, ok := FromClientContext(ctx)
			if !ok {
				t.Errorf("FromServerContext() = %v, want %v", ok, true)
			}
			if !reflect.DeepEqual(md, tt.want) {
				t.Errorf("metadata = %v, want %v", md, tt.want)
			}
		})
	}
}

func TestMetadata_Range(t *testing.T) {
	md := Metadata{"mio": "mio", "https://go-mio.dev/": "https://go-mio.dev/", "go-mio": "go-mio"}
	tmp := Metadata{}
	md.Range(func(k, v string) bool {
		if k == "https://go-mio.dev/" || k == "mio" {
			tmp[k] = v
		}
		return true
	})
	if !reflect.DeepEqual(tmp, Metadata{"https://go-mio.dev/": "https://go-mio.dev/", "mio": "mio"}) {
		t.Errorf("metadata = %v, want %v", tmp, Metadata{"mio": "mio"})
	}
}

func TestMetadata_Clone(t *testing.T) {
	tests := []struct {
		name string
		m    Metadata
		want Metadata
	}{
		{
			name: "mio",
			m:    Metadata{"mio": "mio", "https://go-mio.dev/": "https://go-mio.dev/", "go-mio": "go-mio"},
			want: Metadata{"mio": "mio", "https://go-mio.dev/": "https://go-mio.dev/", "go-mio": "go-mio"},
		},
		{
			name: "go",
			m:    Metadata{"language": "golang"},
			want: Metadata{"language": "golang"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Clone()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clone() = %v, want %v", got, tt.want)
			}
			got["mio"] = "go"
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("want got != want got %v want %v", got, tt.want)
			}
		})
	}
}
