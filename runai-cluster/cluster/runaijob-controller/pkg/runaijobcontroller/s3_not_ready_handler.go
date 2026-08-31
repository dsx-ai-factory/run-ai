/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaijobcontroller

import (
	"bytes"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

const (
	GoofysImageIndicator = "/goofys:"
	ResourcePods         = "pods"
	SubResourceExec      = "exec"
)

// workaround for the scenario that s3 file systems continue to run after main container finishes.
// to resolve this we send "pkill" to any goofys container once the main container finishes.
// based on: https://suraj.io/post/how-to-gracefully-kill-kubernetes-jobs-with-a-sidecar/
func handleS3NotReadyState(pod *corev1.Pod, cm *Controller) {
	if pod.Status.ContainerStatuses == nil || len(pod.Status.ContainerStatuses) == 0 {
		return
	}

	//
	//  Check if the pod meets the following criteria:
	//    #1 -> need to have one or more goofy container(s)
	//    #2 -> need to have one container in not ready state
	//
	goofyContainers := make([]string, 0)
	mainContainerReady := true

	for _, container := range pod.Status.ContainerStatuses {
		if strings.Contains(container.Image, GoofysImageIndicator) {
			goofyContainers = append(goofyContainers, container.Name)
			continue
		}

		if container.State.Terminated != nil && !container.Ready {
			mainContainerReady = false
		}
	}

	if len(goofyContainers) == 0 || mainContainerReady {
		return
	}

	//
	//   send exec "pkill goofys" to the goofy containers, in order to kill all the containers
	//   preventing the job to be marked as completed
	//
	for _, container := range goofyContainers {
		request := cm.kubeClient.CoreV1().RESTClient().
			Post().
			Namespace(pod.Namespace).
			Resource(ResourcePods).
			Name(pod.Name).
			SubResource(SubResourceExec).
			VersionedParams(&corev1.PodExecOptions{
				Command:   []string{"pkill", "goofys"},
				Stdin:     false,
				Stdout:    false,
				Stderr:    true,
				TTY:       false,
				Container: container,
			}, scheme.ParameterCodec)

		exec, err := remotecommand.NewSPDYExecutor(cm.config, http.MethodPost, request.URL())
		if err != nil {
			klog.Errorf("failed to kill goofy of contianer %s in pod %s:%s failed: %v", container, pod.Name, pod.Namespace, err)
		}

		//
		//   we were able to send the kill command successfully, but we need to check that it completed ok
		//
		errBuf := &bytes.Buffer{}
		err = exec.Stream(remotecommand.StreamOptions{
			Stderr: errBuf,
		})
		if err == nil {
			if errBuf.Len() != 0 {
				klog.Errorf("command 'pkill goofys' of container %s in pod %s:%s returned: %v", container, pod.Name, pod.Namespace, errBuf.String())
			}
		}
	}
}
