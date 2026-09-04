/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"bytes"
	"context"
	stderrors "errors"
	"fmt"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/thoas/go-funk"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8Yaml "k8s.io/apimachinery/pkg/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RoleBindingsResourceHandler struct {
	common.WithLoggerAndCli
	desiredRoleBindings     *desiredRoleBindingsCache
	roleBindingCm           string
	roleBindingCmNamespace  string
	validateNamespaceExists bool
	isOpenshift             bool
}

type desiredRoleBindingsCache struct {
	initialized     bool
	resourceVersion string
	roleBindings    map[string]*rbacv1.RoleBinding
}

func NewRoleBindingsResourceHandler(client client.Client,
	roleBindingCm string, roleBindingCmNamespace string, validateNamespaceExists, isOpenshift bool) RoleBindingsResourceHandler {
	logger := ctrl.Log.WithName("resource_handlers").WithName(common.LogRoleBindingTag)

	return RoleBindingsResourceHandler{
		WithLoggerAndCli: common.WithLoggerAndCli{
			Client: client,
			Log:    logger,
		},
		desiredRoleBindings:     &desiredRoleBindingsCache{},
		roleBindingCm:           roleBindingCm,
		roleBindingCmNamespace:  roleBindingCmNamespace,
		validateNamespaceExists: validateNamespaceExists,
		isOpenshift:             isOpenshift,
	}
}

func (handler RoleBindingsResourceHandler) getRoleBindingMapFromCm(ctx context.Context) (map[string]*rbacv1.RoleBinding, error) {
	if handler.roleBindingCm == "" {
		return map[string]*rbacv1.RoleBinding{}, nil
	}

	var cm v1.ConfigMap
	if err := handler.Client.Get(ctx, client.ObjectKey{Namespace: handler.roleBindingCmNamespace, Name: handler.roleBindingCm}, &cm); err != nil {
		return nil, err
	}

	if handler.desiredRoleBindings.initialized &&
		handler.desiredRoleBindings.resourceVersion == cm.ResourceVersion {
		return handler.desiredRoleBindings.roleBindings, nil
	}

	desiredRoleBindings := make(map[string]*rbacv1.RoleBinding)
	var parseErrors []error
	for key, roleBindingYaml := range cm.Data {
		roleBindingObject := &rbacv1.RoleBinding{}
		decoder := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(roleBindingYaml)), 1000)

		err := decoder.Decode(&roleBindingObject)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("parse role binding %q: %w", key, err))
			continue
		}
		if roleBindingObject.Name == "" {
			parseErrors = append(parseErrors, fmt.Errorf("parse role binding %q: metadata.name is required", key))
			continue
		}
		desiredRoleBindings[roleBindingObject.Name] = roleBindingObject
	}
	if err := stderrors.Join(parseErrors...); err != nil {
		return nil, err
	}

	handler.desiredRoleBindings.initialized = true
	handler.desiredRoleBindings.resourceVersion = cm.ResourceVersion
	handler.desiredRoleBindings.roleBindings = desiredRoleBindings

	return desiredRoleBindings, nil
}

func (handler RoleBindingsResourceHandler) HandleResource(
	project kaiv1alpha1.Project,
) ([]kaiv1alpha1.ProjectCondition, error) {
	err := handler.handleResourceInner(context.Background(), project)
	return []kaiv1alpha1.ProjectCondition{{
		Type:    kaiv1alpha1.RoleBindingsReady,
		Status:  GetStatusFromError(err),
		Reason:  getReasonFromError(err, RoleBindingsHandlerFailed),
		Message: GetMessageFromError(err),
	}}, err
}

func (handler RoleBindingsResourceHandler) handleResourceInner(
	ctx context.Context, project kaiv1alpha1.Project,
) (err error) {
	namespaceName, err := handler.KaiProjectToNamespace(&project)
	if err != nil {
		return err
	}

	if handler.validateNamespaceExists {
		var namespace v1.Namespace
		err := handler.Client.Get(context.TODO(), client.ObjectKey{Name: namespaceName}, &namespace)
		if errors.IsNotFound(err) {
			handler.Log.Info(fmt.Sprintf(
				"Could not create role-bindings because the namespace %v does not exists", namespaceName))
			return nil
		}
	}

	desiredRoleBindings, err := handler.getRoleBindingMapFromCm(ctx)
	if err != nil {
		handler.Log.Error(err, "could not parse role bindings from configmap",
			common.LogNamespaceTag, handler.roleBindingCmNamespace)
		return err
	}

	for name, roleBinding := range desiredRoleBindings {
		err = handler.manageStaticRoleBinding(name, namespaceName, project, roleBinding)
		if err != nil {
			return err
		}
	}

	return handler.deleteOmittedRoleBindings(ctx, namespaceName, project, desiredRoleBindings)
}

