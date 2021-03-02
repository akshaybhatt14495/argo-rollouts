package info

import (
	"fmt"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/akshaybhatt14495/argo-rollouts/pkg/apis/rollouts/v1alpha1"
)

type ExperimentInfo struct {
	Metadata
	Icon         string
	Revision     int
	Status       string
	Message      string
	ReplicaSets  []ReplicaSetInfo
	AnalysisRuns []AnalysisRunInfo
}

func NewExperimentInfo(
	exp *v1alpha1.Experiment,
	allReplicaSets []*appsv1.ReplicaSet,
	allAnalysisRuns []*v1alpha1.AnalysisRun,
	allPods []*corev1.Pod,
) *ExperimentInfo {

	expInfo := ExperimentInfo{
		Metadata: Metadata{
			Name:              exp.Name,
			Namespace:         exp.Namespace,
			CreationTimestamp: exp.CreationTimestamp,
			UID:               exp.UID,
		},
		Status:  string(exp.Status.Phase),
		Message: exp.Status.Message,
	}
	expInfo.Icon = analysisIcon(exp.Status.Phase)
	expInfo.Revision = parseRevision(exp.ObjectMeta.Annotations)
	expInfo.ReplicaSets = getReplicaSetInfo(exp.UID, nil, allReplicaSets, allPods)
	expInfo.AnalysisRuns = getAnalysisRunInfo(exp.UID, allAnalysisRuns)
	return &expInfo
}

func getExperimentInfo(
	ro *v1alpha1.Rollout,
	allExperiments []*v1alpha1.Experiment,
	allReplicaSets []*appsv1.ReplicaSet,
	allAnalysisRuns []*v1alpha1.AnalysisRun,
	allPods []*corev1.Pod,
) []ExperimentInfo {

	var expInfos []ExperimentInfo
	for _, exp := range allExperiments {
		if ownerRef(exp.OwnerReferences, []types.UID{ro.UID}) == nil {
			continue
		}
		expInfo := NewExperimentInfo(exp, allReplicaSets, allAnalysisRuns, allPods)
		expInfos = append(expInfos, *expInfo)
	}
	sort.Slice(expInfos[:], func(i, j int) bool {
		if expInfos[i].Revision > expInfos[j].Revision {
			return true
		}
		return expInfos[i].CreationTimestamp.Before(&expInfos[j].CreationTimestamp)
	})
	return expInfos
}

// Images returns a list of images that are currently running along with tags on which stack they belong to
func (r *ExperimentInfo) Images() []ImageInfo {
	var images []ImageInfo
	for _, rsInfo := range r.ReplicaSets {
		if rsInfo.Replicas > 0 {
			for _, image := range rsInfo.Images {
				newImage := ImageInfo{
					Image: image,
				}
				if rsInfo.Template != "" {
					newImage.Tags = append(newImage.Tags, fmt.Sprintf("Σ:%s", rsInfo.Template))
				}
				images = mergeImageAndTags(newImage, images)
			}
		}
	}
	return images
}

func analysisIcon(status v1alpha1.AnalysisPhase) string {
	switch status {
	case v1alpha1.AnalysisPhaseSuccessful:
		return IconOK
	case v1alpha1.AnalysisPhaseInconclusive:
		return IconUnknown
	case v1alpha1.AnalysisPhaseFailed:
		return IconBad
	case v1alpha1.AnalysisPhaseError:
		return IconWarning
	case v1alpha1.AnalysisPhaseRunning:
		return IconProgressing
	case v1alpha1.AnalysisPhasePending:
		return IconWaiting
	}
	return " "
}
