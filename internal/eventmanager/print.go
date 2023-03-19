package eventmanager

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jschwinger233/gofuncgraph/internal/bpf"
	"github.com/jschwinger233/gofuncgraph/internal/uprobe"
)

const (
	placeholder = "        "
)

func (m *EventManager) SprintCallChain(event bpf.GofuncgraphEvent) (chain string, err error) {
	if event.CallerIp == 0 {
		return "", nil
	}
	syms, off, err := m.elf.ResolveAddress(event.CallerIp)
	if err != nil {
		return
	}
	return fmt.Sprintf("%s+%d", syms[0].Name, off), nil
}

func (m *EventManager) PrintStack(StackId uint64) (err error) {
	indent := ""
	fmt.Println()
	startTimeStack := []uint64{}
	var lastEvent bpf.GofuncgraphEvent
	for _, event := range m.goroutine2events[StackId] {
		t := m.bootTime.Add(time.Duration(event.TimeNs)).Format("02 15:04:05.0000")
		syms, offset, err := m.elf.ResolveAddress(event.Ip)
		if err != nil {
			return err
		}

		switch event.Location {
		case 0: // entpoint
			startTimeStack = append(startTimeStack, event.TimeNs)
			callChain, err := m.SprintCallChain(event)
			if err != nil {
				return err
			}
			uprobe, err := m.GetUprobe(event)
			if err != nil {
				return err
			}
			sinceLastEvent := 0.
			if lastEvent.TimeNs != 0 {
				sinceLastEvent = time.Duration(event.TimeNs - lastEvent.TimeNs).Seconds()
			}
			if len(uprobe.FetchArgs) == 0 {
				fmt.Printf("%s %8.4f %s %s+%d { %s\n", t, sinceLastEvent, indent, uprobe.Funcname, uprobe.RelOffset, callChain)
			} else {
				argStrings := []string{}
				data := event.Data[:]
				for _, arg := range uprobe.FetchArgs {
					argString, err := m.SprintArg(arg, data)
					if err != nil {
						return err
					}
					argStrings = append(argStrings, argString)
					data = data[arg.Size:]
				}
				fmt.Printf("%s %8.4f %s %s+%d(%s) { %s\n", t, sinceLastEvent, indent, uprobe.Funcname, uprobe.RelOffset, strings.Join(argStrings, ", "), callChain)
			}
			indent += "  "

		case 1: // retpoint
			if len(indent) == 0 {
				continue
			}
			elapsed := event.TimeNs - startTimeStack[len(startTimeStack)-1]
			startTimeStack = startTimeStack[:len(startTimeStack)-1]
			indent = indent[:len(indent)-2]
			fmt.Printf("%s %8.4f %s } %s+%d\n", t, time.Duration(elapsed).Seconds(), indent, syms[0].Name, offset)

		case 2: // custompoint
			if len(indent) == 0 {
				continue
			}
			uprobe, err := m.GetUprobe(event)
			if err != nil {
				return err
			}
			sinceLastEvent := 0.
			if lastEvent.TimeNs != 0 {
				sinceLastEvent = time.Duration(event.TimeNs - lastEvent.TimeNs).Seconds()
			}
			if len(uprobe.FetchArgs) == 0 {
				fmt.Printf("%s %8.4f %s %s+%d\n", t, sinceLastEvent, indent, uprobe.Funcname, uprobe.RelOffset)
			} else {
				argStrings := []string{}
				data := event.Data[:]
				for _, arg := range uprobe.FetchArgs {
					argString, err := m.SprintArg(arg, data)
					if err != nil {
						return err
					}
					argStrings = append(argStrings, argString)
					data = data[arg.Size:]
				}
				fmt.Printf("%s %8.4f %s %s+%d(%s)\n", t, sinceLastEvent, indent, uprobe.Funcname, uprobe.RelOffset, strings.Join(argStrings, ", "))
			}
		}

		lastEvent = event
	}
	return
}

func (m *EventManager) SprintArg(arg *uprobe.FetchArg, data []uint8) (_ string, err error) {
	value := arg.SprintValue(data)
	if arg.Varname != "__call__" {
		return fmt.Sprintf("%s=%s", arg.Varname, value), nil
	}
	addr, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return "", err
	}
	syms, offset, err := m.elf.ResolveAddress(addr)
	if err != nil {
		return "", err
	}
	if offset != 0 {
		return "", fmt.Errorf("not a valid __call__ target: %lld", addr)
	}
	return fmt.Sprintf("__call__=%s", syms[0].Name), nil
}

func (m *EventManager) PrintRemaining() (err error) {
	for StackId := range m.goroutine2events {
		if err = m.PrintStack(StackId); err != nil {
			break
		}
	}
	return
}
