package printer

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/heptio/developer-dash/internal/cache"
	cachefake "github.com/heptio/developer-dash/internal/cache/fake"
	"github.com/heptio/developer-dash/internal/testutil"
	"github.com/heptio/developer-dash/internal/view/component"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_ReplicationControllerListHandler(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	printOptions := Options{
		Cache: cachefake.NewMockCache(controller),
	}

	got, err := ReplicationControllerListHandler(validReplicationControllerList, printOptions)
	require.NoError(t, err)

	containers := component.NewContainers()
	containers.Add("nginx", "nginx:1.15")

	cols := component.NewTableCols("Name", "Labels", "Status", "Age", "Containers", "Selector")
	expected := component.NewTable("ReplicationControllers", cols)
	expected.Add(component.TableRow{
		"Name":       component.NewLink("", "rc-test", "/content/overview/namespace/default/workloads/replication-controllers/rc-test"),
		"Labels":     component.NewLabels(validReplicationControllerLabels),
		"Status":     component.NewText("0/3"),
		"Age":        component.NewTimestamp(validReplicationControllerCreationTime),
		"Containers": containers,
		"Selector":   component.NewSelectors([]component.Selector{component.NewLabelSelector("app", "myapp")}),
	})

	assert.Equal(t, expected, got)
}

var (
	validReplicationControllerLabels = map[string]string{
		"foo": "bar",
	}

	validReplicationControllerCreationTime = time.Unix(1547211430, 0)

	validReplicationController = &corev1.ReplicationController{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ReplicationController",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc-test",
			Namespace: "default",
			CreationTimestamp: metav1.Time{
				Time: validReplicationControllerCreationTime,
			},
			Labels: validReplicationControllerLabels,
		},
		Status: corev1.ReplicationControllerStatus{
			Replicas:          3,
			AvailableReplicas: 0,
		},
		Spec: corev1.ReplicationControllerSpec{
			Selector: map[string]string{
				"app": "myapp",
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "nginx",
							Image: "nginx:1.15",
						},
					},
				},
			},
		},
	}

	validReplicationControllerList = &corev1.ReplicationControllerList{
		Items: []corev1.ReplicationController{
			*validReplicationController,
		},
	}
)

func TestReplicationControllerStatus(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	c := cachefake.NewMockCache(controller)

	replicationController := testutil.CreateReplicationController("rc")
	replicationController.Labels = map[string]string{
		"foo": "bar",
	}
	replicationController.Spec.Selector = map[string]string{
		"foo": "bar",
	}
	replicationController.Namespace = "testing"

	p1 := *createPodWithPhase(
		"nginx-g7f72",
		replicationController.Labels,
		corev1.PodRunning,
		metav1.NewControllerRef(replicationController, replicationController.GroupVersionKind()))

	p2 := *createPodWithPhase(
		"nginx-p64jr",
		replicationController.Labels,
		corev1.PodRunning,
		metav1.NewControllerRef(replicationController, replicationController.GroupVersionKind()))

	p3 := *createPodWithPhase(
		"nginx-x8nrk",
		replicationController.Labels,
		corev1.PodRunning,
		metav1.NewControllerRef(replicationController, replicationController.GroupVersionKind()))

	pods := &corev1.PodList{
		Items: []corev1.Pod{p1, p2, p3},
	}

	var podList []*unstructured.Unstructured
	for _, p := range pods.Items {
		u := testutil.ToUnstructured(t, &p)
		podList = append(podList, u)
	}
	key := cache.Key{
		Namespace:  "testing",
		APIVersion: "v1",
		Kind:       "Pod",
	}

	c.EXPECT().List(gomock.Eq(key)).Return(podList, nil)
	rcs := NewReplicationControllerStatus(replicationController)
	got, err := rcs.Create(c)
	require.NoError(t, err)

	expected := component.NewQuadrant()
	require.NoError(t, expected.Set(component.QuadNW, "Running", "3"))
	require.NoError(t, expected.Set(component.QuadNE, "Waiting", "0"))
	require.NoError(t, expected.Set(component.QuadSW, "Succeeded", "0"))
	require.NoError(t, expected.Set(component.QuadSE, "Failed", "0"))

	assert.Equal(t, expected, got)
}