package registry

import (
	"fmt"
	"sync"
)

type Agent struct {
	Id     string
	Name   string
	Skills []Skill
	Traits map[string]string
}

type Skill struct {
	Name   string
	Type   string
	Traits map[string]string
}

var (
	buddies = make(map[string]*Agent)
	lock    = sync.RWMutex{}
)

func Register(buddy *Agent) error {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := buddies[buddy.Id]; ok {
		return fmt.Errorf("There already exists buddy'%s'", buddy.Name)
	}

	buddies[buddy.Id] = buddy
	return nil
}

func Unregister(id string) {
	lock.Lock()
	delete(buddies, id)
	lock.Unlock()
}

func GetById(id string) (*Agent, error) {
	lock.RLock()
	defer lock.RUnlock()

	buddy, ok := buddies[id]

	if !ok {
		return &Agent{}, fmt.Errorf("No buddy with id '%s' found", id)
	}

	return buddy, nil
}

func All() []*Agent {
	lock.RLock()
	defer lock.RUnlock()

	result := make([]*Agent, 0, len(buddies))

	for _, b := range buddies {
		result = append(result, b)
	}
	return result
}
