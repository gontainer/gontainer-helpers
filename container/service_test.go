package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_SetField(t *testing.T) {
	t.Run("Remove duplicates", func(t *testing.T) {
		s := NewService()
		s.SetField("age", NewDependencyValue(30))
		for _, n := range []string{"Jane", "John", "Mary"} {
			s.SetField("name", NewDependencyValue(n))
		}
		s.SetField("eyeColor", NewDependencyValue("blue"))
		require.Len(t, s.fields, 3)
		assert.Equal(t, "age", s.fields[0].name)
		assert.Equal(t, "name", s.fields[1].name)
		assert.Equal(t, "Mary", s.fields[1].dep.value)
		assert.Equal(t, "eyeColor", s.fields[2].name)
	})
}

func TestService_SetFields(t *testing.T) {
	s := NewService()
	s.SetFields(map[string]Dependency{
		"lastname":  NewDependencyValue("Stark"),
		"firstname": NewDependencyValue("Tony"),
	})
	assert.Equal(
		t,
		[]serviceField{
			{
				name: "firstname",
				dep: Dependency{
					type_: dependencyValue,
					value: "Tony",
				},
			},
			{
				name: "lastname",
				dep: Dependency{
					type_: dependencyValue,
					value: "Stark",
				},
			},
		},
		s.fields,
	)
}
