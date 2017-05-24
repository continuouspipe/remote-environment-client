package messages

const PleaseContactSupport = `Something went wrong in the cp-remote tool application logic.
Please contact support specifying the session number '%s'.`

const NoActivePodsFoundForSpecifiedServiceName = `No running pods were found for the specified service name '%s'.`

const ProjectsNotFound = `No projects were found. Please ensure that you have at least one project set up in ContinuousPipe.`

const FlowsNotFound = `No flows were found. Please ensure that the project has at least one flow.`

const EnvironmentsNotFound = `No environments were found. Please ensure that the flow has at least one environment.`

const RunningPodNotFound = `No running pods were found. Please ensure that the environment has at least one running pod.`

const InteractiveModeSuggestingFlags = `The pod specified will be used this time.
Next time, if you want to connect directly to the same pod, you can also use the flag values '%s'.`

const InvalidConfigSettings = `The remote settings file is missing or the require parameters are missing (%v), please run the 'cp-remote init' command.`

const EnvironmentSpecifiedEmpty = `The environment specified is empty. Please ensure that the environment specified in the configuration file is not empty or override it with the -e flag.`

const ItWillDeleteGitBranchAndRemoteEnvironment = `This will delete the remote Git branch and remote environment. Do you want to proceed? %s`

const YesNoOptions = `(yes/no)`

const InvalidAnswerForYesNo = `Your answer needs to be either 'yes' or 'no'. Your answer was '%s'.`

const ServiceSpecifiedEmpty = `The service name specified is an empty string. Please ensure that the service specified in the configuration file is not empty or override it with the -s flag.`

const RemoteProjectPathEmpty = `The remote project path is an empty string. Please ensure that the remote project path specified with the --remote-project-path flag is a valid path.`

const FetchInProgress = `Fetch in progress.`

const FetchCompleted = `Fetch completed.`

const PushInProgress = `Push in progress`

const LatencyValueTooSmall = `Please specify a latency of at least 100 milli-seconds.`

const CheckingConnectionForEnvironment = `Checking connection for environment %s.`

const PodsFoundCount = `%d pods have been found:`

const PodsNotFound = `"We could not find any pods on this environment.`

const BuildCommandShortDescription = `Create/update the remote environment.`

const BuildCommandLongDescription = `The build command will push any local Git commits to your remote Git branch. ContinuousPipe will then build the environment. You can use the ContinuousPipe console (https://ui.continuouspipe.io/) to see when the environment has finished building and to find its URL.`

const DestroyCommandShortDescription = `Destroy the remote environment.`

const DestroyCommandLongDescription = `The destroy command will delete the remote branch used for your remote environment. ContinuousPipe will then automatically delete the remote environment.`

const ExecCommandShortDescription = `Execute a command on a container.`

const ExecCommandLongDescription = `To execute a command on a container without first getting a bash session use the exec command. The remote command and its arguments need to follow a double dash (--).`

const ExecCommandExampleDescription = `
# execute -ls -all on the web pod
%[1]s exec -- ls -all

# execute -ls -all on the web pod overriding the environment id
%[1]s exec -e techup-dev-user -s web -- ls -all

# execute -ls -all on a different environment (without knowing which one yet)
%[1]s exec --interactive -- ls -all`

const FetchCommandShortDescription = `Transfers file changes from the remote environment to the local filesystem.`

const FetchCommandLongDescription = `When the remote environment is rebuilt it may contain changes that you do not have on the local filesystem. For example, for a PHP project part of building the remote environment could be installing the vendors using composer. Any new or updated vendors would be on the remote environment but not on the local filesystem which would cause issues, such as autocomplete in your IDE not working correctly. The fetch command will copy changes from the remote to the local filesystem. This will resync with the default container specified during setup but you can specify another container.`

const FetchCommandExampleDescription = `
# fetch files and folders from the remote pod
%[1]s fetch
# fetch files and folders from the remote pod specifying the environment and pod
%[1]s fetch -e techup-dev-user -s web
`

const PushCmdExampleDescription = `
# push files and folders to the remote pod
%[1]s %[2]s

# push files and folders to the remote pod specifying the environment and pod
%[1]s %[2]s -e techup-dev-user -s web
`

const SyncCommandShortDescription = `Sync local changes to the remote filesystem (alias for push).`

const SyncCommandLongDescription = `The sync command will copy changes from the local filesystem to the remote environment.
Note: this will delete any files/folders in the remote environment that are not present locally.`

const PushCommandShortDescription = `Push local changes to the remote filesystem.`

const PushCommandLongDescription = `The push command will copy changes from the local filesystem to the remote environment.
Note: this will delete any files/folders in the remote environment that are not present locally.`

const WatchCommandShortDescription = `Watch local changes and synchronize with the remote environment.`

const WatchCommandLongDescription = `The watch command will sync changes from the local filesystem to the remote environment. The default container (specified during setup) will be used but you can specify another container to sync with using the -s flag.`

const PortForwardCommandShortDescription = `Forward a port to a container`

