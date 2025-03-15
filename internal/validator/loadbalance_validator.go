package validator

import (
	"github.com/lccxxo/bailuoli/internal/constants"
	"github.com/lccxxo/bailuoli/internal/model"
)

type LoadBalanceValidator struct {
	BaseValidator
}

func (l *LoadBalanceValidator) Validate(route *model.Route) error {
	if route.LoadBalance.Strategy == "weighted" && len(route.LoadBalance.Weighted) == 0 {
		return constants.ErrNoWeightedConfig
	}

	if route.LoadBalance.MaxConn < 0 {
		return constants.ErrCountIllegal
	}

	return nil
}
