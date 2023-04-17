//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KustomizationAutoDeployer) DeepCopyInto(out *KustomizationAutoDeployer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KustomizationAutoDeployer.
func (in *KustomizationAutoDeployer) DeepCopy() *KustomizationAutoDeployer {
	if in == nil {
		return nil
	}
	out := new(KustomizationAutoDeployer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KustomizationAutoDeployer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KustomizationAutoDeployerList) DeepCopyInto(out *KustomizationAutoDeployerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KustomizationAutoDeployer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KustomizationAutoDeployerList.
func (in *KustomizationAutoDeployerList) DeepCopy() *KustomizationAutoDeployerList {
	if in == nil {
		return nil
	}
	out := new(KustomizationAutoDeployerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KustomizationAutoDeployerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KustomizationAutoDeployerSpec) DeepCopyInto(out *KustomizationAutoDeployerSpec) {
	*out = *in
	out.KustomizationRef = in.KustomizationRef
	out.Interval = in.Interval
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KustomizationAutoDeployerSpec.
func (in *KustomizationAutoDeployerSpec) DeepCopy() *KustomizationAutoDeployerSpec {
	if in == nil {
		return nil
	}
	out := new(KustomizationAutoDeployerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KustomizationAutoDeployerStatus) DeepCopyInto(out *KustomizationAutoDeployerStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KustomizationAutoDeployerStatus.
func (in *KustomizationAutoDeployerStatus) DeepCopy() *KustomizationAutoDeployerStatus {
	if in == nil {
		return nil
	}
	out := new(KustomizationAutoDeployerStatus)
	in.DeepCopyInto(out)
	return out
}
