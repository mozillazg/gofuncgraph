package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"

	"github.com/jschwinger233/gofuncgraph/elf"
	"github.com/jschwinger233/gofuncgraph/internal/bpf"
	"github.com/jschwinger233/gofuncgraph/internal/eventmanager"
	"github.com/jschwinger233/gofuncgraph/internal/uprobe"
	log "github.com/sirupsen/logrus"
)

var (
	OffsetPattern *regexp.Regexp
)

func init() {
	OffsetPattern = regexp.MustCompile(`\+\d+$`)
}

type Tracer struct {
	bin       string
	elf       *elf.ELF
	args      []string
	backtrace bool
	depth     int

	bpf *bpf.BPF
}

func NewTracer(bin string, args []string, backtrace bool, depth int) (_ *Tracer, err error) {
	elf, err := elf.New(bin)
	if err != nil {
		return
	}

	return &Tracer{
		bin:       bin,
		elf:       elf,
		args:      args,
		backtrace: backtrace,
		depth:     depth,

		bpf: bpf.New(),
	}, nil
}

func (t *Tracer) ParseArgs(inputs []string) (in, ex []string, fetch map[string]map[string]string, offsets map[string][]uint64, err error) {
	fetch = map[string]map[string]string{}
	offsets = map[string][]uint64{}
	for _, input := range inputs {
		if input[len(input)-1] == ')' {
			stack := []byte{')'}
			for i := len(input) - 2; i >= 0; i-- {
				if input[i] == ')' {
					stack = append(stack, ')')
				} else if input[i] == '(' {
					if len(stack) > 0 && stack[len(stack)-1] == ')' {
						stack = stack[:len(stack)-1]
					} else {
						err = fmt.Errorf("imbalanced parenthese: %s", input)
						return
					}
				}

				if len(stack) == 0 {
					funcname := input[:i]
					fetch[funcname] = map[string]string{}
					for _, part := range strings.Split(input[i+1:len(input)-1], ",") {
						varState := strings.Split(part, "=")
						if len(varState) != 2 {
							err = fmt.Errorf("invalid variable statement: %s", varState)
							return
						}
						fetch[funcname][strings.TrimSpace(varState[0])] = strings.TrimSpace(varState[1])
					}
					input = input[:i]
					break
				}
			}
			if len(stack) > 0 {
				err = fmt.Errorf("imbalanced parenthese: %s", input)
				return
			}
		}

		if OffsetPattern.MatchString(input) {
			idx := OffsetPattern.FindAllStringIndex(input, -1)[0][0]
			offset, e := strconv.ParseUint(input[idx+1:len(input)], 10, 64)
			if e != nil {
				err = fmt.Errorf("invalid custom offset: %s", input[idx+1:len(input)])
				return
			}
			offsets[input[:idx]] = append(offsets[input[:idx]], offset)
			input = input[:idx]
		}

		if input[0] == '!' {
			ex = append(ex, input[1:])
		} else {
			in = append(in, input)
		}
	}
	return
}

func (t *Tracer) Start() (err error) {
	in, ex, fetch, offsets, err := t.ParseArgs(t.args)
	if err != nil {
		return
	}
	uprobes, err := uprobe.Parse(t.elf, &uprobe.ParseOptions{
		Wildcards:     in,
		ExWildcards:   ex,
		Fetch:         fetch,
		CustomOffsets: offsets,
		SearchDepth:   t.depth,
		Backtrace:     t.backtrace,
	})
	if err != nil {
		return
	}
	log.Infof("found %d uprobes\n", len(uprobes))

	if err = t.bpf.Load(uprobes); err != nil {
		return
	}
	if err = t.bpf.Attach(t.bin, uprobes); err != nil {
		return
	}

	defer t.bpf.Detach()
	log.Info("start tracing\n")

	eventManager, err := eventmanager.New(uprobes, t.elf)
	if err != nil {
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for event := range t.bpf.PollEvents(ctx) {
		if err = eventManager.Handle(event); err != nil {
			return
		}
	}
	return eventManager.PrintRemaining()
}
