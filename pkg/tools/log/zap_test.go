package log

import (
	"context"
	"reflect"
	"testing"
)

func TestKVAppendAddsSafeContextFields(t *testing.T) {
	ctx := context.Background()
	for key, value := range map[string]string{
		"trace_id":    "trace-123",
		"operation":   "POST /tx/sign",
		"merchant_id": "merchant-1",
		"chain":       "tron",
	} {
		ctx = context.WithValue(ctx, key, value)
	}

	got := (&ZapLogger{}).kvAppend(ctx, []any{"stage", "transaction_signing"})
	want := []any{
		"stage", "transaction_signing",
		"trace_id", "trace-123",
		"operation", "POST /tx/sign",
		"merchant_id", "merchant-1",
		"chain", "tron",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("kvAppend() = %#v, want %#v", got, want)
	}
}

func TestKVAppendDoesNotDuplicateExplicitField(t *testing.T) {
	ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
	got := (&ZapLogger{}).kvAppend(ctx, []any{"trace_id", "override"})
	want := []any{"trace_id", "override"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("kvAppend() = %#v, want %#v", got, want)
	}
}
