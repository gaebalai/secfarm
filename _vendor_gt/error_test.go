package gt_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gaebalai/gt"
)

func TestError(t *testing.T) {
	err := errors.New("test")
	cnt := newRecorder()
	gt.Error(cnt, err)

	if cnt.errs > 0 {
		t.Error("error test has unexpected result")
	}
}

type testError struct{ N int }

func (x testError) Error() string { return "test error" }

func TestErrorAs(t *testing.T) {
	testErr := testError{N: 5}
	err := fmt.Errorf("run error: %w", testErr)

	var called int
	gt.ErrorAs(t, err, func(tgt *testError) {
		called++
		if tgt.N != 5 {
			t.Errorf("testError.N must be 5, but %+v", tgt.N)
		}
	})
	if called != 1 {
		t.Errorf("callback must be called once, but %+v times", called)
	}
}

func TestErrorContains(t *testing.T) {
	t.Run("error contains expected string", func(t *testing.T) {
		err := errors.New("test error message")
		cnt := newRecorder()
		gt.Error(cnt, err).Contains("error message")
		if cnt.errs > 0 {
			t.Error("error test has unexpected result")
		}
	})

	t.Run("error does not contain expected string", func(t *testing.T) {
		err := errors.New("test error message")
		cnt := newRecorder()
		gt.Error(cnt, err).Contains("unexpected message")
		if cnt.errs == 0 {
			t.Error("error test should report error when message does not contain expected string")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		cnt := newRecorder()
		// Error() 함수 자체가 nil 에러인 경우 에러를 보고하므로,
		// 초기 에러 수를 기록
		initialErrs := cnt.errs
		gt.Error(cnt, nil).Contains("any message")
		// 에러 수가 증가했는지 확인
		if cnt.errs <= initialErrs {
			t.Error("error test should report additional error for Contains check")
		}
	})
}
