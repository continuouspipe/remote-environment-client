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