func (handler RoleBindingsResourceHandler) deleteOmittedRoleBindings(
	ctx context.Context, namespace string, project kaiv1alpha1.Project,
	desiredRoleBindings map[string]*rbacv1.RoleBinding,
) error {
	if handler.roleBindingCm == "" {
		return nil
	}

	roleBindings := &rbacv1.RoleBindingList{}
	if err := handler.Client.List(ctx, roleBindings, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("list RoleBindings in namespace %s: %w", namespace, err)
	}

	for i := range roleBindings.Items {
		roleBinding := &roleBindings.Items[i]
		if !metav1.IsControlledBy(roleBinding, &project) || !isManagedByProjectController(roleBinding) {
			continue
		}
		if _, desired := desiredRoleBindings[roleBinding.Name]; desired {
			continue
		}

		handler.Log.Info("Deleting RoleBinding that is no longer in the desired set",
			common.LogRoleBindingTag, roleBinding.Name,
			common.LogNamespaceTag, namespace, common.LogProjectTag, project.Name)
		if err := handler.Client.Delete(ctx, roleBinding); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("delete omitted RoleBinding %s/%s: %w", namespace, roleBinding.Name, err)
		}
	}
	return nil
}

func isManagedByProjectController(roleBinding *rbacv1.RoleBinding) bool {
	return roleBinding.Labels[common.ManagedByLabel] == common.ProjectControllerName
}

func (handler RoleBindingsResourceHandler) manageStaticRoleBinding(
	name, namespace string, project kaiv1alpha1.Project, desiredRoleBinding *rbacv1.RoleBinding,
) (err error) {
	// Prepare subjects with namespace filled in
	var subjects []rbacv1.Subject
	for _, subject := range desiredRoleBinding.Subjects {
		if subject.Namespace == "" {
			subject.Namespace = namespace
		}
		subjects = append(subjects, subject)
	}

	var currentRoleBinding rbacv1.RoleBinding
	currentRoleBinding, err = handler.getOrCreateRoleBinding(
		name, namespace, project.Name, desiredRoleBinding.RoleRef.Name, subjects, project.UID)
	if err == nil && shouldUpdateRoleBinding(currentRoleBinding, *desiredRoleBinding) {
		shouldRecreateRoleBinding := currentRoleBinding.RoleRef.Name != desiredRoleBinding.RoleRef.Name
		currentRoleBinding.RoleRef = desiredRoleBinding.RoleRef
		currentRoleBinding.Subjects = subjects
		if currentRoleBinding.Labels == nil {
			currentRoleBinding.Labels = map[string]string{}
		}
		currentRoleBinding.Labels[common.ManagedByLabel] = common.ProjectControllerName

		if shouldRecreateRoleBinding {
			handler.Log.Info(fmt.Sprintf("Recreating %s, %s RoleBinding", name, namespace), common.LogRoleBindingTag, name, common.LogNamespaceTag, namespace, common.LogProjectTag, project.Name)
			return handler.recreateRoleBinding(&currentRoleBinding, desiredRoleBinding, project)
		}

		handler.Log.Info(fmt.Sprintf("Updating %s, %s RoleBinding", name, namespace), common.LogRoleBindingTag, name, common.LogNamespaceTag, namespace, common.LogProjectTag, project.Name)
		if err = handler.Client.Update(context.Background(), &currentRoleBinding); err != nil {
			handler.Log.Error(err, "Error updating RoleBinding", common.LogRoleBindingTag, name, common.LogNamespaceTag, namespace, common.LogProjectTag, project.Name)
		}
	}
	return err
}

func (handler RoleBindingsResourceHandler) getOrCreateRoleBinding(roleBindingName, namespace, projectName,
	clusterRoleName string, subjects []rbacv1.Subject, projectUid types.UID) (roleBinding rbacv1.RoleBinding, err error) {
	if getErr := handler.Client.Get(context.Background(),
		client.ObjectKey{Namespace: namespace, Name: roleBindingName}, &roleBinding); getErr != nil {
		if errors.IsNotFound(getErr) {
			handler.Log.Info("RoleBinding doesn't exist, creating it",
				common.LogRoleBindingTag, roleBindingName, common.LogNamespaceTag, namespace)
			roleBinding = handler.buildRoleBinding(roleBindingName, namespace, projectName, clusterRoleName, subjects, projectUid)
			if err = handler.Client.Create(context.Background(), &roleBinding); err != nil && !errors.IsAlreadyExists(err) {
				handler.Log.Error(err, "Error creating RoleBinding",
					common.LogRoleBindingTag, roleBindingName, common.LogNamespaceTag, namespace)
			}
		} else {
			handler.Log.Error(getErr, "Error retrieving RoleBinding",
				common.LogRoleBindingTag, roleBindingName, common.LogNamespaceTag, namespace)
			err = getErr
		}
	}
	return roleBinding, err
}

