package errors

import (
	"errors"
	"testing"
)

func TestAppErrorUnwrap(t *testing.T) {
	cause := ErrNotLoggedIn
	appErr := NewAppError("LOGIN_001", "操作失败", cause)

	if !errors.Is(appErr, ErrNotLoggedIn) {
		t.Error("errors.Is 应能穿透 AppError 找到 cause")
	}
}

func TestAppErrorNilCause(t *testing.T) {
	appErr := NewAppError("EXPORT_001", "导出失败", nil)
	if appErr.Unwrap() != nil {
		t.Error("Unwrap() 应返回 nil")
	}
}

func TestAppErrorMessage(t *testing.T) {
	appErr := NewAppError("CFG_001", "配置无效", nil)
	if appErr.Error() != "配置无效" {
		t.Errorf("Error() = %q, want %q", appErr.Error(), "配置无效")
	}
}
