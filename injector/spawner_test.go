package injector

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpawnerFromFactoryFunc(t *testing.T) {
	factoryFunc := func() (fakeInterface, error) { return nil, nil }

	tests := []struct {
		name        string
		factoryFunc interface{}
		wantErr     bool
	}{
		{
			name:        "Valid factory function",
			factoryFunc: factoryFunc,
		},
		{
			name:        "Valid factory function pointer",
			factoryFunc: &factoryFunc,
		},
		{
			name:        "Invalid factory function",
			factoryFunc: func() {},
			wantErr:     true,
		},
		{
			name:        "Non-function value as factory function",
			factoryFunc: struct{}{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spawner, err := SpawnerFromFactoryFunc(tt.factoryFunc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				if assert.NoError(t, err) {
					assert.NotNil(t, spawner)
				}
			}
		})
	}
}

func TestSpawner_Get(t *testing.T) {
	tests := []struct {
		name        string
		s           Spawner
		cfgProvider CfgProvider
		want        *reflect.Value
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Get(tt.cfgProvider)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				if assert.NoError(t, err) {
					assert.NotNil(t, got)
				}
			}
		})
	}
}

func TestSpawner_ReturnsErr(t *testing.T) {
	tests := []struct {
		name string
		s    Spawner
		want bool
	}{
		{
			name: "Returns error",
			s:    spawnerFrom(func() (fakeInterface, error) { return nil, nil }),
			want: true,
		},
		{
			name: "Does not return error",
			s:    spawnerFrom(func() fakeInterface { return nil }),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.ReturnsErr())
		})
	}
}

func TestSpawner_Type(t *testing.T) {
	tests := []struct {
		name string
		s    Spawner
		want reflect.Type
	}{
		{
			name: "Basic test",
			s:    spawnerFrom(func() fakeInterface { return nil }),
			want: reflect.TypeOf((*fakeInterface)(nil)).Elem(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.Type())
		})
	}
}

func TestSpawner_Validate(t *testing.T) {
	tests := []struct {
		name    string
		s       Spawner
		wantErr bool
	}{
		{
			name: "Valid spawner returns single value",
			s:    spawnerFrom(func() fakeInterface { return nil }),
		},
		{
			name: "Valid spawner returns value and error",
			s:    spawnerFrom(func() (fakeInterface, error) { return nil, nil }),
		},
		{
			name: "Valid spawner has args",
			s:    spawnerFrom(func(*fakeConfig, *fakeConfig) fakeInterface { return nil }),
		},
		{
			name:    "Spawner from non-function value",
			s:       spawnerFrom("foo"),
			wantErr: true,
		},
		{
			name:    "Invalid spawner returns no value",
			s:       spawnerFrom(func() {}),
			wantErr: true,
		},
		{
			name:    "Invalid spawner returns non-error type as 2nd value",
			s:       spawnerFrom(func() (fakeInterface, fakeInterface) { return nil, nil }),
			wantErr: true,
		},
		{
			name:    "Invalid spawner returns more than 2 values",
			s:       spawnerFrom(func() (fakeInterface, error, fakeInterface) { return nil, nil, nil }),
			wantErr: true,
		},
		{
			name:    "Invalid spawner returns non-interface",
			s:       spawnerFrom(func() (*struct{}, error) { return &struct{}{}, nil }),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Error(t, tt.s.Validate())
			} else {
				assert.NoError(t, tt.s.Validate())
			}
		})
	}
}

func spawnerFrom(value interface{}) Spawner {
	return Spawner(reflect.ValueOf(value))
}

type (
	fakeInterface interface{}
	fakeConfig    struct{}
)
