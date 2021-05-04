// +build !ignore_autogenerated

/*
Copyright 2021.

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MariaDBCluster) DeepCopyInto(out *MariaDBCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBCluster.
func (in *MariaDBCluster) DeepCopy() *MariaDBCluster {
	if in == nil {
		return nil
	}
	out := new(MariaDBCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MariaDBCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MariaDBClusterList) DeepCopyInto(out *MariaDBClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MariaDBCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBClusterList.
func (in *MariaDBClusterList) DeepCopy() *MariaDBClusterList {
	if in == nil {
		return nil
	}
	out := new(MariaDBClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MariaDBClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MariaDBClusterSpec) DeepCopyInto(out *MariaDBClusterSpec) {
	*out = *in
	in.RootPassword.DeepCopyInto(&out.RootPassword)
	if in.MariaDBConf != nil {
		in, out := &in.MariaDBConf, &out.MariaDBConf
		*out = make(MariaDBConf, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBClusterSpec.
func (in *MariaDBClusterSpec) DeepCopy() *MariaDBClusterSpec {
	if in == nil {
		return nil
	}
	out := new(MariaDBClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MariaDBClusterStatus) DeepCopyInto(out *MariaDBClusterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBClusterStatus.
func (in *MariaDBClusterStatus) DeepCopy() *MariaDBClusterStatus {
	if in == nil {
		return nil
	}
	out := new(MariaDBClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in MariaDBConf) DeepCopyInto(out *MariaDBConf) {
	{
		in := &in
		*out = make(MariaDBConf, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MariaDBConf.
func (in MariaDBConf) DeepCopy() MariaDBConf {
	if in == nil {
		return nil
	}
	out := new(MariaDBConf)
	in.DeepCopyInto(out)
	return *out
}
