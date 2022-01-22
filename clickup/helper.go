package clickup

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

func maxInt64(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func warnErrorIf(l *zap.Logger, err error, msg string, pairs ...interface{}) {
	if err != nil {
		l.With(zap.Error(err)).Sugar().Warnw(msg, pairs...)
	}
}

func warnIfFailedRequest(l *zap.Logger, res interface{ StatusOK() bool }) {
	if !res.StatusOK() {
		warnErrorIf(l, errors.New("failed request in ClickUp API"), fmt.Sprintf("response type %T (for detect kind of request)", res))
	}
}
