package container

import (
	"errors"
	"fmt"

	"github.com/gontainer/gontainer-helpers/grouperror"
)

func (c *Container) GetParam(paramID string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.getParam(paramID)
}

func (c *Container) getParam(id string) (result any, err error) {
	c.paramsLockers[id].Lock()
	defer c.paramsLockers[id].Unlock()

	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("getParam(%+q): ", id), err)
		}
	}()

	err = c.graphBuilder.paramCircularDeps(id)
	if err != nil {
		return nil, grouperror.Prefix("circular dependencies: ", err)
	}

	param, ok := c.params[id]
	if !ok {
		return nil, errors.New("param does not exist")
	}

	result, err = c.resolveDep(nil, param)
	if err != nil {
		return nil, err
	}

	c.cacheParams.set(id, result)

	return result, nil
}
