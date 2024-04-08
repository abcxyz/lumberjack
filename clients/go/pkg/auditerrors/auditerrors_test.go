// Copyright 2024 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auditerrors

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/abcxyz/pkg/testutil"
)

func TestInterceptorError(t *testing.T) {
	t.Parallel()

	errMsg := "test error"
	grpcErr := status.Error(codes.FailedPrecondition, errMsg)

	gotErr := InterceptorError(grpcErr)
	if !errors.Is(gotErr, ErrInterceptor) {
		t.Errorf("expected error %v to be wrapped with %v", gotErr, ErrInterceptor)
	}
	if status.Code(gotErr) != codes.FailedPrecondition {
		t.Errorf("expected error %v to have code %v", gotErr, codes.FailedPrecondition)
	}
	if diff := testutil.DiffErrString(gotErr, errMsg); diff != "" {
		t.Errorf("unexpected error message:\n%s", diff)
	}
}
