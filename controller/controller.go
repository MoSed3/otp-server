package controller

type Operator int

const (
	OperatorSystem Operator = iota // for cli, tui or  cron tasks
	OperatorWeb
	OperatorApi
)

type Controller struct {
	operator Operator
}

func (c *Controller) Operator() Operator {
	return c.operator
}

func New(o Operator) Controller {
	return Controller{operator: o}
}
