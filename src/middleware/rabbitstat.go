package middleware

import (
	"../ntos"
	"../model"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strconv"
)

//rabbitmq metrics  http call rabbitmq manager
type RabbitNode struct {
	Node        string
	MsgReady    string
	MsgUnack    string
	Connections string
	Consumers   string
	Queues      string
}

type RabbitQueue struct {
	Node             string
	Vhost            string
	Queue            string
	QueueStat        string
	QueueConsumers   string
	QueueMsgUnack    string
	QueueMsgReady    string
	PublishRate      string
	DeliverNoAckRate string
	DeliverGetRate   string
	AckRate          string
}

type RabbitConfig struct {
	Host   string
	Port   uint64
	User   string
	Passwd string
}

func listRabbitNodes(configs []RabbitConfig) (nodes []*RabbitNode) {
	for _, oneConfig := range configs {
		var url = "http://" + oneConfig.Host + ":" + strconv.FormatUint(oneConfig.Port, 10) + "/api/overview"
		client := http.DefaultClient
		req, _ := http.NewRequest("GET", url, nil)
		req.SetBasicAuth(oneConfig.User, oneConfig.Passwd)
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		bs, _ := ioutil.ReadAll(resp.Body)
		//println(string(bs))
		var target interface{}
		json.Unmarshal(bs, &target)
		m := target.(map[string]interface{})
		nodeName := m["node"].(string)
		queueTotalMap := m["queue_totals"].(map[string]interface{})
		msgReady := ntos.F64toS(queueTotalMap["messages_ready"].(float64))
		msgUnack := ntos.F64toS(queueTotalMap["messages_unacknowledged"].(float64))
		objectMap := m["object_totals"].(map[string]interface{})
		conns := ntos.F64toS(objectMap["connections"].(float64))
		consumers := ntos.F64toS(objectMap["consumers"].(float64))
		queues := ntos.F64toS(objectMap["queues"].(float64))
		nodes = append(nodes, &RabbitNode{Node: nodeName, MsgReady: msgReady, MsgUnack: msgUnack, Connections: conns, Consumers: consumers, Queues: queues})
	}
	return
}

func listRabbitQueus(configs []RabbitConfig) (queues []*RabbitQueue) {
	for _, oneConfig := range configs {
		var url = "http://" + oneConfig.Host + ":" + strconv.FormatUint(oneConfig.Port, 10) + "/api/queues"
		client := http.DefaultClient
		req, _ := http.NewRequest("GET", url, nil)
		req.SetBasicAuth(oneConfig.User, oneConfig.Passwd)
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		var target interface{}
		bs, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bs, &target)
		v := target.([]interface{})
		for _, oneQueue := range v {
			oneQueueMap := oneQueue.(map[string]interface{})
			one := RabbitQueue{Node: oneQueueMap["node"].(string), Queue: oneQueueMap["name"].(string), Vhost: oneQueueMap["vhost"].(string)}
			one.QueueStat = oneQueueMap["state"].(string)
			one.QueueConsumers = ntos.F64toS(oneQueueMap["consumers"].(float64))
			one.QueueMsgUnack = ntos.F64toS(oneQueueMap["messages_unacknowledged"].(float64))
			one.QueueMsgReady = ntos.F64toS(oneQueueMap["messages_ready"].(float64))

			if message_stats, exist := oneQueueMap["message_stats"]; exist {
				messageStats := message_stats.(map[string]interface{})
				if _, exist := messageStats["publish_details"]; exist {
					publish_details := messageStats["publish_details"].(map[string]interface{})
					one.PublishRate = ntos.F64toS(publish_details["rate"].(float64))
				}
				if _, exist := messageStats["deliver_no_ack_details"]; exist {
					deliver_no_ack_details := messageStats["deliver_no_ack_details"].(map[string]interface{})
					one.DeliverNoAckRate = ntos.F64toS(deliver_no_ack_details["rate"].(float64))
				}
				if _, exist := messageStats["deliver_get_details"]; exist {
					deliver_get_details := messageStats["deliver_get_details"].(map[string]interface{})
					one.DeliverGetRate = ntos.F64toS(deliver_get_details["rate"].(float64))
				}
				if _, exist := messageStats["ack_details"]; exist {
					ack_details := messageStats["ack_details"].(map[string]interface{})
					one.AckRate = ntos.F64toS(ack_details["rate"].(float64))
				}
			}
			queues = append(queues, &one)
		}
	}
	return nil
}

func (*RabbitNode) Metrics(configs []RabbitConfig) (metrics []*model.MetricValue) {
	nodes := listRabbitNodes(configs)
	for _, node := range nodes {
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.overview.message_ready", Value: node.MsgReady, Tags: map[string]string{"node": node.Node}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.overview.message_unack", Value: node.MsgUnack, Tags: map[string]string{"node": node.Node}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.overview.connections", Value: node.Connections, Tags: map[string]string{"node": node.Node}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.overview.consumers", Value: node.Consumers, Tags: map[string]string{"node": node.Node}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.overview.queues", Value: node.Queues, Tags: map[string]string{"node": node.Node}})
	}
	queues := listRabbitQueus(configs)
	for _, queue := range queues {
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.queue.consumers", Value: queue.QueueConsumers, Tags: map[string]string{"node": queue.Node, "vhost": queue.Vhost, "queue": queue.Queue}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.queue.messages_unacknowledged", Value: queue.QueueMsgUnack, Tags: map[string]string{"node": queue.Node, "vhost": queue.Vhost, "queue": queue.Queue}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.queue.messages_ready", Value: queue.QueueMsgReady, Tags: map[string]string{"node": queue.Node, "vhost": queue.Vhost, "queue": queue.Queue}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.queue.publish_rate", Value: queue.PublishRate, Tags: map[string]string{"node": queue.Node, "vhost": queue.Vhost, "queue": queue.Queue}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.queue.deliver_no_ack_rate", Value: queue.DeliverNoAckRate, Tags: map[string]string{"node": queue.Node, "vhost": queue.Vhost, "queue": queue.Queue}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.queue.deliver_get_rate", Value: queue.DeliverGetRate, Tags: map[string]string{"node": queue.Node, "vhost": queue.Vhost, "queue": queue.Queue}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "rabbit", Metric: "rabbit.queue.ack_rate", Value: queue.AckRate, Tags: map[string]string{"node": queue.Node, "vhost": queue.Vhost, "queue": queue.Queue}})
	}
	return
}
