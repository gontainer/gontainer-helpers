// Copyright (c) 2023 Bartłomiej Krukowski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package container

import (
	"context"
	"errors"
	"fmt"

	"github.com/gontainer/gontainer-helpers/v3/grouperror"
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

	result, err = c.resolveDep(context.Background(), nil, param)
	if err != nil {
		return nil, err
	}

	c.cacheParams.set(id, result)

	return result, nil
}
