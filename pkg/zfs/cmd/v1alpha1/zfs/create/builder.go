/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vcreate

import (
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	"github.com/openebs/maya/pkg/zfs/cmd/v1alpha1/bin"
	"github.com/pkg/errors"
)

const (
	// Operation defines type of zfs operation
	Operation = "create"
)

//VolumeCreate defines structure for volume 'Create' operation
type VolumeCreate struct {
	//name of dataset
	Name string

	//size of dataset
	Size string

	//blocksize of dataset
	BlockSize string

	//property for dataset
	Property []string

	//enable reservation for dataset
	Reservation bool

	//createall the non-existing parent datasets
	CreateParent bool

	//command string
	Command string

	// checks is list of predicate function used for validating object
	checks []PredicateFunc

	// error
	err error
}

// NewVolumeCreate returns new instance of object VolumeCreate
func NewVolumeCreate() *VolumeCreate {
	return &VolumeCreate{}
}

// WithCheck add given check to checks list
func (v *VolumeCreate) WithCheck(check ...PredicateFunc) *VolumeCreate {
	v.checks = append(v.checks, check...)
	return v
}

// WithName method fills the Name field of VolumeCreate object.
func (v *VolumeCreate) WithName(Name string) *VolumeCreate {
	v.Name = Name
	return v
}

// WithSize method fills the Size field of VolumeCreate object.
func (v *VolumeCreate) WithSize(Size string) *VolumeCreate {
	v.Size = Size
	return v
}

// WithBlockSize method fills the BlockSize field of VolumeCreate object.
func (v *VolumeCreate) WithBlockSize(BlockSize string) *VolumeCreate {
	v.BlockSize = BlockSize
	return v
}

// WithProperty method fills the Property field of VolumeCreate object.
func (v *VolumeCreate) WithProperty(key, value string) *VolumeCreate {
	v.SetProperty(key, value)
	return v
}

// WithReservation method fills the Reservation field of VolumeCreate object.
func (v *VolumeCreate) WithReservation(Reservation bool) *VolumeCreate {
	v.Reservation = Reservation
	return v
}

// WithCreateParent method fills the CreateParent field of VolumeCreate object.
func (v *VolumeCreate) WithCreateParent(CreateParent bool) *VolumeCreate {
	v.CreateParent = CreateParent
	return v
}

// WithCommand method fills the Command field of VolumeCreate object.
func (v *VolumeCreate) WithCommand(Command string) *VolumeCreate {
	v.Command = Command
	return v
}

// Validate is to validate generated VolumeCreate object by builder
func (v *VolumeCreate) Validate() *VolumeCreate {
	for _, check := range v.checks {
		if !check(v) {
			v.err = errors.Wrapf(v.err, "validation failed {%v}", runtime.FuncForPC(reflect.ValueOf(check).Pointer()).Name())
		}
	}
	return v
}

// Execute is to execute generated VolumeCreate object
func (v *VolumeCreate) Execute() ([]byte, error) {
	v, err := v.Build()
	if err != nil {
		return nil, err
	}

	// execute command here
	return exec.Command(bin.BASH, "-c", v.Command).CombinedOutput()
}

// Build returns the VolumeCreate object generated by builder
func (v *VolumeCreate) Build() (*VolumeCreate, error) {
	var c strings.Builder

	v = v.Validate()
	v.appendCommand(&c, bin.ZFS)
	v.appendCommand(&c, fmt.Sprintf(" %s ", Operation))

	v.appendCommand(&c, fmt.Sprintf(" -V %s", v.Size))

	if !IsReservationSet()(v) {
		v.appendCommand(&c, " -s ")
	}

	if IsBlockSizeSet()(v) {
		v.appendCommand(&c, fmt.Sprintf(" -b %s", v.BlockSize))
	}

	if IsPropertySet()(v) {
		for _, p := range v.Property {
			v.appendCommand(&c, fmt.Sprintf(" -o %s", p))
		}
	}

	v.appendCommand(&c, fmt.Sprintf(" %s", v.Name))

	// set the command
	v.Command = c.String()
	return v, v.err
}

// appendCommand append string to given string builder
func (v *VolumeCreate) appendCommand(c *strings.Builder, cmd string) {
	_, err := c.WriteString(cmd)
	if err != nil {
		v.err = errors.Wrapf(v.err, "Failed to append cmd{%s} : %s", cmd, err.Error())
	}
}
