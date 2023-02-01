package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	tektonAPI       = os.Getenv("TEKTON_API")
	tektonNamespace = os.Getenv("TEKTON_NAMESPACE")
	eventID         = os.Getenv("EVENT_ID")
	tkn             = os.Getenv("TEKTON_JWT")
	containerLogs   []*ContainerLogs
	pipelineRuns    *PipelineRuns
	taskRuns        *TaskRuns
)

func getRunForTriggerID(id string) (*PipelineRuns, error) {
	if pipelineRuns != nil {
		//return pipelineRuns, nil
	}
	tr := &PipelineRuns{}
	c := &http.Client{}
	r, err := http.NewRequest("GET", tektonAPI+"/apis/tekton.dev/v1beta1/namespaces/"+tektonNamespace+"/pipelineruns/?labelSelector=triggers.tekton.dev%2Ftriggers-eventid%3D"+id, nil)
	if err != nil {
		return tr, err
	}
	setAuthHeaders(r)
	res, rerr := c.Do(r)
	if rerr != nil {
		return tr, rerr
	}
	defer res.Body.Close()
	if res.StatusCode > 202 {
		return tr, fmt.Errorf("res.Status: %v", res.Status)
	}
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return tr, berr
	}
	jerr := json.Unmarshal(bd, &tr)
	if jerr != nil {
		return tr, jerr
	}
	pipelineRuns = tr
	return tr, nil
}

func getTaskRunsForPipelineRun(id string) (*TaskRuns, error) {
	if taskRuns != nil {
		//return taskRuns, nil
	}
	tr := &TaskRuns{}
	c := &http.Client{}
	r, err := http.NewRequest("GET", tektonAPI+"/apis/tekton.dev/v1beta1/namespaces/"+tektonNamespace+"/taskruns/?labelSelector=tekton.dev%2FpipelineRun%3D"+id, nil)
	if err != nil {
		return tr, err
	}
	setAuthHeaders(r)
	res, rerr := c.Do(r)
	if rerr != nil {
		return tr, rerr
	}
	defer res.Body.Close()
	if res.StatusCode > 202 {
		return tr, fmt.Errorf("res.Status: %v", res.Status)
	}
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return tr, berr
	}
	jerr := json.Unmarshal(bd, &tr)
	if jerr != nil {
		return tr, jerr
	}
	taskRuns = tr
	return tr, nil
}

func setAuthHeaders(r *http.Request) {
	if tkn != "" {
		r.Header.Set("x-mesh-sso", tkn)
		r.Header.Set("Authorization", "Bearer "+tkn)
	}
}

func (tr *TaskRuns) PodSteps() ([]PodSteps, error) {
	var ps []PodSteps
	var e error
	for _, v := range tr.Items {
		p := PodSteps{
			PodName: v.Status.PodName,
			Steps:   v.Status.Steps,
		}
		ps = append(ps, p)
	}
	return ps, e
}

func (cl *ContainerLogs) Append(l string) {
	cl.Tail = strings.ReplaceAll(l, cl.Logs, "")
	cl.Logs = l
}

func (ps PodSteps) Logs() error {
	var e error
	for _, c := range ps.Steps {
		ll, e := getPodLogs(ps.PodName, c.Container)
		if e != nil {
			return e
		}
		var clexist bool
		for _, k := range containerLogs {
			if k.PodName == ps.PodName && k.ContainerName == c.Container {
				k.Append(ll)
				if strings.TrimSpace(k.Tail) != "" {
					fmt.Println(k.Tail)
				}
				clexist = true
			}
		}
		if !clexist {
			cl := &ContainerLogs{
				PodName:       ps.PodName,
				ContainerName: c.Container,
			}
			cl.Append(ll)
			if strings.TrimSpace(cl.Tail) != "" {
				fmt.Println(cl.Tail)
			}
			containerLogs = append(containerLogs, cl)
		}
	}
	return e
}

func getPodLogs(pod, container string) (string, error) {
	var logs string
	r, err := http.NewRequest("GET", tektonAPI+"/api/v1/namespaces/"+tektonNamespace+"/pods/"+pod+"/log?container="+container, nil)
	if err != nil {
		return logs, err
	}
	setAuthHeaders(r)
	c := &http.Client{}
	res, rerr := c.Do(r)
	if rerr != nil {
		return logs, rerr
	}
	defer res.Body.Close()
	if res.StatusCode > 202 {
		return logs, fmt.Errorf("res.Status: %v", res.Status)
	}
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return logs, berr
	}
	return string(bd), nil
}

func exitCode(ps PodSteps) int {
	var s int
	for _, v := range ps.Steps {
		if v.Terminated.ExitCode > 0 {
			s = v.Terminated.ExitCode
			return s
		}
	}
	return s
}

func runComplete(pr *PipelineRuns) bool {
	for _, v := range pr.Items {
		if v.Status.CompletionTime.IsZero() {
			return false
		}
	}
	return true
}

func logs(pr PipelineRuns) {
	if len(pr.Items) < 1 {
		log.Println("no PipelineRun Items")
		os.Exit(1)
	}
	tr, err := getTaskRunsForPipelineRun(pr.Items[0].Metadata.Name)
	if err != nil {
		log.Printf("getTaskRunsForPipelineRun error: %v\n", err)
	}
	ps, perr := tr.PodSteps()
	if perr != nil {
		log.Printf("PodSteps error: %v\n", perr)
	}
	for _, s := range ps {
		s.Logs()
		e := exitCode(s)
		if e > 0 {
			os.Exit(e)
		}
	}
}

func init() {
	if tektonAPI == "" {
		tektonAPI = "http://tekton-dashboard.tekton-pipelines:9097"
	}
	if tektonNamespace == "" {
		tektonNamespace = "tekton-pipelines"
	}
}

func main() {
	if eventID == "" && len(os.Args) <= 1 {
		fmt.Println("event ID required")
		os.Exit(1)
	} else if eventID == "" {
		eventID = os.Args[1]
	}
	var c bool
	for !c {
		pr, err := getRunForTriggerID(eventID)
		if err != nil {
			log.Printf("getRunForTriggerID(%s) error: %v\n", eventID, err)
		}
		c = runComplete(pr)
		logs(*pr)
		if !c {
			time.Sleep(time.Second * 5)
		}
	}
}
