package common

import (
	"context"
	"fmt"

	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logger.Log.WithName("creator.common")

func (self *OldSharedBuilder) Create(ctx context.Context) error {

	createOptions, updateOptions, logPrefix := self.HandleDryRun()

	var err error

	/////////////////////////////////////////////////////////////////////////////////
	//  PVCs
	/////////////////////////////////////////////////////////////////////////////////
	for _, pvc := range self.SharedResources.Pvcs {

		err = controller_utils.CreateOrUpdateWithOptionsIfNeeded(ctx, self.K8sClient, log, &pvc, createOptions, updateOptions)
		if err != nil {
			self.ReportEvent(common.NewEvent(common.SeverityError, fmt.Sprintf("failed to create PVC: %s", err.Error())))
			return err
		}
		self.ReportCreatedResource(&pvc)
		log.V(5).Info(logPrefix, "pvc", pvc)
	}

	/////////////////////////////////////////////////////////////////////////////////
	// Secrets
	/////////////////////////////////////////////////////////////////////////////////
	for _, secret := range self.SharedResources.Secrets {
		err = controller_utils.CreateOrUpdateWithOptionsIfNeeded(ctx, self.K8sClient, log, &secret, createOptions, updateOptions)
		if err != nil {
			self.ReportEvent(common.NewEvent(common.SeverityError, fmt.Sprintf("failed to create Secret: %s", err.Error())))
			return err
		}
		self.ReportCreatedResource(&secret)
		log.V(5).Info(logPrefix, "secret", secret)

	}

	/////////////////////////////////////////////////////////////////////////////////
	//  Services
	/////////////////////////////////////////////////////////////////////////////////
	for _, svc := range self.SharedResources.Services {
		//
		//   note! we deliberately send a pointer to Create and not use "svc" because we
		//   need the update spec of the created service (we load auto generated port information
		//   from it)
		//
		if self.Options.Owner != nil {
			err = controller_utils.SetControllerReference(self.Options.Owner, &svc, self.K8sClient.Scheme())
			if err != nil {
				self.ReportEvent(common.NewEvent(common.SeverityError, fmt.Sprintf("failed to set controller reference: %s", err.Error())))
				return err
			}
		}
		log.Info("creating service", "name", svc.Name, "namespace", svc.Namespace)
		err = controller_utils.CreateOrUpdateWithOptionsIfNeeded(ctx, self.K8sClient, log, &svc, createOptions, updateOptions)
		if err != nil {
			self.ReportEvent(common.NewEvent(common.SeverityError, fmt.Sprintf("failed to create service: %s: %s", workloadproject.NamespacedName(svc.ObjectMeta), err.Error())))
			return err
		}

		self.ReportCreatedResource(&svc)
		log.V(5).Info(logPrefix, "service", svc)
	}

	/////////////////////////////////////////////////////////////////////////////////
	//  Ingresses
	/////////////////////////////////////////////////////////////////////////////////
	for _, ingress := range self.SharedResources.Ingress {
		if self.Options.Owner != nil {
			err = controller_utils.SetControllerReference(self.Options.Owner, &ingress, self.K8sClient.Scheme())
			if err != nil {
				self.ReportEvent(common.NewEvent(common.SeverityError, fmt.Sprintf("failed to set controller reference: %s", err.Error())))
				return err
			}
		}

		err = controller_utils.CreateOrUpdateWithOptionsIfNeeded(ctx, self.K8sClient, log, &ingress, createOptions, updateOptions)
		if err != nil {
			self.ReportEvent(common.NewEvent(common.SeverityError, fmt.Sprintf("failed to create ingress: %s: %s", workloadproject.NamespacedName(ingress.ObjectMeta), err.Error())))
			return err
		}
		self.ReportCreatedResource(&ingress)
		log.V(5).Info(logPrefix, "ingress", ingress)
	}

	/////////////////////////////////////////////////////////////////////////////////
	//  Openshift Routes
	/////////////////////////////////////////////////////////////////////////////////
	for _, route := range self.SharedResources.Route {
		if self.Options.Owner != nil {
			err = controller_utils.SetControllerReference(self.Options.Owner, &route, self.K8sClient.Scheme())
			if err != nil {
				self.ReportEvent(common.NewEvent(common.SeverityError, fmt.Sprintf("failed to set controller reference: %s", err.Error())))
				return err
			}
		}
		err = controller_utils.CreateOrUpdateWithOptionsIfNeeded(ctx, self.K8sClient, log, &route, createOptions, updateOptions)
		if err != nil {
			self.ReportEvent(common.NewEvent(common.SeverityError, fmt.Sprintf("failed to create route: %s: %s", workloadproject.NamespacedName(route.ObjectMeta), err.Error())))
			return err
		}
		self.ReportCreatedResource(&route)
		log.Info(logPrefix, "route", route)
	}

	return nil
}
