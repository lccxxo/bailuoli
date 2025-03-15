package validator

import "github.com/lccxxo/bailuoli/internal/model"

// 使用责任链模式实现路由验证

type Validator interface {
	Validate(*model.Route) error
	SetNext(Validator)
}

type BaseValidator struct {
	next Validator
}

func (v *BaseValidator) SetNext(next Validator) {
	v.next = next
}

func NewValidationChain() Validator {
	pathValidator := &PathValidator{}
	matchTypeValidator := &MatchTypeValidator{}
	lbValidator := &LoadBalanceValidator{}

	pathValidator.SetNext(matchTypeValidator)
	matchTypeValidator.SetNext(lbValidator)
	return pathValidator
}
