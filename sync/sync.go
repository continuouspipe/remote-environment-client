package sync

const SyncExcluded = ".cp-remote-ignore"

//rsync --relative -rlptDv --exclude-from="$(excludes_file)" -e 'kubectl --context='"$(context)"' --namespace='"$(namespace)"' exec -i '"$POD" -- "$file" --:/app

type DirectoryEventSyncAll struct{}

func GetDirectoryEventSyncAll() *DirectoryEventSyncAll {
	return &DirectoryEventSyncAll{}
}

func (o *DirectoryEventSyncAll) OnLastChange() error {



	return nil
}