const PortForwardCommandLongDescription = `The forward command will set up port forwarding from the local environment
to a container on the remote environment that has a port exposed. This is useful for tasks
such as connecting to a database using a local client. You need to specify the container and
the port number to forward.`

const PortFowardCommandExampleDescription = `
# Listen on the same port locally and remotely
%[1]s forward 8080

# Listen on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in the pod
%[1]s forward 5000 6000

# Listen on port 8888 locally, forwarding to 5000 in the pod
%[1]s forward 8888:5000

# Listen on a random port locally, forwarding to 5000 in the pod
%[1]s forward :5000

# Listen on a random port locally, forwarding to 5000 in the pod
%[1]s forward 0:5000

# Overriding the project-key and remote-branch
%[1]s forward -e techup-dev-user -s mysql 5000`

const DeleteCommandLongDescription = `Delete resources by filenames, stdin, resources and names, or by resources and label selector.

JSON and YAML formats are accepted. Only one type of the arguments may be specified: filenames,
resources and names, or resources and label selector.

Some resources, such as pods, support graceful deletion. These resources define a default period
before they are forcibly terminated (the grace period) but you may override that value with
the --grace-period flag, or pass --now to set a grace-period of 1. Because these resources often
represent entities in the cluster, deletion may not be acknowledged immediately. If the node
hosting a pod is down or cannot reach the API server, termination may take significantly longer
than the grace period. To force delete a resource,	you must pass a grace	period of 0 and specify
the --force flag.

IMPORTANT: Force deleting pods does not wait for confirmation that the pod's processes have been
terminated, which can leave those processes running until the node detects the deletion and
completes graceful deletion. If your processes use shared storage or talk to a remote API and
depend on the name of the pod to identify themselves, force deleting those pods may result in
multiple processes running on different machines using the same identification which may lead
to data corruption or inconsistency. Only force delete pods when you are sure the pod is
terminated, or if your application can tolerate multiple copies of the same pod running at once.
Also, if you force delete pods the scheduler may place new pods on those nodes before the node
has released those resources and causing those pods to be evicted immediately.

Note that the delete command does NOT do resource version checks, so if someone
submits an update to a resource right when you submit a delete, their update
will be lost along with the rest of the resource.`

const DeleteCommandShortDescription = `Delete resources by filenames, stdin, resources and names, or by resources and label selector`

const DeleteCommandExampleDescription = `
# Delete a pod with minimal delay
kubectl delete pod foo --now

# Force delete a pod on a dead node
kubectl delete pod foo --grace-period=0 --force

# Delete a pod with UID 1234-56-7890-234234-456456.
kubectl delete pod 1234-56-7890-234234-456456

# Delete all pods
kubectl delete pods --all

# Delete pods and services with same names "baz" and "foo"
kubectl delete pod,service baz foo

# Delete pods and services with label name=myLabel.
kubectl delete pods,services -l name=myLabel`

const LogsCommandShortDescription = `Print the logs for a pod`

const LogsCommandLongDescription = `Print the logs for a pod`

const LogsCommandExampleDescription = `
# Return snapshot logs from the pod that matches the default service
%[1]s logs

# Return snapshot logs from pod mysql with only one container
%[1]s logs mysql

# Return snapshot of previous terminated ruby container logs from pod mysql
%[1]s logs -p mysql

# Begin streaming the logs of the ruby container in pod mysql
%[1]s logs -f mysql

# Display only the most recent 20 lines of output in pod mysql
%[1]s logs --tail=20 mysql

# Show all logs from pod mysql written in the last hour
%[1]s logs --since=1h mysql`

const CheckConnectionCommandShortDescription = `Check the connection to the remote environment`

const CheckConnectionCommandLongDescription = `The checkconnection command can be used to check that the connection details
for the Kubernetes cluster are correct and lists any pods that can be found for the environment.
It can be used with the environment option to check another environment`

const InitCommandShortDescription = `Initialises the remote environment`

