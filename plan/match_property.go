// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package plan

import (
	"math"
)

func (ts *PhysicalTableScan) MatchProperty(prop requiredProperty, rowCounts []uint64, _ ...*responseProperty) *responseProperty {
	rowCount := float64(rowCounts[0])
	cost := rowCount * netWorkFactor
	if len(prop) == 0 {
		return &responseProperty{p: ts, cost: cost}
	}
	if len(prop) == 1 && ts.pkCol != nil && ts.pkCol == prop[0].col {
		sortedTs := *ts
		sortedTs.Desc = prop[0].desc
		return &responseProperty{p: &sortedTs, cost: cost}
	}
	return &responseProperty{p: ts, cost: math.MaxFloat64}
}

func (is *PhysicalIndexScan) MatchProperty(prop requiredProperty, rowCounts []uint64, _ ...*responseProperty) *responseProperty {
	rowCount := float64(rowCounts[0])
	// currently index read from kv 2 times.
	cost := rowCount * netWorkFactor * 2
	if len(prop) == 0 {
		return &responseProperty{p: is, cost: cost}
	}
	matched := 0
	allDesc, allAsc := true, true
	for i, indexCol := range is.Index.Columns {
		if prop[matched].col.ColName.L != indexCol.Name.L {
			if matched != 0 || i >= is.accessEqualCount {
				break
			}
			continue
		}
		matched++
		if prop[matched].desc {
			allAsc = false
		} else {
			allDesc = false
		}
		if matched == len(prop) {
			break
		}
	}
	if matched == len(prop) {
		sortedCost := cost + rowCount*math.Log2(rowCount)
		if allDesc {
			sortedIs := *is
			sortedIs.Desc = true
			sortedIs.OutOfOrder = false
			return &responseProperty{p: &sortedIs, cost: sortedCost}
		}
		if allAsc {
			sortedIs := *is
			sortedIs.OutOfOrder = false
			return &responseProperty{p: &sortedIs, cost: sortedCost}
		}
	}
	return &responseProperty{p: is, cost: math.MaxFloat64}
}

func (p *PhysicalHashSemiJoin) MatchProperty(prop requiredProperty, _ []uint64, response ...*responseProperty) *responseProperty {
	lRes, rRes := response[0], response[1]
	np := *p
	np.SetChildren(lRes.p, rRes.p)
	cost := lRes.cost + rRes.cost
	return &responseProperty{p: &np, cost: cost}
}

func (p *PhysicalApply) MatchProperty(prop requiredProperty, rowCounts []uint64, response ...*responseProperty) *responseProperty {
	np := *p
	np.SetChildren(response[0].p)
	return &responseProperty{p: &np, cost: response[0].cost}
}

func (p *PhysicalHashJoin) MatchProperty(prop requiredProperty, rowCounts []uint64, response ...*responseProperty) *responseProperty {
	lRes, rRes := response[0], response[1]
	lCount, rCount := float64(rowCounts[0]), float64(rowCounts[1])
	np := *p
	np.SetChildren(lRes.p, rRes.p)
	cost := lRes.cost + rRes.cost
	if p.smallTable == 1 {
		cost += lCount + memoryFactor*rCount
	} else {
		cost += rCount + memoryFactor*lCount
	}
	return &responseProperty{p: &np, cost: cost}
}

func (p *NewUnion) MatchProperty(prop requiredProperty, _ []uint64, response ...*responseProperty) *responseProperty {
	np := *p
	children := make([]Plan, 0, len(response))
	cost := float64(0)
	for _, res := range response {
		children = append(children, res.p)
		cost += res.cost
	}
	np.SetChildren(children...)
	return &responseProperty{p: &np, cost: cost}
}

func (p *Selection) MatchProperty(prop requiredProperty, rowCounts []uint64, response ...*responseProperty) *responseProperty {
	if len(response) == 0 {
		res := p.GetChildByIndex(0).(PhysicalPlan).MatchProperty(prop, nil, nil)
		sel := *p
		sel.SetChildren(res.p)
		res.p = &sel
		return res
	}
	np := *p
	np.SetChildren(response[0].p)
	return &responseProperty{p: &np, cost: response[0].cost}
}

func (p *Projection) MatchProperty(_ requiredProperty, _ []uint64, response ...*responseProperty) *responseProperty {
	np := *p
	np.SetChildren(response[0].p)
	return &responseProperty{p: &np, cost: response[0].cost}
}

func (p *MaxOneRow) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *Exists) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *Trim) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *Aggregation) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *Limit) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *Distinct) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *NewTableDual) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *NewSort) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *Insert) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}

func (p *SelectLock) MatchProperty(_ requiredProperty, _ []uint64, _ ...*responseProperty) *responseProperty {
	panic("You can't call this function!")
	return nil
}