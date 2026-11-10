/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotkey

import (
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// The canonical form is an exclude-list: a field this package does not name
// stays in the hash. That makes the hashed set open-ended, so it cannot be
// covered by listing fields one at a time - the list would be stale the next
// time Kubernetes adds one.
//
// These two tests walk corev1.PodSpec and corev1.Container by reflection
// instead. Every field is changed and the key must move, unless the field is
// named below with the reason it is excluded. A field neither excluded nor
// mutatable fails the test, so a new Kubernetes field forces a decision rather
// than silently going uncovered.

// excludedPodSpecFields are the pod-level fields the key deliberately ignores.
// The value is why, and is only documentation.
var excludedPodSpecFields = map[string]string{
	// Scheduling: where a pod runs is not what it is, and the restore webhook
	// adds node affinity of its own.
	"NodeName":                  "scheduling",
	"NodeSelector":              "scheduling",
	"Affinity":                  "scheduling",
	"Tolerations":               "scheduling",
	"SchedulerName":             "scheduling",
	"PriorityClassName":         "scheduling",
	"Priority":                  "scheduling",
	"TopologySpreadConstraints": "scheduling",
	"SchedulingGates":           "scheduling",

	// Debug attachments: kubectl debug adds one to a running pod, which must
	// not read as drift.
	"EphemeralContainers": "debug containers",

	// A volume is hashed only when a hashed container mounts it, so changing
	// this field alone correctly does nothing.
	"Volumes": "only hashed when mounted; see TestTargetVolumeIsHashed",
}

// excludedContainerFields are the fields ignored on a hashed container.
var excludedContainerFields = map[string]string{
	"LivenessProbe":  "probes",
	"ReadinessProbe": "probes",
	"StartupProbe":   "probes",
	"Lifecycle":      "lifecycle hooks",
}

func TestEveryPodSpecFieldIsHashedOrExcluded(t *testing.T) {
	assertFieldCoverage(t, reflect.TypeOf(corev1.PodSpec{}), excludedPodSpecFields,
		func(pod *corev1.Pod) reflect.Value { return reflect.ValueOf(&pod.Spec).Elem() })
}

func TestEveryContainerFieldIsHashedOrExcluded(t *testing.T) {
	assertFieldCoverage(t, reflect.TypeOf(corev1.Container{}), excludedContainerFields,
		func(pod *corev1.Pod) reflect.Value { return reflect.ValueOf(&pod.Spec.Containers[0]).Elem() })
}

func assertFieldCoverage(t *testing.T, structType reflect.Type, excluded map[string]string,
	target func(*corev1.Pod) reflect.Value) {
	t.Helper()

	baseline := keyOrFail(t, gpuPod())

	for i := range structType.NumField() {
		field := structType.Field(i)

		t.Run(field.Name, func(t *testing.T) {
			pod := gpuPod()
			changed, err := changeField(target(pod).Field(i), field)
			if err != nil {
				// Not knowing how to change a field is a failure, not a skip:
				// otherwise a new Kubernetes field would quietly go untested.
				t.Fatalf("cannot change %s.%s: %v - add a case to changeField, "+
					"or name the field in the exclusion map", structType.Name(), field.Name, err)
			}
			if !changed {
				t.Fatalf("%s.%s was not changed", structType.Name(), field.Name)
			}

			key, err := ComputeKey(pod)
			if err != nil {
				t.Fatalf("ComputeKey after changing %s: %v", field.Name, err)
			}

			reason, isExcluded := excluded[field.Name]
			switch {
			case isExcluded && key != baseline:
				t.Errorf("%s is listed as excluded (%s) but changed the key", field.Name, reason)
			case !isExcluded && key == baseline:
				t.Errorf("%s did not change the key, and is not listed as excluded. "+
					"Either it should be hashed, or add it to the exclusion map with a reason.", field.Name)
			}
		})
	}
}

// changeField sets value to something different from its zero value. It
// reports false when it could not produce a change, and an error when it does
// not know how to handle the type at all.
func changeField(value reflect.Value, field reflect.StructField) (bool, error) {
	if special, ok := specialFieldValues[value.Type().String()]; ok {
		if !special.Type().AssignableTo(value.Type()) {
			return false, fmt.Errorf("special value for %s is %s, want %s",
				field.Name, special.Type(), value.Type())
		}
		value.Set(special)
		return true, nil
	}

	switch value.Kind() {
	case reflect.String:
		value.SetString("snapshotkey-coverage")
		return true, nil
	case reflect.Bool:
		value.SetBool(!value.Bool())
		return true, nil
	case reflect.Int32, reflect.Int64:
		value.SetInt(4242)
		return true, nil
	case reflect.Map:
		value.Set(reflect.MakeMap(value.Type()))
		key := reflect.New(value.Type().Key()).Elem()
		if key.Kind() != reflect.String {
			return false, fmt.Errorf("map key is %s, want string", key.Kind())
		}
		key.SetString("snapshotkey-coverage")
		element := reflect.New(value.Type().Elem()).Elem()
		if _, err := changeField(element, field); err != nil {
			return false, err
		}
		value.SetMapIndex(key, element)
		return true, nil
	case reflect.Slice:
		element := reflect.New(value.Type().Elem()).Elem()
		if _, err := changeFirstScalar(element); err != nil {
			return false, err
		}
		value.Set(reflect.Append(reflect.MakeSlice(value.Type(), 0, 1), element))
		return true, nil
	case reflect.Pointer:
		allocated := reflect.New(value.Type().Elem())
		if _, err := changeFirstScalar(allocated.Elem()); err != nil {
			return false, err
		}
		value.Set(allocated)
		return true, nil
	case reflect.Struct:
		return changeFirstScalar(value)
	default:
		return false, fmt.Errorf("unhandled kind %s", value.Kind())
	}
}

// changeFirstScalar sets the first settable scalar it finds, depth first. It is
// how a nested struct such as PodSecurityContext is given a distinguishable
// value without naming its fields.
func changeFirstScalar(value reflect.Value) (bool, error) {
	switch value.Kind() {
	case reflect.String:
		value.SetString("snapshotkey-coverage")
		return true, nil
	case reflect.Bool:
		value.SetBool(true)
		return true, nil
	case reflect.Int32, reflect.Int64:
		value.SetInt(4242)
		return true, nil
	case reflect.Pointer:
		allocated := reflect.New(value.Type().Elem())
		if changed, err := changeFirstScalar(allocated.Elem()); err != nil || !changed {
			return changed, err
		}
		value.Set(allocated)
		return true, nil
	case reflect.Slice:
		element := reflect.New(value.Type().Elem()).Elem()
		if changed, err := changeFirstScalar(element); err != nil || !changed {
			return changed, err
		}
		value.Set(reflect.Append(reflect.MakeSlice(value.Type(), 0, 1), element))
		return true, nil
	case reflect.Struct:
		for i := range value.NumField() {
			if !value.Field(i).CanSet() {
				continue
			}
			if changed, _ := changeFirstScalar(value.Field(i)); changed {
				return true, nil
			}
		}
		return false, fmt.Errorf("no settable scalar in %s", value.Type())
	default:
		return false, fmt.Errorf("unhandled kind %s", value.Kind())
	}
}

// specialFieldValues are fields whose generic mutation would not survive
// canonicalisation, or would make the pod unkeyable.
var specialFieldValues = map[string]reflect.Value{
	// The generic walk would reach SeccompProfile first, which is stripped as
	// one of Snapshot's own additions. These pick a field that is hashed, so
	// the test proves the rest of the security context still counts.
	"*v1.PodSecurityContext": reflect.ValueOf(&corev1.PodSecurityContext{FSGroup: ptr.To(int64(2000))}),
	"*v1.SecurityContext":    reflect.ValueOf(&corev1.SecurityContext{RunAsUser: ptr.To(int64(1000))}),

	// Resources must keep a valid GPU request or the pod stops being keyable.
	"v1.ResourceRequirements": reflect.ValueOf(corev1.ResourceRequirements{
		Limits: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1"), "cpu": resource.MustParse("2")},
	}),
	"*v1.ResourceRequirements": reflect.ValueOf(&corev1.ResourceRequirements{
		Limits: corev1.ResourceList{"cpu": resource.MustParse("2")},
	}),

	// Replacing the container list wholesale would drop the GPU container.
	// This keeps exactly one, changed.
	"[]v1.Container": reflect.ValueOf([]corev1.Container{{
		Name:      "vllm",
		Image:     "registry.example.com/llama@sha256:coverage",
		Resources: corev1.ResourceRequirements{Limits: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")}},
	}}),
}
