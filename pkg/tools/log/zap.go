// Copyright © 2025 Little Bee. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/lestrrat-go/file-rotatelogs"
)

type LogFormatter interface {
	Format() any
}

type Logger interface {
	// Debug logs a message at the debug level including any supplementary key-value pairs.
	// Useful for detailed output for debugging purposes.
	Debug(ctx context.Context, msg string, keysAndValues ...any)

	// Info logs a message at the info level along with any supplementary key-value pairs.
	// Ideal for general operational messages that inform about the system's state.
	Info(ctx context.Context, msg string, keysAndValues ...any)

	// Warn logs a message at the warning level, indicating potential issues in the system.
	// It includes an error object and any supplementary key-value pairs.
	Warn(ctx context.Context, msg string, err error, keysAndValues ...any)

	// Error logs a message at the error level, indicating serious problems that need attention.
	// It includes an error object and any supplementary key-value pairs.
	Error(ctx context.Context, msg string, err error, keysAndValues ...any)

	// WithValues returns a new Logger instance that will include the specified key-value pairs
	// in all subsequent log messages. Useful for adding consistent context to a series of logs.
	WithValues(keysAndValues ...any) Logger

	// WithName returns a new Logger instance prefixed with the specified name.
	// This is helpful for distinguishing logs generated from different sources or components.
	WithName(name string) Logger

	// WithCallDepth returns a new Logger instance that adjusts the call depth for identifying
	// the source of log messages. Useful in wrapper or middleware layers to maintain accurate log source information.
	WithCallDepth(depth int) Logger
}

var (
	pkgLogger   Logger
	osStdout    Logger
	sp          = string(filepath.Separator)
	logLevelMap = map[int]zapcore.Level{
		6: zapcore.DebugLevel,
		5: zapcore.DebugLevel,
		4: zapcore.InfoLevel,
		3: zapcore.WarnLevel,
		2: zapcore.ErrorLevel,
		1: zapcore.FatalLevel,
		0: zapcore.PanicLevel,
	}
)

const callDepth = 2
const hoursPerDay = 24

// InitFromConfig initializes a Zap-based logger.
func InitFromConfig(
	loggerPrefixName, moduleName string,
	logLevel int,
	isStdout bool,
	isJson bool,
	logLocation string,
	rotateCount uint,
	rotationTime uint,
	moduleVersion string,
	isSimplify bool,
) error {
	l, err := NewZapLogger(loggerPrefixName, moduleName, logLevel, isStdout, isJson, logLocation,
		rotateCount, rotationTime, moduleVersion, isSimplify)
	if err != nil {
		return err
	}
	pkgLogger = l.WithCallDepth(callDepth)
	if isJson {
		pkgLogger = pkgLogger.WithName(moduleName)
	}
	return nil
}

// InitConsoleLogger init osStdout and osStderr.
func InitConsoleLogger(moduleName string,
	logLevel int,
	isJson bool, moduleVersion string) error {
	l, err := NewConsoleZapLogger(moduleName, logLevel, isJson, moduleVersion, os.Stdout)
	if err != nil {
		return err
	}
	osStdout = l.WithCallDepth(callDepth)
	if isJson {
		osStdout = osStdout.WithName(moduleName)
	}
	return nil

}

func ZDebug(ctx context.Context, msg string, keysAndValues ...any) {
	if pkgLogger == nil {
		return
	}
	pkgLogger.Debug(ctx, msg, keysAndValues...)
}

func ZInfo(ctx context.Context, msg string, keysAndValues ...any) {
	if pkgLogger == nil {
		return
	}
	pkgLogger.Info(ctx, msg, keysAndValues...)
}

func ZWarn(ctx context.Context, msg string, err error, keysAndValues ...any) {
	if pkgLogger == nil {
		return
	}
	pkgLogger.Warn(ctx, msg, err, keysAndValues...)
}

func ZError(ctx context.Context, msg string, err error, keysAndValues ...any) {
	if pkgLogger == nil {
		return
	}
	pkgLogger.Error(ctx, msg, err, keysAndValues...)
}

//func CInfo(ctx context.Context, persister string, keysAndValues ...any) {
//	if osStdout == nil {
//		return
//	}
//	osStdout.Info(ctx, persister, keysAndValues...)
//}

type ZapLogger struct {
	zap              *zap.SugaredLogger
	level            zapcore.Level
	moduleName       string
	moduleVersion    string
	loggerPrefixName string
	rotationTime     time.Duration
	isSimplify       bool
}

