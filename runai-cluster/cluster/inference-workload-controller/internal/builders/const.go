package builders

const (
	// Annotations
	annotationMinScale                    = "autoscaling.knative.dev/min-scale"
	annotationMaxScale                    = "autoscaling.knative.dev/max-scale"
	annotationInitialScale                = "autoscaling.knative.dev/initial-scale"
	annotationActivationScale             = "autoscaling.knative.dev/activation-scale"
	annotationMetric                      = "autoscaling.knative.dev/metric"
	annotationTarget                      = "autoscaling.knative.dev/target"
	annotationTargetUtilizationPercentage = "autoscaling.knative.dev/target-utilization-percentage"
	annotationAutoscalingClass            = "autoscaling.knative.dev/class"
	annotationValueAutoscalingClassHPA    = "hpa.autoscaling.knative.dev"
	annotationMps                         = "mps"
	annotationScaleToZeroRetentionPeriod  = "autoscaling.knative.dev/scale-to-zero-pod-retention-period"
	annotationScaleDownDelay              = "autoscaling.knative.dev/scale-down-delay"
	annotationProgressDeadline            = "serving.knative.dev/progress-deadline"

	LwsEntityNameLeader = "leader"
	LwsEntityNameWorker = "worker"

	InferenceWorkloadInitial            = "infw"
	DistributedInferenceWorkloadInitial = "dinfw"
)
