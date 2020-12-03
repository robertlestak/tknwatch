package main

import "time"

type PodSteps struct {
	PodName string `json:"podName"`
	Steps   []Step `json:"steps"`
}

type Step struct {
	Container  string `json:"container"`
	ImageID    string `json:"imageID"`
	Name       string `json:"name"`
	Terminated struct {
		ContainerID string    `json:"containerID"`
		ExitCode    int       `json:"exitCode"`
		FinishedAt  time.Time `json:"finishedAt"`
		Reason      string    `json:"reason"`
		StartedAt   time.Time `json:"startedAt"`
	} `json:"terminated"`
}

type Status struct {
	CompletionTime time.Time `json:"completionTime"`
	Conditions     []struct {
		LastTransitionTime time.Time `json:"lastTransitionTime"`
		Message            string    `json:"message"`
		Reason             string    `json:"reason"`
		Status             string    `json:"status"`
		Type               string    `json:"type"`
	} `json:"conditions"`
	PodName   string    `json:"podName"`
	StartTime time.Time `json:"startTime"`
	Steps     []Step    `json:"steps"`
	TaskSpec  struct {
		Params []struct {
			Description string `json:"description"`
			Name        string `json:"name"`
			Type        string `json:"type"`
		} `json:"params"`
		Resources struct {
			Inputs []struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"inputs"`
		} `json:"resources"`
		Steps []struct {
			Env []struct {
				Name      string `json:"name"`
				ValueFrom struct {
					FieldRef struct {
						FieldPath string `json:"fieldPath"`
					} `json:"fieldRef"`
				} `json:"valueFrom,omitempty"`
				Value string `json:"value,omitempty"`
			} `json:"env"`
			EnvFrom []struct {
				SecretRef struct {
					Name string `json:"name"`
				} `json:"secretRef"`
			} `json:"envFrom"`
			Image     string `json:"image"`
			Name      string `json:"name"`
			Resources struct {
				Limits struct {
					CPU    string `json:"cpu"`
					Memory string `json:"memory"`
				} `json:"limits"`
				Requests struct {
					CPU    string `json:"cpu"`
					Memory string `json:"memory"`
				} `json:"requests"`
			} `json:"resources"`
			SecurityContext struct {
				Capabilities struct {
					Add []string `json:"add"`
				} `json:"capabilities"`
			} `json:"securityContext"`
		} `json:"steps"`
	} `json:"taskSpec"`
}

// PipelineRuns contains the response from a proxied tekton API request
type PipelineRuns struct {
	APIVersion string `json:"apiVersion"`
	Items      []struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Annotations struct {
				KubectlKubernetesIoLastAppliedConfiguration string `json:"kubectl.kubernetes.io/last-applied-configuration"`
			} `json:"annotations"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
			GenerateName      string    `json:"generateName"`
			Generation        int       `json:"generation"`
			Labels            struct {
				TektonDevPipeline                string `json:"tekton.dev/pipeline"`
				TriggersTektonDevEventlistener   string `json:"triggers.tekton.dev/eventlistener"`
				TriggersTektonDevTrigger         string `json:"triggers.tekton.dev/trigger"`
				TriggersTektonDevTriggersEventid string `json:"triggers.tekton.dev/triggers-eventid"`
			} `json:"labels"`
			Name            string `json:"name"`
			Namespace       string `json:"namespace"`
			ResourceVersion string `json:"resourceVersion"`
			SelfLink        string `json:"selfLink"`
			UID             string `json:"uid"`
		} `json:"metadata"`
		Spec struct {
			Params []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"params"`
			PipelineRef struct {
				Name string `json:"name"`
			} `json:"pipelineRef"`
			Resources []struct {
				Name         string `json:"name"`
				ResourceSpec struct {
					Params []struct {
						Name  string `json:"name"`
						Value string `json:"value"`
					} `json:"params"`
					Type string `json:"type"`
				} `json:"resourceSpec"`
			} `json:"resources"`
			ServiceAccountName string `json:"serviceAccountName"`
			Timeout            string `json:"timeout"`
		} `json:"spec"`
		Status struct {
			CompletionTime time.Time `json:"completionTime"`
			Conditions     []struct {
				LastTransitionTime time.Time `json:"lastTransitionTime"`
				Message            string    `json:"message"`
				Reason             string    `json:"reason"`
				Status             string    `json:"status"`
				Type               string    `json:"type"`
			} `json:"conditions"`
			StartTime time.Time `json:"startTime"`
		} `json:"status"`
	} `json:"items"`
	Kind     string `json:"kind"`
	Metadata struct {
		Continue        string `json:"continue"`
		ResourceVersion string `json:"resourceVersion"`
		SelfLink        string `json:"selfLink"`
	} `json:"metadata"`
}

// TaskRuns contains the task runs in a PipelineRun
type TaskRuns struct {
	APIVersion string `json:"apiVersion"`
	Items      []struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Annotations struct {
				JanitorTTL                                  string `json:"janitor/ttl"`
				KubectlKubernetesIoLastAppliedConfiguration string `json:"kubectl.kubernetes.io/last-applied-configuration"`
				PipelineTektonDevRelease                    string `json:"pipeline.tekton.dev/release"`
			} `json:"annotations"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Generation        int       `json:"generation"`
			Labels            struct {
				AppKubernetesIoManagedBy         string `json:"app.kubernetes.io/managed-by"`
				TektonDevPipeline                string `json:"tekton.dev/pipeline"`
				TektonDevPipelineRun             string `json:"tekton.dev/pipelineRun"`
				TektonDevPipelineTask            string `json:"tekton.dev/pipelineTask"`
				TektonDevTask                    string `json:"tekton.dev/task"`
				TriggersTektonDevEventlistener   string `json:"triggers.tekton.dev/eventlistener"`
				TriggersTektonDevTrigger         string `json:"triggers.tekton.dev/trigger"`
				TriggersTektonDevTriggersEventid string `json:"triggers.tekton.dev/triggers-eventid"`
			} `json:"labels"`
			Name            string `json:"name"`
			Namespace       string `json:"namespace"`
			OwnerReferences []struct {
				APIVersion         string `json:"apiVersion"`
				BlockOwnerDeletion bool   `json:"blockOwnerDeletion"`
				Controller         bool   `json:"controller"`
				Kind               string `json:"kind"`
				Name               string `json:"name"`
				UID                string `json:"uid"`
			} `json:"ownerReferences"`
			ResourceVersion string `json:"resourceVersion"`
			SelfLink        string `json:"selfLink"`
			UID             string `json:"uid"`
		} `json:"metadata"`
		Spec struct {
			Params []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"params"`
			Resources struct {
				Inputs []struct {
					Name         string `json:"name"`
					ResourceSpec struct {
						Params []struct {
							Name  string `json:"name"`
							Value string `json:"value"`
						} `json:"params"`
						Type string `json:"type"`
					} `json:"resourceSpec"`
				} `json:"inputs"`
			} `json:"resources"`
			ServiceAccountName string `json:"serviceAccountName"`
			TaskRef            struct {
				Kind string `json:"kind"`
				Name string `json:"name"`
			} `json:"taskRef"`
			Timeout string `json:"timeout"`
		} `json:"spec"`
		Status Status `json:"status"`
	} `json:"items"`
	Kind     string `json:"kind"`
	Metadata struct {
		Continue        string `json:"continue"`
		ResourceVersion string `json:"resourceVersion"`
		SelfLink        string `json:"selfLink"`
	} `json:"metadata"`
}

type ContainerLogs struct {
	PodName       string
	ContainerName string
	Logs          string
	Tail          string
}
