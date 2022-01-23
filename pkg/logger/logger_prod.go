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

//go:build !dev

package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// production log config.
func getLogger() *zap.Logger { //nolint:ireturn
	prod := zapcore.EncoderConfig{
		TimeKey:        timeKey,
		NameKey:        nameKey,
		MessageKey:     messageKey,
		CallerKey:      callerKey,
		LevelKey:       levelKey,
		StacktraceKey:  traceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	return zap.New(
		zapcore.NewCore(zapcore.NewJSONEncoder(prod), zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}
