package container_test

import (
	"log"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
	"github.com/gontainer/gontainer-helpers/errors"
	assertErr "github.com/gontainer/gontainer-helpers/errors/assert"
	"github.com/stretchr/testify/assert"
)

func Test_container_executeServiceCalls(t *testing.T) {
	s := container.NewService()
	s.SetValue(struct{}{})
	s.AppendCall("SetName", container.NewDependencyProvider(func() (interface{}, error) {
		return nil, errors.New("could not fetch the name from the config")
	}))
	s.AppendCall("SetAge", container.NewDependencyValue(21))
	s.AppendCall("SetColor", container.NewDependencyValue("red"))
	s.AppendWither("WithLogger", container.NewDependencyValue(log.Default()))
	// this call will be ignored, because it's after the error returned by a wither
	s.AppendCall("SetLanguage", container.NewDependencyValue("en"))

	c := container.NewContainer()
	c.OverrideService("service", s)

	expected := []string{
		"container.get(\"service\"): resolve args \"SetName\": arg #0: could not fetch the name from the config",
		"container.get(\"service\"): call \"SetAge\": invalid func `*interface {}`.\"SetAge\"",
		"container.get(\"service\"): call \"SetColor\": invalid func `*interface {}`.\"SetColor\"",
		"container.get(\"service\"): wither \"WithLogger\": invalid wither `struct {}`.\"WithLogger\"",
	}

	svc, err := c.Get("service")
	assert.Nil(t, svc)
	assertErr.EqualErrorGroup(t, err, expected)
}
