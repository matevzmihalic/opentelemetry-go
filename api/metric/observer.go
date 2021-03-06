// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metric

// BatchObserver represents an Observer callback that can report
// observations for multiple instruments.
type BatchObserver struct {
	meter  Meter
	runner AsyncBatchRunner
}

// Int64Observer is a metric that captures a set of int64 values at a
// point in time.
type Int64Observer struct {
	asyncInstrument
}

// Float64Observer is a metric that captures a set of float64 values
// at a point in time.
type Float64Observer struct {
	asyncInstrument
}

// Observation returns an Observation, a BatchObserverCallback
// argument, for an asynchronous integer instrument.
// This returns an implementation-level object for use by the SDK,
// users should not refer to this.
func (i Int64Observer) Observation(v int64) Observation {
	return Observation{
		number:     NewInt64Number(v),
		instrument: i.instrument,
	}
}

// Observation returns an Observation, a BatchObserverCallback
// argument, for an asynchronous integer instrument.
// This returns an implementation-level object for use by the SDK,
// users should not refer to this.
func (f Float64Observer) Observation(v float64) Observation {
	return Observation{
		number:     NewFloat64Number(v),
		instrument: f.instrument,
	}
}
