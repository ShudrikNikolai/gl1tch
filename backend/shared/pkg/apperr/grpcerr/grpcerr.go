package grpcerr

import (
	"context"
	"errors"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/apperr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Domain значение google.rpc.ErrorInfo.domain для ошибок проекта.
const Domain = "gl1tch"

// ToCode сопоставляет вид ошибки коду gRPC.
func ToCode(k apperr.Kind) codes.Code {
	switch k {
	case apperr.KindInvalidArgument:
		return codes.InvalidArgument
	case apperr.KindUnauthenticated:
		return codes.Unauthenticated
	case apperr.KindPermissionDenied:
		return codes.PermissionDenied
	case apperr.KindNotFound:
		return codes.NotFound
	case apperr.KindAlreadyExists:
		return codes.AlreadyExists
	case apperr.KindRateLimited:
		return codes.ResourceExhausted
	case apperr.KindUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}

func toKind(c codes.Code) apperr.Kind {
	switch c {
	case codes.InvalidArgument:
		return apperr.KindInvalidArgument
	case codes.Unauthenticated:
		return apperr.KindUnauthenticated
	case codes.PermissionDenied:
		return apperr.KindPermissionDenied
	case codes.NotFound:
		return apperr.KindNotFound
	case codes.AlreadyExists:
		return apperr.KindAlreadyExists
	case codes.ResourceExhausted:
		return apperr.KindRateLimited
	case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
		return apperr.KindUnavailable
	default:
		return apperr.KindInternal
	}
}

// ToStatus готовит ошибку обработчика к отправке клиенту. Причина (Err) и детали внутренних ошибок
// клиенту не уходят: для неизвестных ошибок возвращается Internal с нейтральным текстом.
func ToStatus(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "request canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	}

	e := apperr.From(err)
	st := status.New(ToCode(e.Kind), e.Message)
	if withDetails, derr := st.WithDetails(&errdetails.ErrorInfo{Reason: e.Code, Domain: Domain}); derr == nil {
		st = withDetails
	}

	return st.Err()
}

// FromStatus превращает ошибку, полученную от gRPC-сервера, в *apperr.Error. Ошибки, не являющиеся статусом gRPC,
// возвращаются без изменений. Тексты сообщений Internal не пробрасываются дальше.
func FromStatus(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	kind := toKind(st.Code())
	code := apperr.DefaultCode(kind)
	for _, d := range st.Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok && info.GetDomain() == Domain && info.GetReason() != "" {
			code = info.GetReason()
			break
		}
	}

	msg := st.Message()
	if kind == apperr.KindInternal {
		msg = "internal error"
	}

	return apperr.Wrap(kind, code, msg, err)
}
