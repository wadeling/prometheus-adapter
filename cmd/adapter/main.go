package main

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	//"github.com/prometheus/client_golang/prometheus"
	//"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/wadeling/prometheus-adapter/pkg/logger"
	"io/ioutil"
	"net/http"
	"os"
	//"sync"
)

type prometheusAdapter struct {
	listenPort string
	listenIP string
}

var log = logger.NewLogger(true)

func defaultAdapter() *prometheusAdapter{
	return &prometheusAdapter{
		listenPort: "9100",
	}
}

func GetCmd() *cobra.Command {
	adapter := defaultAdapter()

	cmd := &cobra.Command{
		Use:   "prometheus-adapter",
		Short: "prometheus-adapter",
		Run: func(cmd *cobra.Command, args []string) {
			adapter.run()
		},
	}

	f := cmd.PersistentFlags()
	f.StringVarP(&adapter.listenPort, "port", "p", adapter.listenPort, "TCP port to use for adapter")
	f.StringVarP(&adapter.listenIP, "listen IP", "i", adapter.listenIP, "TCP IP to use for adapter")

	return cmd
}

func samplesToString(samples *model.Samples) string {
	var result string
	for _, s := range *samples {
		//fmt.Println("time ",s.Timestamp)
		//fmt.Println("value ",s.Value)
		//fmt.Println("metric ",s.Metric.String())
		tmp := fmt.Sprintln("%d %f %s",s.Timestamp,s.Value,s.Metric.String())
		result += tmp
	}

	return result
}

func protoToSamples(req *prompb.WriteRequest) model.Samples {
	var samples model.Samples
	for _, ts := range req.Timeseries {
		metric := make(model.Metric, len(ts.Labels))
		for _, l := range ts.Labels {
			//fmt.Println("label name",l.Name)
			//fmt.Println("label value",l.Value)
			metric[model.LabelName(l.Name)] = model.LabelValue(l.Value)
		}

		for _, s := range ts.Samples {
			samples = append(samples, &model.Sample{
				Metric:    metric,
				Value:     model.SampleValue(s.Value),
				Timestamp: model.Time(s.Timestamp),
			})
		}
	}
	return samples
}

func (m *prometheusAdapter) run() error {
	log.Info("run")
	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		compressed, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("read error",zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//log.Info("recieve data",zap.Any("compressed data",compressed))

		reqBuf, err := snappy.Decode(nil, compressed)
		if err != nil {
			log.Error("Decode err",zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//log.Info("req buf",zap.Any("buff",reqBuf))

		var req prompb.WriteRequest
		if err := proto.Unmarshal(reqBuf, &req); err != nil {
			log.Error("Unmarshal error",zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//log.Debug("unmarshal to promb write request",zap.Any("req",req.String()))

		samples := protoToSamples(&req)
		log.Info("Recieve sample count",zap.Any("sample-count",samples.Len()))

		//_ = samplesToString(&samples)
		//log.Debug("samples content",zap.String("content",sampleStrs))

	})

	addr := fmt.Sprintf("%s:%s",m.listenIP,m.listenPort)

	return http.ListenAndServe(addr, nil)
}

func main() {
	fmt.Println("hello adapter")
	log.Info("hello adapter")

	cmd := GetCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(-1)
	}
}