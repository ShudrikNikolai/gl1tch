package grpcerr

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestKindCodeRoundTrip(t *testing.T) {
	kinds := []apperr.Kind{
		apperr.KindInvalidArgument, apperr.KindUnauthenticated, apperr.KindPermissionDenied,
		apperr.KindNotFound, apperr.KindAlreadyExists, apperr.KindRateLimited, apperr.KindUnavailable, apperr.KindInternal,
	}
	for _, k := range kinds {
		err := FromStatus(ToStatus(apperr.New(k, "SOME_CODE", "msg")))
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Kind != k {
			t.Errorf("kind %d did not survive the round trip: %#v", k, err)
		}
	}
}

func TestContextErrors(t *testing.T) {
	if status.Code(ToStatus(context.Canceled)) != codes.Canceled {
		t.Fatal("canceled")
	}
	if status.Code(ToStatus(fmt.Errorf("x: %w", context.DeadlineExceeded))) != codes.DeadlineExceeded {
		t.Fatal("deadline")
	}
	if ToStatus(nil) != nil || FromStatus(nil) != nil {
		t.Fatal("nil must stay nil")
	}
}

func TestFromStatusWithoutDetailsUsesDefaultCode(t *testing.T) {
	err := FromStatus(status.Error(codes.NotFound, "no such thing"))
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeNotFound || ae.Message != "no such thing" {
		t.Fatalf("%#v", err)
	}
}

func TestFromStatusPassesThroughNonStatusErrors(t *testing.T) {
	plain := errors.New("not a grpc error")
	if got := FromStatus(plain); got == nil {
		t.Fatal("non-status errors are returned as is")
	}
}
