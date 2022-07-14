// Code generated by bpf2go; DO NOT EDIT.
//go:build 386 || amd64
// +build 386 amd64

package bpf

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

type UfuncgraphEvent struct {
	StackId    uint64
	CallerIp   uint64
	Ip         uint64
	TimeNs     uint64
	StackDepth uint32
	HookPoint  uint16
	Errno      uint16
	Args       [104]uint8
}

// LoadUfuncgraph returns the embedded CollectionSpec for Ufuncgraph.
func LoadUfuncgraph() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_UfuncgraphBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load Ufuncgraph: %w", err)
	}

	return spec, err
}

// LoadUfuncgraphObjects loads Ufuncgraph and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//     *UfuncgraphObjects
//     *UfuncgraphPrograms
//     *UfuncgraphMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func LoadUfuncgraphObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := LoadUfuncgraph()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// UfuncgraphSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type UfuncgraphSpecs struct {
	UfuncgraphProgramSpecs
	UfuncgraphMapSpecs
}

// UfuncgraphSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type UfuncgraphProgramSpecs struct {
	OnEntry       *ebpf.ProgramSpec `ebpf:"on_entry"`
	OnEntryGolang *ebpf.ProgramSpec `ebpf:"on_entry_golang"`
	OnExit        *ebpf.ProgramSpec `ebpf:"on_exit"`
	OnExitGolang  *ebpf.ProgramSpec `ebpf:"on_exit_golang"`
}

// UfuncgraphMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type UfuncgraphMapSpecs struct {
	BpToEvent  *ebpf.MapSpec `ebpf:"bp_to_event"`
	EventQueue *ebpf.MapSpec `ebpf:"event_queue"`
	Goids      *ebpf.MapSpec `ebpf:"goids"`
}

// UfuncgraphObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to LoadUfuncgraphObjects or ebpf.CollectionSpec.LoadAndAssign.
type UfuncgraphObjects struct {
	UfuncgraphPrograms
	UfuncgraphMaps
}

func (o *UfuncgraphObjects) Close() error {
	return _UfuncgraphClose(
		&o.UfuncgraphPrograms,
		&o.UfuncgraphMaps,
	)
}

// UfuncgraphMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to LoadUfuncgraphObjects or ebpf.CollectionSpec.LoadAndAssign.
type UfuncgraphMaps struct {
	BpToEvent  *ebpf.Map `ebpf:"bp_to_event"`
	EventQueue *ebpf.Map `ebpf:"event_queue"`
	Goids      *ebpf.Map `ebpf:"goids"`
}

func (m *UfuncgraphMaps) Close() error {
	return _UfuncgraphClose(
		m.BpToEvent,
		m.EventQueue,
		m.Goids,
	)
}

// UfuncgraphPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to LoadUfuncgraphObjects or ebpf.CollectionSpec.LoadAndAssign.
type UfuncgraphPrograms struct {
	OnEntry       *ebpf.Program `ebpf:"on_entry"`
	OnEntryGolang *ebpf.Program `ebpf:"on_entry_golang"`
	OnExit        *ebpf.Program `ebpf:"on_exit"`
	OnExitGolang  *ebpf.Program `ebpf:"on_exit_golang"`
}

func (p *UfuncgraphPrograms) Close() error {
	return _UfuncgraphClose(
		p.OnEntry,
		p.OnEntryGolang,
		p.OnExit,
		p.OnExitGolang,
	)
}

func _UfuncgraphClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//go:embed ufuncgraph_bpfel_x86.o
var _UfuncgraphBytes []byte
