// Copyright 2025 Arcade Team
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

package util

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

var StateStore = &sync.Map{}

// GenState generate a random state string
func GenState() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// LoadAndDeleteState load and delete the state from the state store
func LoadAndDeleteState(state string) (string, bool) {
	value, ok := StateStore.LoadAndDelete(state)
	if ok {
		return value.(string), true
	}
	return "", false
}
