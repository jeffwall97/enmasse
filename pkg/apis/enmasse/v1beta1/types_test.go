/*
 * Copyright 2019, EnMasse authors.
 * License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
 */

package v1beta1

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSerialize(t *testing.T) {
	addressSpace := AddressSpace{}
	bytes, err := json.Marshal(addressSpace)
	if err != nil {
		t.Fatalf("Failed to serialize")
		t.Fail()
		return
	}

	fmt.Println(string(bytes))
}
