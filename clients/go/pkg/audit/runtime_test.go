// Copyright 2021 Lumberjack authors (see AUTHORS file)
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
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"

	alpb "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
	"github.com/abcxyz/lumberjack/clients/go/pkg/testutil"
)

func TestRuntimeInfo_Process(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		r          *runtimeInfo
		logReq     *alpb.AuditLogRequest
		wantLogReq *alpb.AuditLogRequest
	}{
		{
			name: "should_write_monitored_resource_to_payload_metadata",
			r: &runtimeInfo{
				structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type": structpb.NewStringValue("gce_instance"),
						"labels": structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"instanceId": structpb.NewStringValue("testID"),
								"zone":       structpb.NewStringValue("testZone"),
							},
						}),
					},
				}),
			},
			logReq: testutil.ReqBuilder().Build(),
			wantLogReq: testutil.ReqBuilder().WithMetadata(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"originating_resource": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"type": structpb.NewStringValue("gce_instance"),
							"labels": structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"instanceId": structpb.NewStringValue("testID"),
									"zone":       structpb.NewStringValue("testZone"),
								},
							}),
						},
					}),
				},
			}).Build(),
		},
		{
			name: "should_append_monitored_resources_to_payload_metadata",
			r: &runtimeInfo{
				structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type": structpb.NewStringValue("gce_instance"),
						"labels": structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"instanceId": structpb.NewStringValue("testID"),
								"zone":       structpb.NewStringValue("testZone"),
							},
						}),
					},
				}),
			},
			logReq: testutil.ReqBuilder().WithMetadata(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"existing_key": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"existing_subkey": structpb.NewStringValue("existing_val"),
						},
					}),
				},
			}).Build(),
			wantLogReq: testutil.ReqBuilder().WithMetadata(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"existing_key": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"existing_subkey": structpb.NewStringValue("existing_val"),
						},
					}),
					"originating_resource": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"type": structpb.NewStringValue("gce_instance"),
							"labels": structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"instanceId": structpb.NewStringValue("testID"),
									"zone":       structpb.NewStringValue("testZone"),
								},
							}),
						},
					}),
				},
			}).Build(),
		},
		{
			name: "nil_runtimeinfo_should_leave_metadata_untouched",
			logReq: testutil.ReqBuilder().WithMetadata(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"existing_key": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"existing_subkey": structpb.NewStringValue("existing_val"),
						},
					}),
				},
			}).Build(),
			wantLogReq: testutil.ReqBuilder().WithMetadata(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"existing_key": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"existing_subkey": structpb.NewStringValue("existing_val"),
						},
					}),
				},
			}).Build(),
		},
		{
			name: "nil_monitored_resource_should_leave_metadata_untouched",
			r:    &runtimeInfo{},
			logReq: testutil.ReqBuilder().WithMetadata(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"existing_key": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"existing_subkey": structpb.NewStringValue("existing_val"),
						},
					}),
				},
			}).Build(),
			wantLogReq: testutil.ReqBuilder().WithMetadata(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"existing_key": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"existing_subkey": structpb.NewStringValue("existing_val"),
						},
					}),
				},
			}).Build(),
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.r.Process(context.Background(), tc.logReq)
			if err != nil {
				t.Errorf("Process(%+v) error unexpected error: %v", tc.logReq, err)
			}

			if diff := cmp.Diff(tc.wantLogReq, tc.logReq, protocmp.Transform()); diff != "" {
				t.Errorf("Process(%+v) got diff (-want, +got): %v", tc.logReq, diff)
			}
		})
	}
}
