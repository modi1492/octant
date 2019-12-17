/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package component

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminal_Marshall(t *testing.T) {
	details := TerminalDetails{
		Container: "container-id",
		Command:   "/bin/bash",
		UUID:      "0000-0000-0000-0000-0000",
		Active:    false,
	}
	input := NewTerminal("default", "term-test", details)
	actual, err := json.Marshal(input)
	assert.NoError(t, err)

	var expected = `
            {
                "metadata": {
									"type": "terminal",
									"title": [
										{
											"config": { "value": "default / term-test" },
											"metadata": { "type": "text" }
										}
									]
                },
                "config": {
                  	"name": "term-test",
                  	"namespace": "default",
					"terminal": {
						"active": false,
						"command": "/bin/bash",
						"container": "container-id",
						"createdAt": "0001-01-01T00:00:00Z",
						"uuid": "0000-0000-0000-0000-0000"
                    }
                }
            }
`
	assert.JSONEq(t, expected, string(actual))
}
