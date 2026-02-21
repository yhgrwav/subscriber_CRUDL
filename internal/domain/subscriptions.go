// В subscriptions.go я буду пробрасывать интерфейс, описывать сущность, с которой буду взаимодействовать,
// создам кастомные ошибки, возможно допишу какие-то простые проверки в две строки сюда
package domain

import (
	"github.com/google/uuid"
)

type Subscription struct {
	ID          int       `json:"id" db:"id"`                     // Идентификатор подписки
	ServiceName string    `json:"service_name" db:"service_name"` // Название покупаемого сервиса
	Price       int       `json:"price" db:"price"`               // Цена подписки
	UserID      uuid.UUID `json:"user_id" db:"user_id"`           // UUID пользователя

	// Дату реализовал через строку по примеру из ТЗ, планирую строку валидировать и преобразовывать с помощью time.Parse
	// в сервисе, в базу буду сохранять в TIMESTAMP.
	StartDate string  `json:"start_date" db:"start_date"`       // Дата активации подписки
	EndDate   *string `json:"end_date,omitempty" db:"end_date"` // Дата окончания подписки, EndDate реализовал через указатель на строку для проверки на nil,

}
