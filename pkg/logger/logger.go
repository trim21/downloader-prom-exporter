// Copyright (c) 2021-2022 Trim21 <trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.

// Package logger config a zap logger, functions have same signature with `zap.logger`.
// Can be configured by env `LOG_LEVEL` and `PRODUCTION`.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log = setup()

const (
	timeKey    = "time"
	nameKey    = "logger"
	messageKey = "msg"
	callerKey  = "caller"
	levelKey   = "level"
	traceKey   = "trace"
)

func setup() *zap.Logger {
	return getLogger()
}

// Named create a named logger.
func Named(name string) *zap.Logger {
	return log.Named(name).WithOptions(zap.AddCallerSkip(-1))
}

// Debug level logging.
func Debug(msg string, fields ...zapcore.Field) {
	log.Debug(msg, fields...)
}

// Info level logging.
func Info(msg string, fields ...zapcore.Field) {
	log.Info(msg, fields...)
}

// Warn level logging.
func Warn(msg string, fields ...zapcore.Field) {
	log.Warn(msg, fields...)
}

// Error level logging.
func Error(msg string, fields ...zapcore.Field) {
	log.Error(msg, fields...)
}

// WithE is a shortcut for `logger.With(zap.Error(err))`.
func WithE(err error) *zap.Logger {
	return log.With(zap.Error(err)).WithOptions(zap.AddCallerSkip(-1))
}

// With return a logger with common fields.
func With(fields ...zapcore.Field) *zap.Logger {
	return log.With(fields...)
}

// DPanic will panic in development and log error message at production env.
func DPanic(msg string, fields ...zapcore.Field) {
	log.DPanic(msg, fields...)
}

// Panic will log at panic level then panic.
func Panic(msg string, fields ...zapcore.Field) {
	log.Panic(msg, fields...)
}

// Fatal level message and call `os.Exit(1)`.
func Fatal(msg string, fields ...zapcore.Field) {
	log.Fatal(msg, fields...)
}

// Sugar return a zap.SugaredLogger.
func Sugar() *zap.SugaredLogger {
	return log.WithOptions(zap.AddCallerSkip(-1)).Sugar()
}
