package container

import (
	"fmt"
	"sync"

	"github.com/gontainer/gontainer-helpers/caller"
	"github.com/gontainer/gontainer-helpers/errors"
)

type paramContainer struct {
	params       map[string]Dependency
	cachedParams map[string]interface{}
	rwlocker     rwlocker
	lockers      map[string]sync.Locker
}

func NewParamContainer() *paramContainer {
	return &paramContainer{
		params:       make(map[string]Dependency),
		cachedParams: make(map[string]interface{}),
		rwlocker:     &sync.RWMutex{},
		lockers:      make(map[string]sync.Locker),
	}
}

func (p *paramContainer) OverrideParam(id string, d Dependency) {
	p.rwlocker.Lock()
	defer p.rwlocker.Unlock()

	switch d.type_ {
	case
		dependencyNil,
		dependencyProvider:
	default:
		panic(fmt.Sprintf("paramContainer.OverrideParam does not accept `%s`", d.type_.String()))
	}

	p.params[id] = d
	delete(p.cachedParams, id)
	p.lockers[id] = &sync.Mutex{}
}

func (p *paramContainer) GetParam(id string) (result interface{}, err error) {
	p.rwlocker.RLock()
	defer p.rwlocker.RUnlock()

	p.lockers[id].Lock()
	defer p.lockers[id].Unlock()

	defer func() {
		if err != nil {
			err = errors.PrefixedGroup(fmt.Sprintf("paramContainer.GetParam(%+q): ", id), err)
		}
	}()

	param, ok := p.params[id]
	if !ok {
		return nil, errors.New("param does not exist")
	}

	switch param.type_ {
	case dependencyNil:
		result = param.value
	case dependencyProvider:
		result, err = caller.CallProvider(param.provider)
		if err != nil {
			return nil, err
		}
	}

	p.cachedParams[id] = result

	return result, nil
}
