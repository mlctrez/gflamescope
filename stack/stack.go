package stack

import (
	"bufio"
	"math"
	"os"
	"strings"

	"github.com/mlctrez/gflamescope/gfutil"
)

type Frame struct {
	Instruction string
	Library     string
}

type CallData struct {
	Label   string      `json:"l"`
	Name    string      `json:"n"`
	Calls   []*CallData `json:"c"`
	Samples int         `json:"v"`
}

func CalculateStackRange(scanner *bufio.Scanner) (start, end float64) {

	start = math.Inf(1)
	end = math.Inf(-1)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "\t") {
			continue
		}
		matchEvent := gfutil.EventRegexp.FindStringSubmatch(line)
		if matchEvent != nil {
			ts := gfutil.MustParseFloat(matchEvent[1])
			if ts < start {
				start = ts
			} else if ts > end {
				end = ts
			}
		}
	}
	// this is what the original codebase does at
	// flamescope/app/util/stack.py:115
	start = math.Floor(start)
	end = math.Ceil(end)
	return
}

func CreateFlameGraph(file *os.File, start, end float64) *CallData {

	// this has none of the original codebase performance optimizations, yet
	// TODO: combine /proc/uptime with offsets in perf file for exact timestamps?

	var stack []*Frame

	var ts float64
	var comm string

	root := &CallData{}
	root.Name = "root"
	root.Calls = []*CallData{}
	root.Label = ""
	root.Samples = 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		var eventMatch []string

		if !strings.HasPrefix(line, "\t") {
			eventMatch = gfutil.EventRegexp.FindStringSubmatch(line)
		}

		if eventMatch != nil {
			if len(stack) > 0 {
				searchIdle := ""
				for _, sf := range stack {
					searchIdle += sf.Instruction + ";"
				}
				if gfutil.IdleRegexp.FindStringSubmatch(searchIdle) != nil {
					stack = []*Frame{}
				} else if ts >= start && ts <= end {
					root = addStack(root, stack, comm)
				}
				stack = []*Frame{}
			}
			ts = gfutil.MustParseFloat(eventMatch[1])
			if ts > (end + 0.1) {
				break
			}
			commMatch := gfutil.CommRegexp.FindStringSubmatch(line)
			if commMatch != nil {
				comm = strings.TrimRight(commMatch[1], " \t\n\r")
				stack = append(stack, &Frame{Instruction: comm, Library: ""})
			} else {
				stack = append(stack, &Frame{Instruction: "<unknown>", Library: ""})
			}
		} else {
			frameMatch := gfutil.FrameRegexp.FindStringSubmatch(line)
			if frameMatch != nil {
				name := frameMatch[1]
				if i := strings.Index(name, "+"); i > 0 {
					name = name[:i]
				}
				if i := strings.Index(name, "("); i > 0 {
					name = name[:i]
				}
				// https://github.com/golang/go/wiki/SliceTricks
				stack = append(stack, nil)
				copy(stack[2:], stack[1:])
				stack[1] = &Frame{Instruction: name, Library: frameMatch[2]}
			}
		}
	}
	if ts >= start && ts <= end {
		root = addStack(root, stack, comm)
	}

	return root

}

func addStack(root *CallData, stack []*Frame, comm string) *CallData {
	root.Samples++
	last := root
	for _, frame := range stack {
		names := strings.Split(frame.Instruction, "->")
		n := 0
		for _, name := range names {
			if comm == "java" && strings.HasPrefix(name, "L") {
				name = name[1:]
			}
			libraryType := "inlined"
			if n == 0 {
				libraryType = libraryToType(frame.Library)
			}
			n++
			var found = false
			for _, child := range last.Calls {
				if child.Name == name && child.Label == libraryType {
					last = child
					found = true
					break
				}
			}
			if found {
				last.Samples++
			} else {
				newCall := &CallData{
					Name:    name,
					Label:   libraryType,
					Samples: 1,
					Calls:   []*CallData{},
				}
				last.Calls = append(last.Calls, newCall)
				last = newCall
			}
		}

	}
	return root
}

func libraryToType(library string) string {
	if library == "" {
		return ""
	}
	if strings.HasPrefix(library, "/tmp/perf-") {
		return "jit"
	}
	if strings.HasPrefix(library, "[") {
		return "kernel"
	}
	if strings.Contains(library, "vmlinux") {
		return "kernel"
	}
	return "user"
}
