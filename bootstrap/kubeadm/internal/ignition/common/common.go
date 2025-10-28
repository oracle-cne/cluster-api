/*
Copyright 2024 The Kubernetes Authors.

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

package common

import (
	"strings"
	"text/template"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/internal/cloudinit"
)

// Data encapsulates common data for many ignition templates
type Data struct {
	*cloudinit.BaseUserData

	KubeadmConfig            string
	UsersWithPasswordAuth    string
	FilesystemDevicesByLabel map[string]string
	ContainerLinuxConfig     *bootstrapv1.ContainerLinuxConfig
}

// Owner represents a user+group pair
type Owner struct {
	// User is the username
	User *string
	// Group is the name of the group
	Group *string
}

// DefaultTemplateFunctions provides a set of template functions that
// are useful for many ignition templates.
func DefaultTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"Indent":         templateYAMLIndent,
		"Split":          strings.Split,
		"Join":           strings.Join,
		"MountpointName": mountpointName,
		"ParseOwner":     parseOwner,
	}
}


func templateYAMLIndent(i int, input string) string {
	split := strings.Split(input, "\n")
	ident := "\n" + strings.Repeat(" ", i)
	return strings.Join(split, ident)
}

func mountpointName(name string) string {
	return strings.TrimPrefix(strings.ReplaceAll(name, "/", "-"), "-")
}

func parseOwner(ownerRaw string) Owner {
	if ownerRaw == "" {
		return Owner{}
	}

	ownerSlice := strings.Split(ownerRaw, ":")

	parseEntity := func(entity string) *string {
		if entity == "" {
			return nil
		}

		entityTrimmed := strings.TrimSpace(entity)

		return &entityTrimmed
	}

	if len(ownerSlice) == 1 {
		return Owner{
			User: parseEntity(ownerSlice[0]),
		}
	}

	return Owner{
		User:  parseEntity(ownerSlice[0]),
		Group: parseEntity(ownerSlice[1]),
	}
}
