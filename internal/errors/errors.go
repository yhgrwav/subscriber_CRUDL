package errors

import "errors"

// здесь я буду реализовывать кастомные ошибки для общего развития
var (
	ErrSubscriptionNotFound = errors.New("подписка не найдена")
	ErrInvalidDateFormat    = errors.New("указан невалидный формат даты")
	ErrInvalidPrice         = errors.New("указана невалидная цена")
	ErrInvalidUserID        = errors.New("пользователя не существует")

	// можно было бы расписать еще кучу ошибок, если бы у меня была условная база юзеров и сервисов, но есть что есть
)
