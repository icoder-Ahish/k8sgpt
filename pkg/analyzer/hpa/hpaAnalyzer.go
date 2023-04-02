package hpa

import (
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HPAAnalyzer struct {
	common.Analyzer
}

func (a *HPAAnalyzer) Analyze() error {
	list, err := a.Client.GetClient().AutoscalingV1().HorizontalPodAutoscalers(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, hpa := range list.Items {
		var failures []string

		// check ScaleTargetRef exist
		scaleTargetRef := hpa.Spec.ScaleTargetRef
		scaleTargetRefNotFound := false

		switch scaleTargetRef.Kind {
		case "Deployment":
			_, err := a.Client.GetClient().AppsV1().Deployments(a.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		case "ReplicationController":
			_, err := a.Client.GetClient().CoreV1().ReplicationControllers(a.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		case "ReplicaSet":
			_, err := a.Client.GetClient().AppsV1().ReplicaSets(a.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		case "StatefulSet":
			_, err := a.Client.GetClient().AppsV1().StatefulSets(a.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		default:
			failures = append(failures, fmt.Sprintf("HorizontalPodAutoscaler uses %s as ScaleTargetRef which does not possible option.", scaleTargetRef.Kind))
		}

		if scaleTargetRefNotFound {
			failures = append(failures, fmt.Sprintf("HorizontalPodAutoscaler uses %s/%s as ScaleTargetRef which does not exist.", scaleTargetRef.Kind, scaleTargetRef.Name))
		}

		if len(failures) > 0 {
			a.PreAnalysis[fmt.Sprintf("%s/%s", hpa.Namespace, hpa.Name)] = common.PreAnalysis{
				HorizontalPodAutoscalers: hpa,
				FailureDetails:           failures,
			}
		}
	}

	for key, value := range a.PreAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "HorizontalPodAutoscaler",
			Name:  key,
			Error: value.FailureDetails,
		}
		parent, _ := util.GetParent(a.Client, value.Ingress.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Result = append(a.Result, currentAnalysis)
	}
	return nil
}

func (a *HPAAnalyzer) GetResult() []common.Result {
	return a.Result
}
