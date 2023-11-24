// Copyright 2023 IAC. All Rights Reserved.

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

package user

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_execLogin(t *testing.T) {
	type args struct {
		ctx         *gin.Context
		username    string
		password    string
		clienttoken string
		ClientID    string
		Renew       bool
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execLogin(tt.args.ctx, tt.args.username, tt.args.password, tt.args.clienttoken, tt.args.ClientID, tt.args.Renew)
		})
	}
}

func Test_getUserImage(t *testing.T) {
	type args struct {
		username string
		clientid string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "ValidUsername",
			args:    args{username: "user"},
			want:    "user",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserImage(tt.args.username, tt.args.clientid)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getUserImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
