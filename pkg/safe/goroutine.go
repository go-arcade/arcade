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

package safe

import (
	"fmt"
	"runtime/debug"
)

// Go starts a new goroutine to run the given function f safely.
func Go(f func()) {
	go Do(f)
}

// Do runs the given function f and recovers from any panic, printing the stack trace.
func Do(f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic: %v\n", r)
			debug.PrintStack()
		}
	}()
	f()
}
