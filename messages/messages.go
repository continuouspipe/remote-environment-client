package messages

const NoActivePodsFoundForSpecifiedServiceName = `No running pods were found for the specified service name (%s).`
const ProjectsNotFound = `No projects were found. Please ensure that you have at least one project set up in ContinuousPipe.`
const FlowsNotFound = `No flows were found. Please ensure that the project has at least one flow.`
const EnvironmentsNotFound = `No environments were found. Please ensure that the flow has at least one environment.`
const RunningPodNotFound = `No running pods were found. Please ensure that the environment has at least one running pod.`
const InteractiveModeSuggestingFlags = `\nThe pod specified will be used this time. Next time, if you want to connect directly to the same pod, you can also use the flag values '%s'.\n`
const InvalidConfigSettings = `The remote settings file is missing or the require parameters are missing (%v), please run the 'cp-remote init' command.`
const EnvironmentSpecifiedEmpty = `The environment specified is empty. Please ensure that the environment specified in the configuration file is not empty or override it with the -e flag.`
const ItWillDeleteGitBranchAndRemoteEnvironment = `This will delete the remote Git branch and remote environment. Do you want to proceed? %s`
const YesNoOptions = `(yes/no)`
const InvalidAnswerForYesNo = `Your answer needs to be either 'yes' or 'no'. Your answer was '%s'.`
const ServiceSpecifiedEmpty = "The service name specified is an empty string. Please ensure that the service specified in the configuration file is not empty or override it with the -s flag."

//List of Command Description Messages
const BuildCommandShortDescription = `Create/update the remote environment.`
const BuildCommandLongDescription = `The build command will push any local Git commits to your remote Git branch. ContinuousPipe will then build the environment. You can use the ContinuousPipe console (https://ui.continuouspipe.io/) to see when the environment has finished building and to find its URL.`
const DestroyCommandShortDescription = `Destroy the remote environment.`
const DestroyCommandLongDescription = `The destroy command will delete the remote branch used for your remote environment. ContinuousPipe will then automatically delete the remote environment.`
const ExecCommandShortDescription = "Execute a command on a container"
const ExecCommandLongDescription = `To execute a command on a container without first getting a bash session use the exec command. The command and its arguments need to follow --`
const ExecCommandExampleDescription = `
# execute -l -all on the web pod
%[1]s ex -- ls -all

# execute -l -all on the web pod overriding the project-key and remote-branch
%[1]s ex -e techup-dev-user -s web -- ls -all`

