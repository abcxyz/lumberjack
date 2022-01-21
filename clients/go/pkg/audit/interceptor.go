// Copyright 2022 Lumberjack authors (see AUTHORS file)
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

package audit

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/security"
)

// Interceptor contains the fields required for an interceptor
// to autofill and emit audit logs.
type Interceptor struct {
	*Client
	SecurityContext security.GRPCContext
	Rules           []*alpb.AuditRule
}

// UnaryInterceptor is a unary interceptor that autofills and emits audit logs.
func (i *Interceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return nil, fmt.Errorf("not yet implemented")
}
