// Здесь будут описываться тела сообщений (как в gRPC контракте)
package http

import "github.com/google/uuid"

// ТЕГИ:
// json: название заголовка
//
// validate: работает также, как в бд при миграции
// 1.required значит, что поле не может быть <= 0
// 2.uuid значит, что строка должна соответствовать формату uuid
// 3.gt=0 значит, что число должно быть больше нуля
//
// example: теги для сваггера

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" validate:"required" example:"Yandex Plus"`
	Price       int     `json:"price" validate:"required" example:"400"`
	UserID      string  `json:"user_id" validate:"required,uuid" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" validate:"required" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"08-2025"`
}

type CreateSubscriptionResponse struct {
	ID          int     `json:"id" example:"1"`
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"08-2025"`
}

type GetStatsRequest struct {
	UserID      string `json:"user_id" validate:"required,uuid"`
	ServiceName string `json:"service_name" validate:"required"`
	FirstDate   string `json:"first_date" validate:"required" example:"01-2025"`
	LastDate    string `json:"last_date" validate:"required" example:"12-2025"`
}

type StatsResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	TotalSum int       `json:"total_sum" example:"1200"`
}
