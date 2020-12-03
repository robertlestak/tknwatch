# tknwatch

Watch a Tekton Pipeline execution and propagate exit code to calling system.

## Build

```
make
cp bin/tknwatch /bin/tknwatch
chmod +x /bin/tknwatch
```

## Usage

### Environment Variables

```
export EVENT_ID=[tekton-triggers-event-id]
# Default:
export TEKTON_API=http://tekton-dashboard.tekton-pipelines:9097
# Optional:
export TEKTON_JWT=ey...

tknwatch
```

## Purpose

Tekton jobs are triggered with a REST API, and are asynchronous.

Tools can call an API after executing a job to check the status and logs.

However for teams that are wrapping various CI/CD, workflow, etc. tools around Tekton API calls for doing various tasks, they will want to propagate the logs and exit code from Tekton back to the calling system.

This CLI watches a PipelineRun, prints all the logs to STDOUT, and then exits with the exit code of the Tekton execution.

This allows any CI/CD or workflow tool to make a JSON REST API call to Tekton, and then propagate all of the execution details back to the workflow tool.
