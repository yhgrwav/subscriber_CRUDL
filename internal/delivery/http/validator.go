package http

import "github.com/go-playground/validator/v10"

type Validator struct {
	Validater *validator.Validate
}

func (v *Validator) NewValidator() *Validator {
	return &Validator{Validater: validator.New()}
}

func (v *Validator) Validate(i interface{}) error {
	// i - любые данные, которые принимает валидатор
	// гошный валидатор работает через метод .Struct, соответственно мы вызываем метод для принимаемых данных
	return v.Validater.Struct(i)
}