const SuggestionTriggerBuildFailed = `Triggering the build has failed.
Please make sure Git has permission to push to the remote repository and is set up correctly, then retry.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionWaitForEnvironmentReadyFailed = `Something went wrong while waiting for the environment to be created.
Please check the flow in the ContinuousPipe console (https://ui.continuouspipe.io/) for highlighted issues.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionConfigurationSaveFailed = `Something went wrong when saving the configuration file.
This is usually caused by a file system permissions issue. Please check the file permissions for the local and global configuration files.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionGetEnvironmentStatusFailed = `Something went wrong when fetching the environment status.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionRemoteEnvironmentRunningAndExistsFailed = `Something went wrong when checking that the environment was running.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionRemoteEnvironmentDestroyFailed = `Something went wrong when destroying the remote environment.
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

const SuggestionGitPushHasFailed = `Something went wrong when we tried to push to the remote branch using Git.
This issue is usually caused by Git configuration and permission issues. Please try manually pushing some changes to the Git repository to ensure that you have the correct permissions.
If the issue persists please contact support specifying the session number '%s'.

Git Error Details: %s`

const SuggestionGetAPITeamsFailed = `Something went wrong when fetching the list of teams.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionProjectsListEmpty = `No projects were found.
Please verify that you have access to at least one project in the ContinuousPipe console (https://ui.continuouspipe.io/).
If you do have access to projects, and the issue persists, please contact support specifying the session number '%s'.`

const SuggestionGetFlowListFailed = `Something went wrong when fetching the list of flows.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionFlowListEmpty = `No flows were found for the project '%s'.
Please verify that the project has at least one flow in the ContinuousPipe console (https://ui.continuouspipe.io/).
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionPodsListEmpty = `No pods wer found for the environment '%s'.
Please verify that the project has at least one pod running in the ContinuousPipe console (https://ui.continuouspipe.io/).
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionGetApiEnvironmentsFailed = `Something went wrong when fetching the list of environments.
This issue may be caused by an incorrect flow id '%s', by a temporary unavailability of the ContinuousPipe API, or a network issue.
Please try running '%s %s -i' which will guide you to the right pod and let you know the correct flags.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionEnvironmentListEmpty = `Environment '%s' was not found for flow '%s'.
Please verify that the project contains the specified environment in the ContinuousPipe console (https://ui.continuouspipe.io/).
Please try running '%s %s -i' which will guide you to the right pod and let you know the correct flags.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionRunningPodNotFound = `No running pods starting with '%[1]s' were found for the environment '%[2]s'.
This issue may be caused by an incorrect service name '%[1]s', by a temporary unavailability of the ContinuousPipe API, or a network issue.
Please verify that the project has at least one pod running by executing 'cp-remote pods'.
Please try running '%[3]s %[4]s -i' which will guide you to the right pod and let you know the correct flags.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionGetApiEnvironmentsFailedUsingQuestioner = `Something went wrong when fetching the list of environments.
This issue is usually caused by a temporary unavailability of the ContinuousPipe API or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionEnvironmentListEmptyUsingQuestioner = `No environments wer found for the flow '%s'.
Please verify that the project has at least one environment in the ContinuousPipe console (https://ui.continuouspipe.io/).
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionWriteDefaultExclusionFileFailed = `Something went wrong when saving the default exclusion patterns for the file '%[1]s'.
This issue is usually caused by incorrect file permissions.
Please verify that the file '%[1]s' can be written to the local filesystem and then re-try.
If the issue persists please contact support specifying the session number '%[2]s'.`

const SuggestionFailedToDetermineTheAbsPath = `Something went wrong when pushing the file '%s'.
Please try pushing a different file or pushing all files.
If the issue persists please contact support specifying the session number %s`

const SuggestionCheckForLatestVersionFailed = `Something went wrong when fetching or upgrading to the latest version of the cp-remote tool
This issue is usually caused by a network connectivity issue.
Please try an alternative upgrade method at https://docs.continuouspipe.io/remote-development/getting-started/
If you have further queries please contact support specifying the session number %s`

//Application logic errors (unlikely to happen)
const SuggestionQuestionerMultiSelectError = `Something went wrong in the application multi selector logic.
Please contact support specifying the session number '%s'.`

const SuggestionGetSettingsError = `Something went wrong in the application configuration parser logic.
This issue may be cause by the application file containing wrong information. Please try re-initialising the environment.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionFindPodsFailed = `Something went wrong when fetching the list of pods from the Kubernetes cluster.
This issue is usually caused by a temporary unavailability of the cluster or a network issue. Please try again after few minutes.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionExecRunFailed = `Something went wrong during the exec command execution.
This issue is usually caused by a temporary unavailability of the cluster, a network issue or because the pod was deleted or moved to a different node.
Check the pod status with 'cp-remote pods' and reconnect once the pod is running again.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionFetchFailed = `Something went wrong during the fetch command execution.
This issue is usually caused by a temporary unavailability of the cluster, a network issue or because the pod was deleted or moved to a different node.
Check the pod status with 'cp-remote pods' and re-try once the pod is running again.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionPushFailed = `Something went wrong during the push command execution.
This issue is usually caused by a temporary unavailability of the cluster, a network issue or because the pod was deleted or moved to a different node.
Check the pod status with 'cp-remote pods' and re-try once the pod is running again.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionDirectoryMonitorFailed = `Something went wrong during the watch command execution.
This issue is usually caused by a temporary unavailability of the cluster, a network issue or because the pod was deleted or moved to a different node.
Check the pod status with 'cp-remote pods' and reconnect once the pod is running again.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionMalformedToken = `You have provided a malformed token.
Please go to https://continuouspipe.io/ to obtain a valid token.
If the issue persists please contact support specifying the session number '%s'.`

const SuggestionGetAPIUserFailed = `Something went wrong when fetching the list of environments.
This is usually caused by an invalid cp username. You have provided the user %s.
Please ensure the cp user is valid and match the api-key provided.
If the issue persists please contact support specifying the session number '%s'.`

const GetStarted = `
# Get started!
You can now run 'cp-remote watch' to automatically sync your local changes with the deployed environment. Your deployed environment can be found at this address:`

const CheckDocumentation = `Please check the documentation at https://docs.continuouspipe.io/remote-development/.`
