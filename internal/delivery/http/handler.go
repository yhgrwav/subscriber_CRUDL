package http

import (
	"strconv"
	"testovoe_again/internal/domain"
	"testovoe_again/internal/errors"
	"testovoe_again/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	service service.SubService
}

func NewHandler(logger *zap.Logger, service service.SubService) *Handler {
	return &Handler{logger: logger, service: service}
}

// @Summary      создать подписку
// @Description  создает новую запись о подписке и возвращает её тело
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        input body CreateSubscriptionRequest true "данные новой подписки"
// @Success      201 {object} CreateSubscriptionResponse
// @Failure      400 {object} map[string]string "не удалось обработать запрос"
// @Failure      500 {object} map[string]string "ошибка сервера"
// @Router       /api/v1/subscriptions [post]
func (h *Handler) Create(c echo.Context) error {
	// создаём переменную для DTO_шки, куда будем записывать результат для похода в сервис
	var request CreateSubscriptionRequest

	// с помощью Bind метода раскидываем поля в структуру, если ошибка - отдаём 400
	if err := c.Bind(&request); err != nil {
		h.logger.Warn("не удалось обработать запрос", zap.Error(err))
		return echo.NewHTTPError(400, err.Error())
	}

	// с помощью Validate проверяем полученные поля и если получаем не состыковку с тегами - кидаем ошибку(400)
	if err := c.Validate(request); err != nil {
		return echo.NewHTTPError(400, err.Error())
	}

	// т.к. DTO != domain - перекладываем поля в требуемую для метода структуру
	sub, err := h.ToDomain(request)
	if err != nil {
		return echo.NewHTTPError(400, err.Error())
	}

	// если всё ок на этом уровне - вызываем сервис
	id, err := h.service.Create(c.Request().Context(), sub)
	if err != nil {
		// осознанно(!!!) распаковываю ошибку в рамках контексте тестового задания, понимаю что бест практис - вернуть кастом
		return echo.NewHTTPError(500, err.Error())
	}

	// если всё сработало - возвращаем 201(created) и структуру ответа из DTO
	return c.JSON(201, CreateSubscriptionResponse{
		ID:          id,
		ServiceName: request.ServiceName,
		Price:       request.Price,
		UserID:      request.UserID,
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
	})
}

func (h *Handler) ToDomain(input CreateSubscriptionRequest) (domain.Subscription, error) {
	uid, err := uuid.Parse(input.UserID)
	if err != nil {
		h.logger.Warn("не удалось обработать UUID пользователя", zap.String("uuid", input.UserID))
		return domain.Subscription{}, err
	}

	return domain.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      uid,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	}, nil
}

// GetByID godoc
// @Summary      получить подписку по ID
// @Description  возвращает данные конкретной подписки по её уникальному идентификатору
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      int  true  "ID подписки"
// @Success      200  {object}  CreateSubscriptionResponse
// @Failure      400  {object}  map[string]string "невалидный ID"
// @Failure      404  {object}  map[string]string "подписка не найдена"
// @Failure      500  {object}  map[string]string "ошибка сервера"
// @Router       /api/v1/subscriptions/{id} [get]
func (h *Handler) GetByID(c echo.Context) error {
	//читаем из query айдишник, если ошибка - отдаём 400
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn("невалидный id", zap.Int("id", id))
		return echo.NewHTTPError(400, "невалидный id")
	}

	// если айди валидный - вызываем сервис
	sub, err := h.service.Read(c.Request().Context(), id)
	if err != nil {
		if err == errors.ErrSubscriptionNotFound {
			h.logger.Warn("пользователь не найден", zap.Int("id", id))
			return c.JSON(404, err.Error())
		}
		h.logger.Error("не удалось найти пользователя", zap.Error(err))
		return echo.NewHTTPError(500, "ошибка сервера")
	}

	return c.JSON(200, sub)
}

