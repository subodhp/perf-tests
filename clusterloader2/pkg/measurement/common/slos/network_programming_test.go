/*
Copyright 2019 The Kubernetes Authors.

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

package slos

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	measurementutil "k8s.io/perf-tests/clusterloader2/pkg/measurement/util"
)

func TestGather(t *testing.T) {
	cases := []struct {
		samples   []*model.Sample
		err       error
		wantData  *measurementutil.PerfData
		wantError error
	}{{
		samples:  []*model.Sample{{Value: 200.5}, {Value: 100.5}, {Value: 300.5}},
		wantData: createPerfData([]float64{100500, 200500, 300500}),
	}, {
		samples:   []*model.Sample{{Value: 1}},
		wantError: errors.New("got unexpected number of samples: 1"),
	}, {
		samples:   []*model.Sample{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}},
		wantError: errors.New("got unexpected number of samples: 4"),
	}}

	for _, v := range cases {
		fakeExecutor := &fakeExecutor{samples: v.samples, err: v.err}
		testGatherer(t, fakeExecutor, v.wantData, v.wantError)
	}
}

func testGatherer(t *testing.T, executor QueryExecutor, wantData *measurementutil.PerfData, wantError error) {
	g := &netProgGatherer{}
	summary, err := g.Gather(executor, time.Now())
	if err != nil {
		if wantError != nil {
			assert.Equal(t, wantError, err)
			return
		}
		t.Errorf("Unexpected error:  %v", err)
	}
	assert.Equal(t, netProg, summary.SummaryName())
	assert.Equal(t, "json", summary.SummaryExt())
	assert.NotNil(t, summary.SummaryTime())

	var data measurementutil.PerfData
	if err := json.Unmarshal([]byte(summary.SummaryContent()), &data); err != nil {
		t.Errorf("Error while decoding summary: %v. Summary: %v", err, summary.SummaryContent())

	}
	assert.Equal(t, wantData, &data)
}

type fakeExecutor struct {
	samples []*model.Sample
	err     error
}

func (f *fakeExecutor) Query(query string, queryTime time.Time) ([]*model.Sample, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.samples, nil
}
