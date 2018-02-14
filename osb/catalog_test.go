package osb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindService(t *testing.T) {

	c := &Catalog{
		Services: []*Service{
			{
				ID: "service1",
			},
			{
				ID: "service2",
			},
		},
	}

	t.Run("Exists service", func(t *testing.T) {
		s, ok := c.FindService("service1")

		assert.True(t, ok)
		assert.NotNil(t, s)
		assert.Equal(t, "service1", s.ID)
	})

	t.Run("Not exists service", func(t *testing.T) {
		s, ok := c.FindService("foobar")

		assert.False(t, ok)
		assert.Nil(t, s)
	})
}
