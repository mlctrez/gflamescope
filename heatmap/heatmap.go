package heatmap

import (
	"bufio"
	"math"
	"strings"

	"github.com/mlctrez/gflamescope/gfutil"
)

const YRATIO float32 = 1000

type Offsets struct {
	Start   float64
	End     float64
	Offsets []float64
}

type HeatMap struct {
	MaxValue int8      `json:"maxvalue"`
	Columns  []int8    `json:"columns"`
	Rows     []float32 `json:"rows"`
	Values   [][]int8  `json:"values"`
}

func GenerateOffsets(scanner *bufio.Scanner) *Offsets {
	// flamescope/app/util/heatmap.py starting line 84
	o := &Offsets{
		Start:   math.Inf(1),
		End:     math.Inf(-1),
		Offsets: make([]float64, 0),
	}

	var stack string

	var ts float64 = -1

	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}

		// match for java  8192 [000]  5076.100000: cpu-clock:
		m := gfutil.EventRegexp.FindStringSubmatch(text)
		if m != nil {
			if stack != "" {
				matchIdle := gfutil.IdleRegexp.FindStringSubmatch(stack)
				if matchIdle == nil {
					o.Offsets = append(o.Offsets, ts)
				}
				stack = ""
			}
			ts = gfutil.MustParseFloat(m[1])
			if ts < o.Start {
				o.Start = ts
			}
			stack = strings.TrimRight(text, " \t\n\r")
		} else {
			stack += strings.TrimRight(text, " \t\n\r")
		}
	}

	if gfutil.IdleRegexp.FindStringSubmatch(stack) == nil {
		o.Offsets = append(o.Offsets, ts)
	}
	if ts > o.End {
		o.End = ts
	}
	return o
}

func GenerateHeatMap(o *Offsets, rows int) *HeatMap {

	h := &HeatMap{}

	h.Rows = make([]float32, 0)

	for i := rows - 1; i > -1; i-- {
		h.Rows = append(h.Rows, YRATIO*(float32(i)/float32(rows)))
	}

	cols := int(math.Ceil(o.End) - math.Floor(o.Start))
	h.Columns = make([]int8, cols)
	for i := range h.Columns {
		h.Columns[i] = int8(i)
	}
	h.Values = make([][]int8, cols)
	for i := range h.Values {
		h.Values[i] = make([]int8, rows)
	}

	for _, ts := range o.Offsets {
		col := int(math.Floor(ts - math.Floor(o.Start)))
		_, f := math.Modf(ts)
		row := rows - int(math.Floor(float64(rows)*float64(f))) - 1
		h.Values[col][row] += 1
		if h.Values[col][row] > h.MaxValue {
			h.MaxValue = h.Values[col][row]
		}
	}

	return h
}
