package model

type MetricValue struct {
	Metric string `json:metric`
	Value  string  `json:value`
	Tags  map[string]string `json:tags`
	Attributes map[string]string `json:attributes`
	Timestamp uint64    `json:timestamp`
	Endpoint string     `json:endpoint`
}
