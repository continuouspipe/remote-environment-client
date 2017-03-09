//The FSEvents API in Mac OS X allows applications to register for notifications of changes to a given directory tree.
//This files contains an implementation of the DirectoryMonitor interface that must be used for OSX
package monitor

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/fsnotify/fsevents"
	"sync"
	"time"
)

func init() {
	dirMonitor = NewFsEvents()
}

type FsEvents struct {
	Exclusions ExclusionProvider
	Latency    time.Duration //sync latency in milliseconds
}

func NewFsEvents() *FsEvents {
	return &FsEvents{}
}

func (m *FsEvents) SetExclusions(exclusion ExclusionProvider) {
	m.Exclusions = exclusion
}

func (m *FsEvents) SetLatency(latency time.Duration) {
	m.Latency = latency
}

func (m FsEvents) AnyEventCall(directory string, observer EventsObserver) error {
	dev, err := fsevents.DeviceForPath(directory)
	if err != nil {
		return fmt.Errorf("Failed to retrieve device for path: %v", err)
	}
	cplogs.V(5).Infoln(dev)
	cplogs.V(5).Infoln(fsevents.EventIDForDeviceBeforeTime(dev, time.Now()))

	es := &fsevents.EventStream{
		Paths:   []string{directory},
		Latency: 500 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.WatchRoot}
	es.Start()
	ec := es.Events

	cplogs.V(5).Infof("Moditoring directory %s", directory)
	cplogs.V(5).Infof("Device UUID %s", fsevents.GetDeviceUUID(dev))

	var (
		changeLock  sync.Mutex
		dirty       bool
		lastChange  time.Time
		pathsToSync []string
	)

	go func() {
		for msg := range ec {
			for _, e := range msg {
				desc := m.GetEventDescription(e.Flags)
				cplogs.V(1).Infof("filesystem event for %d(%s) %s\n", e.ID, e.Path, desc)
				cplogs.Flush()

				if (e.Flags & (fsevents.ItemCreated | fsevents.ItemRemoved | fsevents.ItemRenamed | fsevents.ItemModified | fsevents.ItemChangeOwner)) == 0 {
					cplogs.V(5).Infof("skipping event %s on path %s as is not an item create, remove, renamed, modified or changed owner", desc, e.Path)
					cplogs.Flush()
					continue
				}
				//check if the file matches the exclusion list, if so ignore the event
				match := m.Exclusions.MatchExclusionList(e.Path)
				if match == true {
					cplogs.V(5).Infof("skipping %s %s as is in the exclusion list", desc, e.Path)
					cplogs.Flush()
					continue
				}

				changeLock.Lock()
				pathsToSync = append(pathsToSync, e.Path)
				lastChange = time.Now()
				dirty = true
				changeLock.Unlock()
			}
		}
	}()

	//default latency 500 ms
	latency := time.Duration(500)

	//allow the user to override but only if is at least 100ms
	if m.Latency > 100 {
		latency = m.Latency
	}
	delay := latency * time.Millisecond
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		changeLock.Lock()
		// if a change happened more than 'delay' seconds ago, sync it now.
		// if a change happened less than 'delay' seconds ago, sleep for 'delay' seconds
		// and see if more changes happen, we don't want to sync when
		// the filesystem is in the middle of changing due to a massive
		// set of changes (such as a local build in progress).
		if dirty && time.Now().After(lastChange.Add(delay)) {
			fmt.Println("Synchronizing filesystem changes...")
			err = observer.OnLastChange(pathsToSync)
			if err != nil {
				return err
			}
			fmt.Println("Done.")
			cplogs.Flush()
			dirty = false
			pathsToSync = []string{}
		}
		changeLock.Unlock()
		<-ticker.C
	}

	return nil
}

func (m FsEvents) GetEventDescription(flags fsevents.EventFlags) string {
	var noteDescription = map[fsevents.EventFlags]string{
		fsevents.MustScanSubDirs: "MustScanSubdirs",
		fsevents.UserDropped:     "UserDropped",
		fsevents.KernelDropped:   "KernelDropped",
		fsevents.EventIDsWrapped: "EventIDsWrapped",
		fsevents.HistoryDone:     "HistoryDone",
		fsevents.RootChanged:     "RootChanged",
		fsevents.Mount:           "Mount",
		fsevents.Unmount:         "Unmount",

		fsevents.ItemCreated:  "Created",
		fsevents.ItemRemoved:  "Removed",
		fsevents.ItemRenamed:  "Renamed",
		fsevents.ItemModified: "Modified",

		fsevents.ItemInodeMetaMod:  "InodeMetaMod",
		fsevents.ItemFinderInfoMod: "FinderInfoMod",
		fsevents.ItemChangeOwner:   "ChangeOwner",
		fsevents.ItemXattrMod:      "XAttrMod",
		fsevents.ItemIsFile:        "IsFile",
		fsevents.ItemIsDir:         "IsDir",
		fsevents.ItemIsSymlink:     "IsSymLink",
	}

	note := ""
	for bit, description := range noteDescription {
		if flags&bit == bit {
			note += description + " "
		}
	}
	return note
}
