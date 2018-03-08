package osb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPlan(t *testing.T) {

	s := &Service{
		Plans: []*Plan{
			{
				ID: "plan1",
			},
			{
				ID: "plan2",
			},
		},
	}

	t.Run("Exists plan", func(t *testing.T) {
		p, ok := s.FindPlan("plan1")

		assert.True(t, ok)
		assert.NotNil(t, p)
		assert.Equal(t, "plan1", p.ID)
	})

	t.Run("Not exists service", func(t *testing.T) {
		p, ok := s.FindPlan("foobar")

		assert.False(t, ok)
		assert.Nil(t, p)
	})
}
