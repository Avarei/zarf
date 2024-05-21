// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package hooks provides HTTP handlers for the mutating webhook.
package hooks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/internal/agent/operations"
	"github.com/defenseunicorns/zarf/src/pkg/cluster"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/transform"
	v1 "k8s.io/api/admission/v1"

	corev1 "k8s.io/api/core/v1"
)

// NewPodMutationHook creates a new instance of pods mutation hook.
func NewPodMutationHook(ctx context.Context, cluster *cluster.Cluster) operations.Hook {
	message.Debug("hooks.NewMutationHook()")
	return operations.Hook{
		Create: func(r *v1.AdmissionRequest) (*operations.Result, error) {
			return mutatePod(ctx, r, cluster)
		},
		Update: func(r *v1.AdmissionRequest) (*operations.Result, error) {
			return mutatePod(ctx, r, cluster)
		},
	}
}

func parsePod(object []byte) (*corev1.Pod, error) {
	message.Debugf("pods.parsePod(%s)", string(object))
	var pod corev1.Pod
	if err := json.Unmarshal(object, &pod); err != nil {
		return nil, err
	}
	return &pod, nil
}

func mutatePod(ctx context.Context, r *v1.AdmissionRequest, cluster *cluster.Cluster) (*operations.Result, error) {
	message.Debugf("hooks.mutatePod()(*v1.AdmissionRequest) - %#v , %s/%s: %#v", r.Kind, r.Namespace, r.Name, r.Operation)

	pod, err := parsePod(r.Object.Raw)
	if err != nil {
		return &operations.Result{Msg: err.Error()}, nil
	}

	if pod.Labels != nil && pod.Labels["zarf-agent"] == "patched" {
		// We've already played with this pod, just keep swimming 🐟
		return &operations.Result{
			Allowed:  true,
			PatchOps: []operations.PatchOperation{},
		}, nil
	}

	state, err := cluster.LoadZarfState(ctx)
	if err != nil {
		return nil, fmt.Errorf(lang.AgentErrGetState, err)
	}
	registryURL := state.RegistryInfo.Address

	var patchOperations []operations.PatchOperation

	// Add the zarf secret to the podspec
	zarfSecret := []corev1.LocalObjectReference{{Name: config.ZarfImagePullSecretName}}
	patchOperations = append(patchOperations, operations.ReplacePatchOperation("/spec/imagePullSecrets", zarfSecret))

	// update the image host for each init container
	for idx, container := range pod.Spec.InitContainers {
		path := fmt.Sprintf("/spec/initContainers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(registryURL, container.Image)
		if err != nil {
			message.Warnf(lang.AgentErrImageSwap, container.Image)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		patchOperations = append(patchOperations, operations.ReplacePatchOperation(path, replacement))
	}

	// update the image host for each ephemeral container
	for idx, container := range pod.Spec.EphemeralContainers {
		path := fmt.Sprintf("/spec/ephemeralContainers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(registryURL, container.Image)
		if err != nil {
			message.Warnf(lang.AgentErrImageSwap, container.Image)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		patchOperations = append(patchOperations, operations.ReplacePatchOperation(path, replacement))
	}

	// update the image host for each normal container
	for idx, container := range pod.Spec.Containers {
		path := fmt.Sprintf("/spec/containers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(registryURL, container.Image)
		if err != nil {
			message.Warnf(lang.AgentErrImageSwap, container.Image)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		patchOperations = append(patchOperations, operations.ReplacePatchOperation(path, replacement))
	}

	// Add a label noting the zarf mutation
	if pod.Labels == nil {
		// If the labels path does not exist - create with map[string]string value
		patchOperations = append(patchOperations, operations.AddPatchOperation("/metadata/labels",
			map[string]string{
				"zarf-agent": "patched",
			}))
	} else {
		patchOperations = append(patchOperations, operations.ReplacePatchOperation("/metadata/labels/zarf-agent", "patched"))
	}

	return &operations.Result{
		Allowed:  true,
		PatchOps: patchOperations,
	}, nil
}
