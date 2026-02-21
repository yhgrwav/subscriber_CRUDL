package repository

import (
	"context"
	"database/sql"
	"testovoe_again/internal/domain"
	"testovoe_again/internal/errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CRUDL методы для репозитория
type SubscriptionRepository interface {
	Create(ctx context.Context, sub domain.Subscription) (int, error)
	GetByID(ctx context.Context, id int) (domain.Subscription, error)
	Update(ctx context.Context, id int, sub domain.Subscription) error
	Delete(ctx context.Context, id int) error
	GetStatsByServiceName(ctx context.Context, userID uuid.UUID, serviceName string, time1, time2 time.Time) (int, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error)
}
type PostgresRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresRepo(db *sql.DB, logger *zap.Logger) *PostgresRepo {
	return &PostgresRepo{db: db, logger: logger}
}

func (p *PostgresRepo) Create(ctx context.Context, sub domain.Subscription) (int, error) {
	query := `INSERT INTO subscriptions (service_name, Price, user_id, start_date) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	var id int

	//использовал не Exec(), т.к. хочу явно передавать контекст для того, чтобы имплементировать интерфейсу
	//и плюсом т.к. мы в ответе всегда получаем одну строку, этот метод идеально подходит для вычитывания одной строки + id для ответа
	err := p.db.QueryRowContext(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate).Scan(&id)
	if err != nil {
		p.logger.Error("ошибка при создании подписки", zap.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *PostgresRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date 
        FROM subscriptions 
        WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("ошибка получения пользователя", zap.Error(err))
		return nil, err
	}

	defer rows.Close()

	var subscriptions []domain.Subscription

	for rows.Next() {
		var sub domain.Subscription
		var startT, endT sql.NullTime

		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&startT,
			&endT,
		)
		if err != nil {
			r.logger.Error("ошибка скана строки подписки", zap.Error(err))
			return nil, err
		}

		sub.StartDate = startT.Time.Format("01-2006")
		if endT.Valid {
			strEnd := endT.Time.Format("01-2006")
			sub.EndDate = &strEnd
		}

		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("ошибка итерации по строке", zap.Error(err))
		return nil, err
	}

	return subscriptions, nil
}

// Update будет давать возможность обновить какую-то запись о подписке по ID.
// в ТЗ сказано "... ручки для операций над записями о подписках", т.е. Я подразумеваю что этим методом будет пользоваться
// не пользователь сервиса подписок, а метод будет систематически вызываться условно для продления подписки, возможно
// в случае обновлении цены, обновлении end_date.
func (r *PostgresRepo) Update(ctx context.Context, id int, sub domain.Subscription) error {
	query := `Update subscriptions 
			  SET Price = $1, end_date = $2
			  WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, sub.Price, sub.EndDate, id)
	if err != nil {
		r.logger.Error("ошибка обновления подписки", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostgresRepo) Delete(ctx context.Context, id int) error {
	query := `Delete from subscriptions 
              WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("ошибка удаления подписки", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostgresRepo) GetStatsByServiceName(ctx context.Context, userID uuid.UUID, serviceName string, time1, time2 time.Time) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0) 
		FROM subscriptions 
		WHERE user_id = $1 
		  AND service_name = $2 
		  AND start_date BETWEEN $3 AND $4`

	var result int

	err := r.db.QueryRowContext(ctx, query, userID, serviceName, time1, time2).Scan(&result)
	if err != nil {
		r.logger.Error("ошибка получения статистики", zap.Error(err))
		return 0, err
	}
	return result, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int) (domain.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date
			  FROM subscriptions
			  WHERE id = $1`

	var (
		result       domain.Subscription
		StartT, EndT sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(&result.ID,
		&result.ServiceName,
		&result.Price,
		&result.UserID,
		&StartT,
		&EndT)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("подписка не найдена", zap.Int("id", id))
			return domain.Subscription{}, errors.ErrSubscriptionNotFound
		}
		r.logger.Warn(err.Error(), zap.Int("id", id))
		return domain.Subscription{}, err
	}

	result.StartDate = StartT.Time.Format("01-2006")
	if EndT.Valid {
		strEnd := EndT.Time.Format("01-2006")
		result.EndDate = &strEnd
	}

	return result, nil
}
