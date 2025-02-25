/*
Copyright 2022 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instrumentation

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/GoogleContainerTools/skaffold/v2/pkg/skaffold/output/log"
)

// a single Google Cloud Monitoring write request can accommodate a maximum of 200 time series
// so we write the first 200 records and ignore the remaining.
// see https://cloud.google.com/monitoring/quotas
const maxRecordCount = 200

var recordCount int32 = 0

type float64ValueRecorder struct {
	name string
	metric.Float64ValueRecorder
}

func (c float64ValueRecorder) Record(ctx context.Context, value float64, labels ...attribute.KeyValue) {
	if atomic.AddInt32(&recordCount, 1) >= maxRecordCount {
		log.Entry(ctx).Debugf("skipping recording metric %q, maximum quota of %q exceeded", c.name, maxRecordCount)
		return
	}
	c.Float64ValueRecorder.Record(ctx, value, labels...)
}

type int64ValueRecorder struct {
	name string
	metric.Int64ValueRecorder
}

func (c int64ValueRecorder) Record(ctx context.Context, value int64, labels ...attribute.KeyValue) {
	if atomic.AddInt32(&recordCount, 1) >= maxRecordCount {
		log.Entry(ctx).Debugf("skipping recording metric %q, maximum quota of %d exceeded", c.name, maxRecordCount)
		return
	}
	c.Int64ValueRecorder.Record(ctx, value, labels...)
}

func NewFloat64ValueRecorder(m metric.Meter, name string, mos ...metric.InstrumentOption) float64ValueRecorder {
	return float64ValueRecorder{name: name, Float64ValueRecorder: metric.Must(m).NewFloat64ValueRecorder(name, mos...)}
}

func NewInt64ValueRecorder(m metric.Meter, name string, mos ...metric.InstrumentOption) int64ValueRecorder {
	return int64ValueRecorder{name: name, Int64ValueRecorder: metric.Must(m).NewInt64ValueRecorder(name, mos...)}
}
