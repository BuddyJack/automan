package base

import "../model"
type BasicStat interface {
	Metrics() []*model.MetricValue
}
