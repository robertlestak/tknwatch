package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	tektonAPI                     = os.Getenv("TEKTON_API")
	tektonNamespace               = os.Getenv("TEKTON_NAMESPACE")
	tektonAPIsStr                 = os.Getenv("TEKTON_APIS")
	tektonAPIs                    = strings.Split(tektonAPIsStr, ",")
	eventID                       = os.Getenv("EVENT_ID")
	tkn                           = os.Getenv("TEKTON_JWT")
	retryCount      int           = 100
	retryInterval   time.Duration = time.Second
	containerLogs   []*ContainerLogs
	pipelineRuns    *PipelineRuns
	taskRuns        *TaskRuns
)

func cleanApisSlice() {
	var apis []string
	for _, v := range tektonAPIs {
		if strings.TrimSpace(v) != "" {
			apis = append(apis, v)
		}
	}
	tektonAPIs = apis
	if len(tektonAPIs) > 0 {
		// this will be set by getRunForTriggerIDFromAPIs
		// when we find a pipeline run
		tektonAPI = ""
	}
}

func getRunForTriggerID(api, id string) (*PipelineRuns, error) {
	tr := &PipelineRuns{}
	c := &http.Client{}
	r, err := http.NewRequest("GET", api+"/apis/tekton.dev/v1beta1/namespaces/"+tektonNamespace+"/pipelineruns/?labelSelector=triggers.tekton.dev%2Ftriggers-eventid%3D"+id, nil)
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

func getRunForTriggerIDFromAPIs(id string) (*PipelineRuns, error) {
	if len(tektonAPIs) == 0 && tektonAPI != "" {
		return getRunForTriggerID(tektonAPI, id)
	}
	for _, v := range tektonAPIs {
		tr, err := getRunForTriggerID(v, id)
		if err != nil {
			log.Errorf("error getting pipeline run from api %v: %v", v, err)
			continue
		}
		if len(tr.Items) > 0 {
			tektonAPI = v
			return tr, nil
		}
	}
	return &PipelineRuns{}, fmt.Errorf("no pipeline run found for trigger id %v", id)
}

func getTaskRunsForPipelineRun(id string) (*TaskRuns, error) {
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
		return
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
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
	cleanApisSlice()
}

func main() {
	l := log.WithFields(log.Fields{
		"app": "tknwatch",
	})
	l.Debug("starting")
	if eventID == "" && len(os.Args) <= 1 {
		l.Fatal("no eventID provided")
	} else if eventID == "" {
		eventID = os.Args[1]
		l = l.WithField("eventID", eventID)
	}
	var c bool
	var retries int
	for !c {
		pr, err := getRunForTriggerIDFromAPIs(eventID)
		if err != nil {
			l.WithError(err).Error("getRunForTriggerID error")
		}
		if len(pr.Items) < 1 {
			l.Info("no PipelineRun Items, waiting...")
			retries++
			if retries > retryCount {
				l.Fatal("retry count exceeded")
			}
			time.Sleep(retryInterval)
			continue
		}
		logs(*pr)
		c = runComplete(pr)
		if !c {
			time.Sleep(time.Second * 5)
		}
	}
}
