package form

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStatusObj struct {
	Status string `json:"status"`
	Type   string `request:"type"`
}

func TestParse_Successfully(t *testing.T) {
	r := bytes.NewBuffer([]byte(`{"status": "success"}`))

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"http://localhost?type=1",
		r,
	)
	assert.NoError(t, err)

	var obj testStatusObj
	err = json.NewDecoder(req.Body).Decode(&obj)
	assert.NoError(t, err)

	err = req.ParseForm()
	assert.NoError(t, err)
	err = Load(req.Form, &obj)
	assert.NoError(t, err)

	assert.Equal(t, obj.Type, "1")
	assert.Equal(t, obj.Status, "success")
}