//List for suggestion messages that are displayed to the user in case of failure
const SuggestionTriggerBuildFailed = `Triggering the build has failed.
Please make sure Git has permission to push to the remote repository and is set up correctly, then retry.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionWaitForEnvironmentReadyFailed = `Something went wrong while waiting for the environment to be created.
Please check the flow in the ContinuousPipe console (https://ui.continuouspipe.io/) for highlighted issues.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionConfigurationSaveFailed = `Something went wrong when saving the configuration file.
This is usually caused by a file system permissions issue. Please check the file permissions for the local and global configuration files.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionGetEnvironmentStatusFailed = `Something went wrong when fetching the environment status from the ContinuousPipe API.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionRemoteEnvironmentDestroyFailed = `Something went wrong when destroying the remote environment using the ContinuousPipe API.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try deleting the remote environment manually in the ContinuousPipe console (https://ui.continuouspipe.io/).
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionGitHasRemoteFailed = `Something went wrong when we tried to check if the remote development branch already existed.
This issue is usually caused by Git configuration issues. Please run 'git ls-remote' and ensure that it returns a list of branches and then terminates successfully.
If the issue persists please contact support specifying the session number '%s'.

Git Error Details: %s`

const SuggestionGitDeleteHasFailed = `Something went wrong when we tried to delete the remote branch using Git.
This issue is usually caused by Git configuration and permission issues. Please try manually pushing some changes to the Git repository to ensure that you have the correct permissions.
If the issue persists please contact support specifying the session number '%s'.

Git Error Details: %s`

const SuggestionGetAPITeamsFailed = `Something went wrong when fetching the list of teams from the ContinuousPipe API.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number %s.`

const SuggestionProjectsListEmpty = `We could not find any project from the ContinuousPipe API.
Please verify that you have access to at least one project in https://ui.continuouspipe.io/
If you do have access to projects, and the issue persists, please contact support specifying the session number %s.`

const SuggestionGetFlowListFailed = `Something went wrong when fetching the list of flows from the ContinuousPipe API.
  This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
  If the issue persists please contact support specifying the session number %s.`

const SuggestionFlowListEmpty = `We could not find any flows for the project %s from the ContinuousPipe API.
Please verify that the project has at least one flow https://ui.continuouspipe.io/
If the issue persists please contact support specifying the session number %s.`

const SuggestionPodsListEmpty = `We could not find any pods for the environment %s from the ContinuousPipe API.
Please verify that the project has at least one pod running https://ui.continuouspipe.io/
If the issue persists please contact support specifying the session number %s.`

const SuggestionGetApiEnvironmentsFailed = `Something went wrong when fetching the list of environments from the ContinuousPipe API
This issue may be caused by an incorrect Flow UUID (%s), by a temporary unavailability of the ContinuousPipe API or a network issue.
Please try running "%s %s -i" which will guide you to the right pod and let you know the correct flags.
If the issue persists please contact support specifying the session number %s.`

const SuggestionEnvironmentListEmpty = `We could not find the environment %s inside the flow %s using the ContinuousPipe API.
Please verify that the project contains the environment specified at https://ui.continuouspipe.io/
If the issue persists please contact support specifying the session number %s.`

const SuggestionRunningPodNotFound = `We could not find any running pods starting with %[1]s for the environment %[2]s from the ContinuousPipe API.
This issue may be caused by an incorrect service name (pod prefix) (%[1]s), by a temporary unavailability of the ContinuousPipe API or a network issue.
Please verify that the project has at least one pod running executing 'cp-remote pods' command and try running "%[3]s %[4]s -i" which will guide you to the right pod and let you know the correct flags.
If the issue persists please contact support specifying the session number %s.`

const SuggestionGetApiEnvironmentsFailedUsingQuestioner = `Something went wrong when fetching the list of environments from the ContinuousPipe API.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number %s.`

const SuggestionEnvironmentListEmptyUsingQuestioner = `We could not find any environments for the flow %s from the ContinuousPipe API.
Please verify that the project has at least one environment https://ui.continuouspipe.io/
If the issue persists please contact support specifying the session number %s.`

//Application logic errors (unlikely to happen)
const SuggestionQuestionerMultiSelectError = `Something went wrong in the application multi selector logic.
Please contact support specifying the session number %s.`

const SuggestionGetSettingsError = `Something went wrong in the application configuration parser logic.
This issue may be cause by the application file containing wrong information. Please try re-initialising the environment.
If the issue persists please contact support specifying the session number %s.`

const SuggestionFindPodsFailed = `Something went wrong when fetching the list of pods from the kubernetes cluster.
This issue is usually caused by a temporary unavailability of the cluster or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number %s.`

const SuggestionExecRunFailed = `Something went wrong during the exec command execution.
This issue is usually caused by a temporary unavailability of the cluster, a network issue or by the fact that the pod was deleted or moved to a different node.
Check the pod status with "cp-remote pods" and reconnect once the pod is running again.
If the issue persists please contact support specifying the session number %s.`

const GetStarted = `
# Get started!
You can now run 'cp-remote watch' to automatically sync your local changes with the deployed environment. Your deployed environment can be found at this address:`
const CheckDocumentation = `\n\nPlease check the documentation at https://docs.continuouspipe.io/remote-development/.\n`