func shouldUpdateRoleBinding(current, desired rbacv1.RoleBinding) bool {
	if !isManagedByProjectController(&current) {
		return true
	}

	if current.RoleRef.Name != desired.RoleRef.Name || current.RoleRef.Kind != desired.RoleRef.Kind ||
		current.RoleRef.APIGroup != desired.RoleRef.APIGroup {
		return true
	}

	if len(current.Subjects) != len(desired.Subjects) {
		return true
	}

	subjects := make(map[string]rbacv1.Subject)
	for _, subject := range current.Subjects {
		subjects[subject.Name] = subject
	}

	for _, desiredSubject := range desired.Subjects {
		currentSubject, exists := subjects[desiredSubject.Name]
		if !exists {
			return true
		}
		if currentSubject.APIGroup != desiredSubject.APIGroup || currentSubject.Kind != desiredSubject.Kind {
			return true
		}
	}

	return false
}

func (handler RoleBindingsResourceHandler) getDefaultSubjectsForRoleBinding() []rbacv1.Subject {
	if !handler.isOpenshift {
		return []rbacv1.Subject{}
	}
	allowedCharsForName := []rune("abcdefghijklmnopqrstuvwxyz")
	randomSuffix := funk.RandomString(20, allowedCharsForName)
	return []rbacv1.Subject{
		{
			Kind:      common.ServiceAccountKind,
			Name:      fmt.Sprintf("non-existing-name-%s", randomSuffix),
			Namespace: fmt.Sprintf("non-existing-namespace-%s", randomSuffix),
		},
	}
}

func (handler RoleBindingsResourceHandler) buildRoleBinding(roleBindingName, namespace, projectName,
	clusterRoleName string, subjects []rbacv1.Subject, projectUid types.UID) rbacv1.RoleBinding {
	// If no subjects provided, use non-existing subjects (for kyverno compatibility)
	if len(subjects) == 0 {
		subjects = handler.getDefaultSubjectsForRoleBinding()
	}

	return rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: namespace,
			Labels:    map[string]string{common.ManagedByLabel: common.ProjectControllerName},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         kaiv1alpha1.GroupVersion.Identifier(),
				Kind:               common.ProjectKind,
				Name:               projectName,
				UID:                projectUid,
				Controller:         &common.TrueRef,
				BlockOwnerDeletion: &common.TrueRef,
			}},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: common.RbacGroup,
			Kind:     common.ClusterRoleKind,
			Name:     clusterRoleName,
		},
		Subjects: subjects,
	}
}

func (handler RoleBindingsResourceHandler) recreateRoleBinding(currentRoleBinding *rbacv1.RoleBinding, desiredRoleBinding *rbacv1.RoleBinding, project kaiv1alpha1.Project) (err error) {
	if err = handler.Client.Delete(context.Background(), currentRoleBinding); err != nil {
		handler.Log.Error(err, "Error deleting RoleBinding while attempting to recreate it",
			common.LogRoleBindingTag, currentRoleBinding.Name,
			common.LogNamespaceTag, currentRoleBinding.Namespace,
			common.LogProjectTag, project.Name)

		return err
	}

	desiredRoleBindingToCreate := desiredRoleBinding.DeepCopy()
	desiredRoleBindingToCreate.Namespace = currentRoleBinding.Namespace
	if desiredRoleBindingToCreate.Labels == nil {
		desiredRoleBindingToCreate.Labels = map[string]string{}
	}
	desiredRoleBindingToCreate.Labels[common.ManagedByLabel] = common.ProjectControllerName
	desiredRoleBindingToCreate.OwnerReferences = []metav1.OwnerReference{{
		APIVersion:         kaiv1alpha1.GroupVersion.Identifier(),
		Kind:               common.ProjectKind,
		Name:               project.Name,
		UID:                project.UID,
		Controller:         &common.TrueRef,
		BlockOwnerDeletion: &common.TrueRef,
	}}
	if err = handler.Client.Create(context.Background(), desiredRoleBindingToCreate); err != nil && !errors.IsAlreadyExists(err) {
		handler.Log.Error(err, "Error creating RoleBinding while attempting to recreate it",
			common.LogRoleBindingTag, desiredRoleBinding.Name,
			common.LogNamespaceTag, desiredRoleBinding.Namespace,
			common.LogProjectTag, project.Name)

		return err
	}

	return nil
}
