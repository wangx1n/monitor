package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"sync"
	"time"
)

var wg sync.WaitGroup

type labels struct {
	Name     string `json:"__name__"`
	Instance string `json:"instance"`
	Job      string `json:"job"`
	Handler  string `json:"handler"`
	Le       string `json:"le"`
	Code     int64  `json:"code"`
	Quantile string `json:"quantile"`
	Slice    string `json:"slice"`
	Reason   string `json:"reason"`
}

type monitorData struct {
	Labels    labels    `json:"labels"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

func main() {
	// init influxdb client
	client := influxdb2.NewClientWithOptions("http://192.168.31.244:8086",
		"iAmvMfGNC4Z2olzE5mfTEA_zvRO5n84ye7N1OX6mTcqZxzrAXQ6LajjXVN6z-Fs_68z3ngFIVquglmdBZfBNhw==",
		influxdb2.DefaultOptions().SetBatchSize(300))
	defer client.Close()

	consumer, err := sarama.NewConsumer([]string{"192.168.31.118:9092"}, nil)
	if err != nil {
		fmt.Println("consumer connect err:", err)
		return
	}
	defer consumer.Close()

	//获取 kafka 主题
	partitions, err := consumer.Partitions("prometheus-metric")
	if err != nil {
		fmt.Println("get partitions failed, err:", err)
		return
	}

	// get non-blocking write client
	writeAPI := client.WriteAPI("wangxin.vettel.co", "monitor_data_to_influxDB")

	for _, p := range partitions {
		//sarama.OffsetNewest：从当前的偏移量开始消费，sarama.OffsetOldest：从最老的偏移量开始消费
		partitionConsumer, err := consumer.ConsumePartition("prometheus-metric", p, sarama.OffsetNewest)
		if err != nil {
			fmt.Println("partitionConsumer err:", err)
			continue
		}
		wg.Add(1)
		go func() {
			for m := range partitionConsumer.Messages() {
				md, err := deserialized(m.Value)
				if err != nil {
					fmt.Println(&md)
				}

				tagMap := map[string]string{
					"__name__": md.Labels.Name,
					"instance": md.Labels.Instance,
					"job":      md.Labels.Job,
					"Handler":  md.Labels.Handler,
					"Le":       md.Labels.Le,
					"code":     string(md.Labels.Code),
					"quantile": md.Labels.Quantile,
					"slice":    md.Labels.Slice,
					"reason":   md.Labels.Reason,
				}

				fieldMap := map[string]interface{}{
					"name":  md.Name,
					"value": md.Value,
				}
				// create point using fluent style
				point := influxdb2.NewPoint("measurement_test",
					tagMap,
					fieldMap,
					md.Timestamp)
				// write point asynchronously
				writeAPI.WritePoint(point)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func deserialized(msg []byte) (md monitorData, err error) {
	err = json.Unmarshal(msg, &md)
	if err != nil {
		return monitorData{}, err
	}
	return
}
