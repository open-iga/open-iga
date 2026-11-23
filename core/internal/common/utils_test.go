package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeSliceLike(t *testing.T) {
	var anonymous []struct{ Name string }

	made := MakeSliceLike(anonymous, 2)
	made[1].Name = "second"

	assert.Len(t, made, 2)
	assert.Equal(t, "second", made[1].Name)
}