func NewZapLogger(
	loggerPrefixName, moduleName string,
	logLevel int,
	isStdout bool,
	isJson bool,
	logLocation string,
	rotateCount uint,
	rotationTime uint,
	moduleVersion string,
	isSimplify bool,
) (*ZapLogger, error) {
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(logLevelMap[logLevel]),
		DisableStacktrace: true,
	}
	if isJson {
		zapConfig.Encoding = "json"
	} else {
		zapConfig.Encoding = "console"
	}
	zl := &ZapLogger{level: logLevelMap[logLevel], moduleName: moduleName, loggerPrefixName: loggerPrefixName,
		rotationTime: time.Duration(rotationTime) * time.Hour, moduleVersion: moduleVersion, isSimplify: isSimplify}
	opts, err := zl.cores(isStdout, isJson, logLocation, rotateCount)
	if err != nil {
		return nil, err
	}
	l, err := zapConfig.Build(opts)
	if err != nil {
		return nil, err
	}
	zl.zap = l.Sugar()
	return zl, nil
}

func NewConsoleZapLogger(
	moduleName string,
	logLevel int,
	isJson bool,
	moduleVersion string,
	outPut *os.File) (*ZapLogger, error) {
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(logLevelMap[logLevel]),
		DisableStacktrace: true,
	}
	if isJson {
		zapConfig.Encoding = "json"
	} else {
		zapConfig.Encoding = "console"
	}
	zl := &ZapLogger{level: logLevelMap[logLevel], moduleName: moduleName, moduleVersion: moduleVersion}
	opts, err := zl.consoleCores(outPut, isJson)
	if err != nil {
		return nil, err
	}
	l, err := zapConfig.Build(opts)
	if err != nil {
		return nil, err
	}
	zl.zap = l.Sugar()
	return zl, nil
}
func (l *ZapLogger) cores(isStdout bool, isJson bool, logLocation string, rotateCount uint) (zap.Option, error) {
	c := zap.NewProductionEncoderConfig()
	c.EncodeTime = l.timeEncoder
	c.EncodeDuration = zapcore.SecondsDurationEncoder
	c.MessageKey = "persister"
	c.LevelKey = "level"
	c.TimeKey = "time"
	c.CallerKey = "caller"
	c.NameKey = "logger"
	var fileEncoder zapcore.Encoder
	if isJson {
		c.EncodeLevel = zapcore.CapitalLevelEncoder
		fileEncoder = zapcore.NewJSONEncoder(c)
		fileEncoder.AddInt("PID", os.Getpid())
		fileEncoder.AddString("version", l.moduleVersion)
	} else {
		c.EncodeLevel = l.capitalLevelEncoder
		c.EncodeCaller = l.customCallerEncoder
		fileEncoder = zapcore.NewConsoleEncoder(c)
	}
	fileEncoder = &alignEncoder{Encoder: fileEncoder}
	writer, err := l.getWriter(logLocation, rotateCount)
	if err != nil {
		return nil, err
	}
	var cores []zapcore.Core
	if logLocation != "" {
		cores = []zapcore.Core{
			zapcore.NewCore(fileEncoder, writer, zap.NewAtomicLevelAt(l.level)),
		}
	}
	if isStdout {
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.Lock(os.Stdout), zap.NewAtomicLevelAt(l.level)))
	}
	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	}), nil
}

func (l *ZapLogger) consoleCores(outPut *os.File, isJson bool) (zap.Option, error) {
	c := zap.NewProductionEncoderConfig()
	c.EncodeTime = l.timeEncoder
	c.EncodeDuration = zapcore.SecondsDurationEncoder
	c.MessageKey = "persister"
	c.LevelKey = "level"
	c.TimeKey = "time"
	c.CallerKey = "caller"
	c.NameKey = "logger"
	var fileEncoder zapcore.Encoder
	if isJson {
		c.EncodeLevel = zapcore.CapitalLevelEncoder
		fileEncoder = zapcore.NewJSONEncoder(c)
		fileEncoder.AddInt("PID", os.Getpid())
		fileEncoder.AddString("version", l.moduleVersion)
	} else {
		c.EncodeLevel = l.capitalColorLevelEncoder
		c.EncodeCaller = l.customCallerEncoder
		fileEncoder = zapcore.NewConsoleEncoder(c)
	}
	var cores []zapcore.Core
	cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.Lock(outPut), zap.NewAtomicLevelAt(l.level)))

	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	}), nil
}

func (l *ZapLogger) customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	fixedLength := 50
	trimmedPath := caller.TrimmedPath()
	trimmedPath = "[" + trimmedPath + "]"
	s := FormatString(trimmedPath, fixedLength, true)
	enc.AppendString(s)
}

func (l *ZapLogger) timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	layout := "2006-01-02 15:04:05.000"
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}
	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}
	enc.AppendString(t.Format(layout))
}

