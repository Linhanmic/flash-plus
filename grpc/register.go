/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package grpc

import (
	"context"

	gm "github.com/getgauge/flash/gauge_messages"
	"google.golang.org/grpc"
)

func registerReporter(s *grpc.Server, h *handler) {
	s.RegisterService(&reporterServiceDesc, h)
}

func decodeCall[T any](dec func(interface{}) error, in *T, call func(*T) (*gm.Empty, error)) (interface{}, error) {
	if err := dec(in); err != nil {
		return nil, err
	}
	return call(in)
}

var reporterServiceDesc = grpc.ServiceDesc{
	ServiceName: "gauge.messages.Reporter",
	HandlerType: (*gm.ReporterServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "NotifyExecutionStarting", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.ExecutionStartingRequest), func(in *gm.ExecutionStartingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifyExecutionStarting(ctx, in)
			})
		}},
		{MethodName: "NotifySpecExecutionStarting", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.SpecExecutionStartingRequest), func(in *gm.SpecExecutionStartingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifySpecExecutionStarting(ctx, in)
			})
		}},
		{MethodName: "NotifyScenarioExecutionStarting", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.ScenarioExecutionStartingRequest), func(in *gm.ScenarioExecutionStartingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifyScenarioExecutionStarting(ctx, in)
			})
		}},
		{MethodName: "NotifyStepExecutionStarting", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.StepExecutionStartingRequest), func(in *gm.StepExecutionStartingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifyStepExecutionStarting(ctx, in)
			})
		}},
		{MethodName: "NotifyStepExecutionEnding", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.StepExecutionEndingRequest), func(in *gm.StepExecutionEndingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifyStepExecutionEnding(ctx, in)
			})
		}},
		{MethodName: "NotifyConceptExecutionStarting", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.StepExecutionStartingRequest), func(in *gm.StepExecutionStartingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifyConceptExecutionStarting(ctx, in)
			})
		}},
		{MethodName: "NotifyConceptExecutionEnding", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.StepExecutionEndingRequest), func(in *gm.StepExecutionEndingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifyConceptExecutionEnding(ctx, in)
			})
		}},
		{MethodName: "NotifyScenarioExecutionEnding", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.ScenarioExecutionEndingRequest), func(in *gm.ScenarioExecutionEndingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifyScenarioExecutionEnding(ctx, in)
			})
		}},
		{MethodName: "NotifySpecExecutionEnding", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.SpecExecutionEndingRequest), func(in *gm.SpecExecutionEndingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifySpecExecutionEnding(ctx, in)
			})
		}},
		{MethodName: "NotifyExecutionEnding", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.ExecutionEndingRequest), func(in *gm.ExecutionEndingRequest) (*gm.Empty, error) {
				return srv.(*handler).NotifyExecutionEnding(ctx, in)
			})
		}},
		{MethodName: "NotifySuiteResult", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.SuiteExecutionResult), func(in *gm.SuiteExecutionResult) (*gm.Empty, error) {
				return srv.(*handler).NotifySuiteResult(ctx, in)
			})
		}},
		{MethodName: "Kill", Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
			return decodeCall(dec, new(gm.KillProcessRequest), func(in *gm.KillProcessRequest) (*gm.Empty, error) {
				return srv.(*handler).Kill(ctx, in)
			})
		}},
	},
}
