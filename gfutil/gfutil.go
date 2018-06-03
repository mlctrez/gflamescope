package gfutil

import (
	"fmt"
	"regexp"
	"strconv"
)

var EventRegexp = regexp.MustCompile(` +([0-9.]+): .+?:`)
var FrameRegexp = regexp.MustCompile(`^[\t ]*[0-9a-fA-F]+ (.+) \((.*)\)`)
var CommRegexp = regexp.MustCompile(`^ *([^0-9]+)`)
var idleProcess = "swapper"
var idleStack = "(cpuidle|cpu_idle|cpu_bringup_and_idle|native_safe_halt|xen_hypercall_sched_op|xen_hypercall_vcpu_op)"
var IdleRegexp = regexp.MustCompile(fmt.Sprintf("%s.*%s", idleProcess, idleStack))

func MustParseFloat(in string) float64 {
	result, err := strconv.ParseFloat(in, 64)
	if err != nil {
		panic(err)
	}
	return result
}
