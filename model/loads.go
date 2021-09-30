package model

import (
	"encoding/json"
)

type CaseMapping struct {
	Dead []string
	Live []string
	Snow []string
	Wind []string
}

func (c *CaseMapping) DeadCases(loadGroups ...string) *CaseMapping {
	c.Dead = loadGroups
	return c
}

func (c *CaseMapping) LiveCases(loadGroups ...string) *CaseMapping {
	c.Live = loadGroups
	return c
}

func (c *CaseMapping) SnowCases(loadGroups ...string) *CaseMapping {
	c.Snow = loadGroups
	return c
}

func (c *CaseMapping) WindCases(loadGroups ...string) *CaseMapping {
	c.Wind = loadGroups
	return c
}

type Case struct {
	Name string
	Dead float64
	Live float64
	Snow float64
	Wind float64
}

type Combination struct {
	Mapping CaseMapping
	Cases   []Case
}

func (a *Combination) MarshalJSON() ([]byte, error) {
	combo := make(map[int]map[string]interface{})

	for i, ca := range a.Cases {
		l := make(map[string]interface{})

		l["name"] = ca.Name
		for _, m := range a.Mapping.Dead {
			l[m] = ca.Dead
		}
		for _, m := range a.Mapping.Live {
			l[m] = ca.Live
		}
		for _, m := range a.Mapping.Snow {
			l[m] = ca.Snow
		}
		for _, m := range a.Mapping.Wind {
			l[m] = ca.Wind
		}

		combo[i+1] = l
	}

	return json.Marshal(combo)
}
