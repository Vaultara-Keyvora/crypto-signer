package log

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type alignEncoder struct {
	zapcore.Encoder
}

// EncodeEntry is a custom method to wrap the original EncodeEntry and align the message.
func (ae *alignEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	return ae.Encoder.EncodeEntry(entry, fields)
}
