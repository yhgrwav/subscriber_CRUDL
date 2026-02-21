package service

import (
	"context"
	"testovoe_again/internal/domain"
	"testovoe_again/internal/errors"
	"testovoe_again/internal/repository"
	"time"

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

	_, err = ValidateDate(sub.StartDate)
	if err != nil {
		s.logger.Warn("невалидная дата", zap.String("StartDate", sub.StartDate))
		return 0, err
	}

	//т.к. EndDate у нас может и не быть, я сделал проверку на nil чтобы не ловить панику в этом кейсе
	if sub.EndDate != nil {
		_, err := ValidateDate(*sub.EndDate)
		if err != nil {
			s.logger.Warn("невалидная дата", zap.String("EndDate", *sub.EndDate))
			return 0, err
		}
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

func (s *SubscriptionService) Update(ctx context.Context, sub domain.Subscription) error {
	// логика такая - идём в базу за подпиской, которую хотим изменить
	// затем записываем её в переменную и обновляем принимаемые поля
	// если подписки нет - отдаём ошибку, если какое-то поле не обновили - оставляем старое
	err := ValidatePrice(sub.Price)
	if err != nil {
		s.logger.Warn("невалидная цена", zap.Int("price", sub.Price))
		return err
	}
	_, err = ValidateDate(sub.StartDate)
	if err != nil {
		s.logger.Warn("невалидная дата", zap.String("Date", sub.StartDate))
		return err
	}
	if sub.EndDate != nil {
		_, err = ValidateDate(*sub.EndDate)
		if err != nil {
			s.logger.Warn("невалидная дата", zap.String("Date", *sub.EndDate))
			return err
		}
	}
	OldVersion, err := s.repo.GetByID(ctx, sub.ID)
	if err != nil {
		s.logger.Warn("такой подписки не существует", zap.Int("id", sub.ID))
		return err
	}
	OldVersion.Price = sub.Price
	OldVersion.StartDate = sub.StartDate
	OldVersion.EndDate = sub.EndDate
	OldVersion.ServiceName = sub.ServiceName
	err = s.repo.Update(ctx, sub.ID, OldVersion)
	if err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionService) GetListByUserID(ctx context.Context, UserID uuid.UUID) ([]domain.Subscription, error) {
	result, err := s.repo.GetByUserID(ctx, UserID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SubscriptionService) CalculateTotal(ctx context.Context, UserID uuid.UUID, serviceName, FirstDate, LastDate string) (int, error) {
	// т.к. по сути своей функция обязательно должна принимать какой-то временной период - вторая дата не будет передаваться
	// через указатель, соответственно валидация у неё будет выглядеть идентично первой дате
	// суть в том, что мы принимаем строки, валидируем и парсим, затем подставляем и возвращаем результат или ошибку
	t1, err := ValidateDate(FirstDate)
	if err != nil {
		return 0, err
	}
	t2, err := ValidateDate(LastDate)
	if err != nil {
		return 0, err
	}

	result, err := s.repo.GetStatsByServiceName(ctx, UserID, serviceName, t1, t2)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func ValidatePrice(price int) error {
	if price < 0 || price > 10000 {
		return errors.ErrInvalidPrice
	}
	return nil
}

// суть валидатора - проверить строку и вернуть time.Time
// НО
// если нам нужна ТОЛЬКО валидация (либо всё хорошо либо ошибка), то мы можем вызвать метод игнорируя time.Time ответ
// и по итогу мы просто проверим, не получили ли мы случайно невалидную строку
func ValidateDate(date string) (time.Time, error) {
	t, err := time.Parse("01-2006", date)
	if err != nil {
		return time.Time{}, errors.ErrInvalidDateFormat
	}
	return t, nil
}
