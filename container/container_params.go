package container

import (
	"errors"
	"fmt"

	"github.com/gontainer/gontainer-helpers/caller"
	"github.com/gontainer/gontainer-helpers/grouperror"
)

func (c *container) GetParam(paramID string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.getParam(paramID)
}

func (c *container) getParam(id string) (result any, err error) {
	c.paramsLockers[id].Lock()
	defer c.paramsLockers[id].Unlock()

	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("paramContainer.GetParam(%+q): ", id), err)
		}
	}()

	// TODO check circular deps

	param, ok := c.params[id]
	if !ok {
		return nil, errors.New("param does not exist")
	}

	switch param.type_ {
	case dependencyValue:
		result = param.value
	case dependencyProvider:
		result, err = caller.CallProvider(param.provider, nil, convertParams)
		if err != nil {
			return nil, err
		}
	case dependencyParam:
		result, err = c.getParam(id)
		if err != nil {
			return nil, err
		}
	}

	c.cacheParams.set(id, result)

	return result, nil
}
