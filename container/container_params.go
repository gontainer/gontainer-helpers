package container

import (
	"errors"
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/grouperror"
)

// GetParam returns a param with the given ID.
func (c *Container) GetParam(paramID string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	c.warmUpGraph()

	return c.getParam(paramID)
}

func (c *Container) getParam(id string) (result any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("getParam(%+q): ", id), err)
		}
	}()

	param, ok := c.params[id]
	if !ok {
		return nil, errors.New("param does not exist")
	}

	c.paramsLockers[id].Lock()
	defer c.paramsLockers[id].Unlock()

	if p, cached := c.cacheParams.get(id); cached {
		return p, nil
	}

	err = c.graphBuilder.paramCircularDeps(id)
	if err != nil {
		return nil, grouperror.Prefix("circular dependencies: ", err)
	}

	result, err = c.resolveDep(nil, param)
	if err != nil {
		return nil, err
	}

	c.cacheParams.set(id, result)

	return result, nil
}
