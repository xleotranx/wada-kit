package kernel_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xleotranx/wada-kit/kernel"
	"github.com/xleotranx/wada-kit/kernel/kerneltest"
)

func TestEventsPreserveNilPayloads(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		checkEventPayloads(t, []error{nil, errors.New("reported"), (*kernel.PanicError)(nil), nil})
	})
	t.Run("any", func(t *testing.T) {
		checkEventPayloads(t, []any{nil, 0, false, "", (*int)(nil), nil})
	})
	t.Run("pointer", func(t *testing.T) {
		value := 42
		checkEventPayloads(t, []*int{nil, &value, nil})
	})
}

func checkEventPayloads[T any](t *testing.T, values []T) {
	t.Helper()
	p := plug("listener", nil)
	h := kerneltest.Start(t, kerneltest.Config{Lenient: true}, p.New)
	topic := kernel.NewTopic[T]("events.nil." + t.Name())
	var got []T
	kernel.On(p.ctx, topic, func(value T) { got = append(got, value) })
	for _, value := range values {
		kernel.Emit(h, topic, value)
	}
	if !reflect.DeepEqual(got, values) {
		t.Fatalf("received %#v, want %#v; status: %+v", got, values, state(t, h, "listener"))
	}
	if st := state(t, h, "listener"); st.State != kernel.Active {
		t.Fatalf("valid payload failed listener: %+v", st)
	}
	closeClean(t, h)
}