func (l *ZapLogger) getWriter(logLocation string, rorateCount uint) (zapcore.WriteSyncer, error) {
	var path string
	if l.rotationTime%(time.Hour*time.Duration(hoursPerDay)) == 0 {
		path = logLocation + sp + l.loggerPrefixName + ".%Y-%m-%d.log"
	} else if l.rotationTime%time.Hour == 0 {
		path = logLocation + sp + l.loggerPrefixName + ".%Y-%m-%d_%H.log"
	} else {
		path = logLocation + sp + l.loggerPrefixName + ".%Y-%m-%d_%H_%M_%S.log"
	}
	logf, err := rotatelogs.New(path,
		rotatelogs.WithRotationCount(rorateCount),
		rotatelogs.WithRotationTime(l.rotationTime),
	)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(logf), nil
}

// Ensure no color for files
func (l *ZapLogger) capitalLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := _levelToCapitalString[level]
	if !ok {
		s = _unknownLevelColor[zapcore.ErrorLevel]
	}
	pid := FormatString(fmt.Sprintf("["+"PID:"+"%d"+"]", os.Getpid()), 15, true)
	enc.AppendString(s)
	enc.AppendString(pid)
	if l.moduleName != "" {
		moduleName := FormatString(l.moduleName, 25, true)
		enc.AppendString(moduleName)
	}
	if l.moduleVersion != "" {
		moduleVersion := FormatString(fmt.Sprintf("["+"version:"+"%s"+"]", l.moduleVersion), 20, true)
		enc.AppendString(moduleVersion)
	}
}

// Enable color for console
func (l *ZapLogger) capitalColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := _levelToCapitalColorString[level]
	if !ok {
		s = _unknownLevelColor[zapcore.ErrorLevel]
	}
	pid := FormatString(fmt.Sprintf("["+"PID:"+"%d"+"]", os.Getpid()), 15, true)
	color := _levelToColor[level]
	enc.AppendString(s)
	enc.AppendString(color.Add(pid))
	if l.moduleName != "" {
		moduleName := FormatString(l.moduleName, 25, true)
		enc.AppendString(color.Add(moduleName))
	}
	if l.moduleVersion != "" {
		moduleVersion := FormatString(fmt.Sprintf("["+"version:"+"%s"+"]", l.moduleVersion), 20, true)
		enc.AppendString(moduleVersion)
	}
}

func (l *ZapLogger) ToZap() *zap.SugaredLogger {
	return l.zap
}

func (l *ZapLogger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	if l.level > zapcore.DebugLevel {
		return
	}
	keysAndValues = l.kvAppend(ctx, keysAndValues)
	l.zap.Debugw(msg, keysAndValues...)
}

func (l *ZapLogger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	if l.level > zapcore.InfoLevel {
		return
	}
	keysAndValues = l.kvAppend(ctx, keysAndValues)
	l.zap.Infow(msg, keysAndValues...)
}

func (l *ZapLogger) Warn(ctx context.Context, msg string, err error, keysAndValues ...any) {
	if l.level > zapcore.WarnLevel {
		return
	}
	if err != nil {
		keysAndValues = append(keysAndValues, "error", err.Error())
	}
	keysAndValues = l.kvAppend(ctx, keysAndValues)
	l.zap.Warnw(msg, keysAndValues...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, err error, keysAndValues ...any) {
	if l.level > zapcore.ErrorLevel {
		return
	}
	if err != nil {
		keysAndValues = append(keysAndValues, "error", err.Error())
	}
	keysAndValues = l.kvAppend(ctx, keysAndValues)
	l.zap.Errorw(msg, keysAndValues...)
}

func (l *ZapLogger) kvAppend(ctx context.Context, keysAndValues []any) []any {
	if ctx == nil {
		return keysAndValues
	}
	for _, key := range []string{"trace_id", "operation", "merchant_id", "chain"} {
		if hasLogField(keysAndValues, key) {
			continue
		}
		if value, ok := ctx.Value(key).(string); ok && value != "" {
			keysAndValues = append(keysAndValues, key, value)
		}
	}
	return keysAndValues
}

func hasLogField(keysAndValues []any, key string) bool {
	for i := 0; i < len(keysAndValues); {
		switch value := keysAndValues[i].(type) {
		case zap.Field:
			if value.Key == key {
				return true
			}
			i++
		case string:
			if value == key {
				return true
			}
			if i+1 < len(keysAndValues) {
				i += 2
			} else {
				i++
			}
		default:
			i++
		}
	}
	return false
}

func (l *ZapLogger) WithValues(keysAndValues ...any) Logger {
	dup := *l
	dup.zap = l.zap.With(keysAndValues...)
	return &dup
}

func (l *ZapLogger) WithName(name string) Logger {
	dup := *l
	dup.zap = l.zap.Named(name)
	return &dup
}

func (l *ZapLogger) WithCallDepth(depth int) Logger {
	dup := *l
	dup.zap = l.zap.WithOptions(zap.AddCallerSkip(depth))
	return &dup
}
