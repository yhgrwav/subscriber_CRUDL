package service

import (
	"context"
	"testovoe_again/internal/domain"
	"testovoe_again/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CRUDL Методы для сервиса
type SubService interface {
	Create(ctx context.Context, sub domain.Subscription) (int, error)
	Read(ctx context.Context, id int) (domain.Subscription, error)
	Update(ctx context.Context, sub domain.Subscription) error
	Delete(ctx context.Context, id int) error
	GetListByUserID(ctx context.Context, UserID uuid.UUID) ([]domain.Subscription, error)

	//Втрой пункт ТЗ
	//ручка "для подсчета суммарной стоимости всех подписок за
	//выбранный период с фильтрацией по id пользователя и названию подписки"
	//
	//дату принимаю в строке, основываясь на примере запроса в ТЗ
	//FirstDate - начало временного отрезка, за который пользователь хочет получить статистику
	//LastDate - конец временного отрезка
	CalculateTotal(ctx context.Context, userID uuid.UUID, serviceName string, FirstDate, LastDate string) (int, error)
}

type SubscriptionService struct {
	logger *zap.Logger
	repo   repository.SubscriptionRepository
}

func NewSubscriptionService(logger *zap.Logger, repo repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{logger: logger, repo: repo}
}

func (s *SubscriptionService) Create(ctx context.Context, sub domain.Subscription) (int, error) {
	// вообще я по идее для такого сервиса должен был бы ходить в базу или кэш для того, чтобы сравнить поля
	// ServiceName, Price, айдишники и уже на основе этой проверки давать ok || !ok, но в рамках проекта я по сути своей
	// ничего кроме наивного решения типа проверки sub.Price < 0 && > 10000 сделать не могу
	// в целом это бы выглядело как-то так: if sub.ServiceName != db.ServiceName { error }

	//если пришла невалидная цена - кидаем Warn и делаем json ошибки для логов
	err := ValidatePrice(sub.Price)
	if err != nil {
		s.logger.Warn("невалидная цена", zap.Error(err), zap.Int("price", sub.Price))
		return 0, err
	}
	// можно было бы сюда добавить проверку на существование указанного сервиса спрашивая у редиса есть ли у нас сервис или нет
	// и можно было бы проверить юзера по базе, но ради одного userID создавать базу смысла не особо много в рамках тестового
	result, err := s.repo.Create(ctx, sub)
	if err != nil {
		s.logger.Error("ошибка создания подписки", zap.Error(err))
		return 0, err
	}
	return result, nil
}

func (s *SubscriptionService) Read(ctx context.Context, id int) (domain.Subscription, error) {
	// Тут можно было бы сделать поход в кэш, в случае если объект был недавно создан и у него не истёк TTL
	// в целом логика простая - принимаем id подписки, если такой нет - кидаем ошибку, если есть - отдаём указатель

	// здесь должна быть условная проверка в кэше, но т.к. кэша нет - идём дальше
	// т.к. технически мы не можем проверить наличие айди без похода в базу - сразу вызываем репозиторий
	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("подписка не найдена", zap.Int("id", id))
		return domain.Subscription{}, err
	}
	return result, nil
}
