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

type GofuncgraphEvent struct {
	Goid     uint64
	CallerIp uint64
	Ip       uint64
	TimeNs   uint64
	Location uint8
	Errno    uint8
	Data     [100]uint8
	_        [2]byte
}

// LoadGofuncgraph returns the embedded CollectionSpec for Gofuncgraph.
func LoadGofuncgraph() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_GofuncgraphBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load Gofuncgraph: %w", err)
	}

	return spec, err
}

// LoadGofuncgraphObjects loads Gofuncgraph and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*GofuncgraphObjects
//	*GofuncgraphPrograms
//	*GofuncgraphMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func LoadGofuncgraphObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := LoadGofuncgraph()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// GofuncgraphSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type GofuncgraphSpecs struct {
	GofuncgraphProgramSpecs
	GofuncgraphMapSpecs
}

// GofuncgraphSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type GofuncgraphProgramSpecs struct {
	Ent *ebpf.ProgramSpec `ebpf:"ent"`
	Ret *ebpf.ProgramSpec `ebpf:"ret"`
}

// GofuncgraphMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type GofuncgraphMapSpecs struct {
	BpfStack   *ebpf.MapSpec `ebpf:"bpf_stack"`
	EventQueue *ebpf.MapSpec `ebpf:"event_queue"`
	Heap       *ebpf.MapSpec `ebpf:"heap"`
}

// GofuncgraphObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to LoadGofuncgraphObjects or ebpf.CollectionSpec.LoadAndAssign.
type GofuncgraphObjects struct {
	GofuncgraphPrograms
	GofuncgraphMaps
}

func (o *GofuncgraphObjects) Close() error {
	return _GofuncgraphClose(
		&o.GofuncgraphPrograms,
		&o.GofuncgraphMaps,
	)
}

// GofuncgraphMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to LoadGofuncgraphObjects or ebpf.CollectionSpec.LoadAndAssign.
type GofuncgraphMaps struct {
	BpfStack   *ebpf.Map `ebpf:"bpf_stack"`
	EventQueue *ebpf.Map `ebpf:"event_queue"`
	Heap       *ebpf.Map `ebpf:"heap"`
}

func (m *GofuncgraphMaps) Close() error {
	return _GofuncgraphClose(
		m.BpfStack,
		m.EventQueue,
		m.Heap,
	)
}

// GofuncgraphPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to LoadGofuncgraphObjects or ebpf.CollectionSpec.LoadAndAssign.
type GofuncgraphPrograms struct {
	Ent *ebpf.Program `ebpf:"ent"`
	Ret *ebpf.Program `ebpf:"ret"`
}

func (p *GofuncgraphPrograms) Close() error {
	return _GofuncgraphClose(
		p.Ent,
		p.Ret,
	)
}

func _GofuncgraphClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed gofuncgraph_bpfel_x86.o
var _GofuncgraphBytes []byte
