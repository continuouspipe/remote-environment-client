package sync

import (
	"fmt"
)

const SyncExcluded = ".cp-remote-ignore"

//fswatch -0 -r -l "$LATENCY" --exclude="/\.[^/]*$" --exclude="\.idea" --exclude="\.git" --exclude="___jb_old___" --exclude="___jb_tmp___" "$(dir)" \
//rsync --relative -rlptDv --exclude-from="$(excludes_file)" -e 'kubectl --context='"$(context)"' --namespace='"$(namespace)"' exec -i '"$POD" -- "$file" --:/app

type DirectoryEventSyncAll struct{}

func GetDirectoryEventSyncAll() *DirectoryEventSyncAll {
	return &DirectoryEventSyncAll{}
}

func (o *DirectoryEventSyncAll) OnLastChange() error {
	fmt.Println("AtAnyEvent() Called")
	return nil
}
