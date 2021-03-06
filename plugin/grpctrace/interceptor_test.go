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
package grpctrace

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"

	export "go.opentelemetry.io/otel/sdk/export/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type testExporter struct {
	spanMap map[string][]*export.SpanData
}

func (t *testExporter) ExportSpan(ctx context.Context, s *export.SpanData) {
	t.spanMap[s.Name] = append(t.spanMap[s.Name], s)
}

type mockCCInvoker struct {
	ctx context.Context
}

func (mcci *mockCCInvoker) invoke(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	mcci.ctx = ctx
	return nil
}

type mockProtoMessage struct {
}

func (mm *mockProtoMessage) Reset()         {}
func (mm *mockProtoMessage) String() string { return "mock" }
func (mm *mockProtoMessage) ProtoMessage()  {}

type nameAttributeTestCase struct {
	testName     string
	expectedName string
	fullNameFmt  string
}

func (tc nameAttributeTestCase) fullName() string {
	return fmt.Sprintf(tc.fullNameFmt, tc.expectedName)
}

func TestUCISetsExpectedServiceNameAttribute(t *testing.T) {
	testCases := []nameAttributeTestCase{
		{
			"FullyQualifiedMethodName",
			"serviceName",
			"/github.com.foo.%s/bar",
		},
		{
			"SimpleMethodName",
			"serviceName",
			"/%s/bar",
		},
		{
			"MethodNameWithoutFullPath",
			"serviceName",
			"%s/bar",
		},
		{
			"InvalidMethodName",
			"",
			"invalidName",
		},
		{
			"NonAlphanumericMethodName",
			"serviceName_123",
			"/github.com.foo.%s/method",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, tc.testUCISetsExpectedNameAttribute)
	}
}

func (tc nameAttributeTestCase) testUCISetsExpectedNameAttribute(t *testing.T) {
	exp := &testExporter{make(map[string][]*export.SpanData)}
	tp, _ := sdktrace.NewProvider(sdktrace.WithSyncer(exp),
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}))

	tr := tp.Tracer("grpctrace/client")
	ctx, span := tr.Start(context.Background(), tc.testName)
	defer span.End()

	clientConn, err := grpc.Dial("fake:connection", grpc.WithInsecure())

	if err != nil {
		t.Fatalf("failed to create client connection: %v", err)
	}

	unaryInt := UnaryClientInterceptor(tr)

	req := &mockProtoMessage{}
	reply := &mockProtoMessage{}
	ccInvoker := &mockCCInvoker{}

	err = unaryInt(ctx, tc.fullName(), req, reply, clientConn, ccInvoker.invoke)
	if err != nil {
		t.Fatalf("failed to run unary interceptor: %v", err)
	}

	spanData, hasSpanData := exp.spanMap[tc.fullName()]

	if !hasSpanData || len(spanData) == 0 {
		t.Fatalf("no span data found for name < %s >", tc.fullName())
	}

	attributes := spanData[0].Attributes

	var actualServiceName string
	for _, attr := range attributes {
		if attr.Key == rpcServiceKey {
			actualServiceName = attr.Value.AsString()
			break
		}
	}

	if tc.expectedName != actualServiceName {
		t.Fatalf("invalid service name found. expected %s, actual %s",
			tc.expectedName, actualServiceName)
	}
}
