/*
Copyright 2018 The Knative Authors

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

package pipelinerun

import (
	"fmt"
	"strings"

	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	clientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/tektoncd/pipeline/pkg/reconciler/v1alpha1/pipelinerun/resources"
	corev1 "k8s.io/api/core/v1"
)

// cancelPipelineRun makrs the PipelineRun as cancelled and any resolved taskrun too.
func cancelPipelineRun(pr *v1alpha1.PipelineRun, pipelineState []*resources.ResolvedPipelineRunTask, clientSet clientset.Interface) error {
	pr.Status.SetCondition(&duckv1alpha1.Condition{
		Type:    duckv1alpha1.ConditionSucceeded,
		Status:  corev1.ConditionFalse,
		Reason:  "PipelineRunCancelled",
		Message: fmt.Sprintf("PipelineRun %q was cancelled", pr.Name),
	})
	errs := []string{}
	for _, rprt := range pipelineState {
		if rprt.TaskRun == nil {
			// No taskrun yet, pass
			continue
		}
		rprt.TaskRun.Spec.Status = v1alpha1.TaskRunSpecStatusCancelled
		if _, err := clientSet.TektonV1alpha1().TaskRuns(pr.Namespace).UpdateStatus(rprt.TaskRun); err != nil {
			errs = append(errs, err.Error())
		}
		if _, err := clientSet.TektonV1alpha1().TaskRuns(pr.Namespace).Update(rprt.TaskRun); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("Error cancelled PipelineRun's TaskRun(s): %s", strings.Join(errs, "\n"))
	}
	return nil
}
