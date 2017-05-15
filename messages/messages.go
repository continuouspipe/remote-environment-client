package messages

const NoActivePodsFoundForSpecifiedServiceName = "No running pods were found for the specified service name (%s)."
const ProjectsNotFound = "No projects were found. Please ensure that you have at least one project set up in ContinuousPipe."
const FlowsNotFound = "No flows were found. Please ensure that the project has at least one flow."
const EnvironmentsNotFound = "No environments were found. Please ensure that the flow has at least one environment."
const RunningPodNotFound = "No running pods were found. Please ensure that the environment has at least one running pod."
const InteractiveModeSuggestingFlags = "\nThe pod specified will be used this time.\nNext time, if you want to connect directly to the same pod, you can also use the flag values '%s'.\n"
const PodKilledOrMoved = "The pod may have been killed or moved to a different node."
const PodKilledOrMovedSuggestingAction = "Check the pod status with `cp-remote pods` and reconnect once the pod is running again."
const InvalidConfigSettings = "The remote settings file is missing or the require parameters are missing (%v), please run the `cp-remote init` command."
const EnvironmentSpecifiedEmpty = "The environment specified is empty. Please ensure that the environment specified in the configuration file is not empty or override it with the -e flag."
const ItWillDeleteGitBranchAndRemoteEnvironment = "This will delete the remote git branch and remote environment, do you want to proceed? %s"
const YesNoOptions = "(yes/no)"
const InvalidAnswerForYesNo = "Your answer needs to be either yes or no. Your answer was %s"

//List of Command Description Messages
const BuildCommandShortDescription = "Create/Update the remote environment"
const BuildCommandLongDescription = `The build command will push changes the branch you have checked out locally to your remote
environment branch. ContinuousPipe will then build the environment. You can use the
https://ui.continuouspipe.io/ to see when the environment has finished building and to
find its IP address.`
const DestroyCommandShortDescription = "Destroy the remote environment"
const DestroyCommandLongDescription = `The destroy command will delete the remote branch used for your remote
environment, ContinuousPipe will then remove the environment.`

//List for suggestion messages that are displayed to the user in case of failure
const SuggestionTriggerBuildFailed = `Triggering the build has failed.
Please make sure Git has permission to push to the remote repository and is set up correctly, then retry.
If the issue persists please contact support specifying the session number %s.`

const SuggestionWaitForEnvironmentReadyFailed = `Something went wrong while waiting for the environment to be created.
Please check the flow on https://ui.continuouspipe.io/ for highlighted issues.
If the issue persists please contact support specifying the session number %s.`

const SuggestionConfigurationSaveFailed = `Something went wrong when saving the configuration file.
This is usually caused by a file system permissions issue. Please check the file permissions for the local and global configuration files.
If the issue persists please contact support specifying the session number %s.`

const SuggestionGetEnvironmentStatusFailed = `Something went wrong when fetching the environment status from the ContinuousPipe API.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number %s.`

const SuggestionRemoteEnvironmentDestroyFailed = `Something went wrong when destroying the remote environment using the ContinuousPipe API.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try deleting the remote environment manually on https://ui.continuouspipe.io/
If the issue persists please contact support specifying the session number %s.`

const SuggestionGitHasRemoteFailed = `Something went wrong when we tried to check if the remote development branch existed already.
This issue is usually caused by git configuration issues. Does "git ls-remote" terminate successfully? If not, can you configure it so that it does return the list of branches?
If the issue persists please contact support specifying the session number %s.

Git Error Details: %s`

const SuggestionGitDeleteHasFailed = `Something went wrong when we tried to delete the remote branch using git.
This issue is usually caused by git configuration and permission issues. Does the command "git push" work? If not, can you fix it?
If the issue persists please contact support specifying the session number %s.

Git Error Details: %s`

const GetStarted = `
# Get started !
You can now run 'cp-remote watch' to watch your local changes with the deployed environment ! Your deployed environment can be found at this address:`
const CheckDocumentation = "\n\nCheckout the documentation at https://docs.continuouspipe.io/remote-development/ \n"
