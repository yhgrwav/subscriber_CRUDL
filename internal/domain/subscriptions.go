// В subscriptions.go я буду пробрасывать интерфейс, описывать сущность, с которой буду взаимодействовать,
// создам кастомные ошибки, возможно допишу какие-то простые проверки в две строки сюда
package domain

import (
	"context"

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

// CRUDL методы для репозитория
type SubscriptionRepository interface {
	Create(ctx context.Context, sub Subscription) (Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	Update(ctx context.Context, sub Subscription) (Subscription, error)
	Delete(ctx context.Context, id int) error
	GetStatsByServiceName(ctx context.Context, serviceName string) ([]Subscription, error)
}

// CRUDL Методы для сервиса
type SubscriptionService interface {
	Create(ctx context.Context, sub Subscription) (*Subscription, error)
	Read(ctx context.Context, sub Subscription) (Subscription, error)
	Update(ctx context.Context, sub Subscription) (*Subscription, error)
	Delete(ctx context.Context, id int) error
	GetListByUserID(ctx context.Context, UserID uuid.UUID) ([]Subscription, error)

	//Втрой пункт ТЗ
	//ручка "для подсчета суммарной стоимости всех подписок за
	//выбранный период с фильтрацией по id пользователя и названию подписки"
	//
	//дату принимаю в строке, основываясь на примере запроса в ТЗ
	//FirstDate - начало временного отрезка, за который пользователь хочет получить статистику
	//LastDate - конец временного отрезка
	CalculateTotal(ctx context.Context, userID uuid.UUID, serviceName string, FirstDate, LastDate string) (int, error)
}
