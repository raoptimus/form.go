// Copyright 2023 Urvantsev Evgenii. All rights reserved.
// Use of this source code is governed by a BSD3-style
// license that can be found in the LICENSE file.

package form

func Load(data map[string][]string, v any) error {
	var d decodeState
	d.init(data)
	return d.parse(v)
}