// Update godoc
// @Summary      обновить подписку
// @Description  обновляет данные существующей подписки по её ID и требует полное тело запроса
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id    path    int                        true  "ID подписки"
// @Param        input body    CreateSubscriptionRequest  true  "новые данные подписки"
// @Success      204   "No Content"
// @Failure      400   {object} map[string]string "невалидный ID или тело запроса"
// @Failure      404   {object} map[string]string "подписка не найдена"
// @Failure      500   {object} map[string]string "ошибка сервера"
// @Router       /api/v1/subscriptions/{id} [put]
func (h *Handler) Update(c echo.Context) error {
	//читаем айди
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn("невалидный id", zap.Int("id", id))
		return c.JSON(400, err.Error())
	}

	// переменная с телом запроса
	var request CreateSubscriptionRequest
	if err := c.Bind(&request); err != nil {
		h.logger.Warn("невалидное тело для обновления", zap.Error(err))
		return echo.NewHTTPError(400, err.Error())
	}

	// переводим всё в необходимую структуру
	result, err := h.ToDomain(request)
	if err != nil {
		//не валидировал здесь, т.к. логер реализовал внутри метода
		return echo.NewHTTPError(400, err.Error())
	}

	//прокидываем ID, т.к. метод ToDomain не работает с ID
	result.ID = id

	//вызываем сервис
	err = h.service.Update(c.Request().Context(), result)
	if err != nil {
		h.logger.Warn("ошибка обработки запроса обновления", zap.Error(err))
		return echo.NewHTTPError(500, err.Error())
	}

	//если всё ок - отдаём 204, т.к. метод Update ничего не возвращает
	return c.NoContent(204)
}

// Delete godoc
// @Summary      удалить подписку
// @Description  удаляет запись о подписке из базы данных по её ID
// @Tags         subscriptions
// @Param        id   path      int  true  "id подписки"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string "невалидный id"
// @Failure      500  {object}  map[string]string "ошибка удаления"
// @Router       /api/v1/subscriptions/{id} [delete]
func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn("невалидный id", zap.Int("id", id))
		return echo.NewHTTPError(400, "невалидный id")
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		h.logger.Warn("не удалось удалить подписку", zap.Error(err))
		return echo.NewHTTPError(500, "ошибка удаления")
	}

	return c.NoContent(204)
}

// List godoc
// @Summary      список подписок пользователя
// @Description  возвращает все активные подписки конкретного пользователя по его UUID
// @Tags         subscriptions
// @Produce      json
// @Param        user_id  path      string  true  "UUID пользователя"
// @Success      200      {array}   CreateSubscriptionResponse
// @Failure      400      {object}  map[string]string "невалидный айди пользователя"
// @Failure      500      {object}  map[string]string "не удалось получить подписки"
// @Router       /api/v1/subscriptions/list/{user_id} [get]
func (h *Handler) List(c echo.Context) error {
	id := c.Param("user_id")

	uid, err := uuid.Parse(id)
	if err != nil {
		h.logger.Warn("невалидный uuid", zap.String("id", id))
		return echo.NewHTTPError(400, "невалидный айди пользователя")
	}

	subscriptions, err := h.service.GetListByUserID(c.Request().Context(), uid)
	if err != nil {
		h.logger.Error("ошибка получения списка подписок", zap.Error(err))
		return echo.NewHTTPError(500, "не удалось получить подписки")
	}

	return c.JSON(200, subscriptions)
}

// GetSum godoc
// @Summary      рассчитать сумму затрат
// @Description  возвращает суммарную стоимость подписок по конкретному сервису за указанный период
// @Tags         analytics
// @Accept       json
// @Produce      json
// @Param        request  body      GetStatsRequest  true  "параметры фильтрации (UserID, ServiceName, Dates)"
// @Success      200      {object}  StatsResponse
// @Failure      400      {object}  map[string]string "невалидный запрос"
// @Failure      400      {object}  map[string]string "невалидный айди пользователя"
// @Failure      500      {object}  map[string]string "ошибка расчёта суммы"
// @Router       /api/v1/stats [post]
func (h *Handler) GetSum(c echo.Context) error {
	var request GetStatsRequest

	if err := c.Bind(&request); err != nil {
		h.logger.Warn("не удалось распарсить тело запроса статистики", zap.Error(err))
		return echo.NewHTTPError(400, "невалидный запрос")
	}

	if err := c.Validate(&request); err != nil {
		return echo.NewHTTPError(400, err.Error())
	}

	uid, err := uuid.Parse(request.UserID)
	if err != nil {
		h.logger.Warn("невалидный uuid в запросе суммы", zap.String("id", request.UserID))
		return echo.NewHTTPError(400, "невалидный айди пользователя")
	}

	result, err := h.service.CalculateTotal(
		c.Request().Context(),
		uid,
		request.ServiceName,
		request.FirstDate,
		request.LastDate,
	)
	if err != nil {
		h.logger.Error("ошибка расчета суммы", zap.Error(err))
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, StatsResponse{
		UserID:   uid,
		TotalSum: result,
	})
}

// Health godoc
// @Summary      проверка работоспособности
// @Description  простой эндпоинт для проверки того, что сервер запущен
// @Tags         system
// @Success      200  {string}  string "OK"
// @Router       /api/v1/healthcheck [get]
func (h *Handler) Health(c echo.Context) error {
	return c.String(200, "OK")
}
