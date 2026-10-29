package middleware

// Рекомендуемый порядок:
//
//	consumer.NewConsumer(group, topics, log,
//	    middleware.DLQ(dlqProducer, log),          // после исчерпания повторов сообщение уходит в <group>.dlq
//	    middleware.Retry(cfg, log),                // повторы с экспоненциальной паузой
//	    middleware.Dedupe(redis, "inbox:", ttl, idOf, log), // пропуск уже обработанных event_id
//	    middleware.Recover(log),                   // паника превращается в ошибку
//	)
//
// Семантика доставки: at-least-once.
//
import "errors"

type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

// Permanent помечает ошибку как не требующую повторов (например, сообщение невозможно разобрать).
// Retry сразу отдаёт такую ошибку дальше, и DLQ забирает сообщение без пауз.
func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return permanentError{err}
}

// IsPermanent сообщает, помечена ли ошибка через Permanent.
func IsPermanent(err error) bool {
	var p permanentError
	return errors.As(err, &p)
}
