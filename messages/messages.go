package messages

const NoActivePodsFoundForSpecifiedServiceName = "No running pods were found for the specified service name (%s)"
const ProjectsNotFound = "We could not find any projects, ensure that you have at least one project on CP."
const FlowsNotFound = "We could not find any flows, ensure that the project has at least one flow."
const EnvironmentsNotFound = "We could not find any environment, ensure that the flow has at least one environment."
const RunningPodNotFound = "We could not find any running pod, ensure that the environment has at least one running pod."
const InteractiveModeSuggestingFlags = "\nThank you, we will use the pod specified.\nNext time, if you want to connect directly the same pod, you can also use this flag values '%s'\n"
const PodKilledOrMoved = "The pod may have been killed or moved to a different node."
const PodKilledOrMovedSuggestingAction = "Check the pod status with `cp-remote pods` and re-connect once the pod is running again."
const InvalidConfigSettings = "The remote settings file is missing or the require parameters are missing (%v), please run the init command."

//List for suggestion messages that are displayed to the user in case of failure
const SuggestionTriggerBuildFailed = `
Triggering the build has failed.
Please make sure git has permission to push to the remote repository and is setup correctly, then retry.
If the issue still persist please contact support specifying the session number %s.`

const SuggestionWaitForEnvironmentReadyFailed = `
Something went wrong happened while waiting for the environment to be created.
Please check on https://ui.continuouspipe.io/ the flow and resolve any highlighted issue.
If the issue still persist please contact support specifying the session number %s.`

const SuggestionConfigurationSaveFailed = `
Something went wrong when saving the configuration file.
This is usually caused by an environmental issue, check the file permissions for the local and global configuration files.
If the issue still persist please contact support specifying the session number %s.`

const SuggestionGetEnvironmentStatusFailed = `
Something went wrong when fetching the environment status from the Continuous Pipe API.
This issue is usually caused by a temporary un-availability of the CP API or a network issue.
Try again after few minutes and if the issue still persist please contact support specifying the session number %s.`
