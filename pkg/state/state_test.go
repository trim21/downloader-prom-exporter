package state

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	t.Parallel()

	assert.True(t, PausedUploading&BasicUploading == BasicUploading)
	assert.True(t, StalledUploading&BasicStalled == BasicStalled)
}

type S struct {
	State *State `json:"state"`
}

func TestPausedUploading(t *testing.T) {
	t.Parallel()

	var s = S{}

	assert.Nil(t, json.Unmarshal([]byte(`{"state":"pausedUP"}`), &s))

	assert.Equal(t, *s.State, PausedUploading)
}

func TestUploading(t *testing.T) {
	t.Parallel()

	var s = S{}

	assert.Nil(t, json.Unmarshal([]byte(`{"state":"uploading"}`), &s))

	assert.Equal(t, *s.State, Uploading)
}

func TestError(t *testing.T) {
	t.Parallel()

	var s = S{}

	assert.NotNil(t, json.Unmarshal([]byte(`{"state":"e"}`), &s))
}
