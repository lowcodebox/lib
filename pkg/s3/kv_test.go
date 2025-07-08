package s3_test

import (
	"fmt"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestLocalKVStore_BasicOperations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		initial    map[string]string
		putKey     string
		putVal     string
		checkKey   string
		wantCheck  bool
		getKey     string
		wantGetVal string
		wantGetErr bool
	}{
		{
			name:       "missing key",
			initial:    nil,
			checkKey:   "foo",
			wantCheck:  false,
			getKey:     "foo",
			wantGetErr: true,
		},
		{
			name:       "basic put-get-check",
			initial:    nil,
			putKey:     "foo",
			putVal:     "bar",
			checkKey:   "foo",
			wantCheck:  true,
			getKey:     "foo",
			wantGetVal: "bar",
			wantGetErr: false,
		},
		{
			name:       "overwrite existing key",
			initial:    map[string]string{"dup": "first"},
			putKey:     "dup",
			putVal:     "second",
			checkKey:   "dup",
			wantCheck:  true,
			getKey:     "dup",
			wantGetVal: "second",
			wantGetErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := s3.NewLocalKVStore()

			// инициализируем initial
			for k, v := range tt.initial {
				err := store.Put(ctx, k, v)
				assert.NoError(t, err, "initial Put failed for key %q", k)
			}

			// выполняем Put, если нужно
			if tt.putKey != "" {
				err := store.Put(ctx, tt.putKey, tt.putVal)
				assert.NoError(t, err, "Put failed for key %q", tt.putKey)
			}

			// проверяем Check
			ok, err := store.Check(ctx, tt.checkKey)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCheck, ok, "unexpected Check result for key %q", tt.checkKey)

			// проверяем Get
			val, err := store.Get(ctx, tt.getKey)
			if tt.wantGetErr {
				assert.Error(t, err, "expected Get error for key %q", tt.getKey)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantGetVal, val, "unexpected value for key %q", tt.getKey)
			}
		})
	}
}

func TestLocalKVStore_WithInitialData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		initial map[string]string
	}{
		{"empty initialization", map[string]string{}},
		{"single key", map[string]string{"a": "1"}},
		{"multiple keys", map[string]string{"a": "1", "b": "2", "c": "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := s3.NewLocalKVStore()
			err := s3.InitializeWithMap(ctx, store, tt.initial)
			assert.NoError(t, err, "InitializeWithMap should not fail")

			for k, want := range tt.initial {
				ok, err := store.Check(ctx, k)
				assert.NoError(t, err)
				assert.True(t, ok, "expected key %q to be present", k)

				val, err := store.Get(ctx, k)
				assert.NoError(t, err)
				assert.Equal(t, want, val, "unexpected value for key %q", k)
			}
		})
	}
}

func TestLocalKVStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	const n = 100
	store := s3.NewLocalKVStore()
	var wg sync.WaitGroup

	// параллельные записи
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			val := fmt.Sprintf("val-%d", i)
			err := store.Put(ctx, key, val)
			assert.NoError(t, err, "Put should succeed for key %q", key)
		}(i)
	}
	wg.Wait()

	// параллельные чтения и проверки
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			want := fmt.Sprintf("val-%d", i)

			ok, err := store.Check(ctx, key)
			assert.NoError(t, err)
			assert.True(t, ok, "expected key %q to exist", key)

			got, err := store.Get(ctx, key)
			assert.NoError(t, err)
			assert.Equal(t, want, got, "unexpected value for key %q", key)
		}(i)
	}
	wg.Wait()
}
