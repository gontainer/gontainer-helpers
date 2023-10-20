package container

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gontainer/gontainer-helpers/caller"
	"github.com/gontainer/gontainer-helpers/grouperror"
)

// TODO: remove it
type paramContainer struct {
	params       map[string]Dependency
	cachedParams keyValue
	rwlocker     rwlocker
	lockers      map[string]sync.Locker
}

// NewParamContainer creates a concurrent-safe container.
func NewParamContainer() *paramContainer {
	return &paramContainer{
		params:       make(map[string]Dependency),
		cachedParams: newSafeMap(),
		rwlocker:     &sync.RWMutex{},
		lockers:      make(map[string]sync.Locker),
	}
}

func (p *paramContainer) OverrideParam(paramID string, d Dependency) {
	p.rwlocker.Lock()
	defer p.rwlocker.Unlock()

	switch d.type_ {
	case
		dependencyValue,
		dependencyParam,
		dependencyProvider:
	default:
		panic(fmt.Sprintf("paramContainer.OverrideParam does not accept `%s`", d.type_.String()))
	}

	p.params[paramID] = d
	p.cachedParams.delete(paramID)
	p.lockers[paramID] = &sync.Mutex{}
}

func (p *paramContainer) GetParam(id string) (result any, err error) {
	p.rwlocker.RLock()
	defer p.rwlocker.RUnlock()

	return p.getParam(id)
}

func (p *paramContainer) getParam(id string) (result any, err error) {
	p.lockers[id].Lock()
	defer p.lockers[id].Unlock()

	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("paramContainer.GetParam(%+q): ", id), err)
		}
	}()

	param, ok := p.params[id]
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
	case dependencyParam: // we don't need to handle circular deps here, parent container handles that
		result, err = p.getParam(id)
		if err != nil {
			return nil, err
		}
	}

	p.cachedParams.set(id, result)

	return result, nil
}
