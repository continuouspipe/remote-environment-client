//The FSEvents API in Mac OS X allows applications to register for notifications of changes to a given directory tree.
//This files contains an implementation of the DirectoryMonitor interface that must be used for OSX
package monitor

import (
	"github.com/fsnotify/fsevents"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"fmt"
	"time"
)

func init() {
	dirMonitor = NewFsEvents()
}

type FsEvents struct {
	Exclusions ExclusionProvider
}

func NewFsEvents() *FsEvents {
	return &FsEvents{}
}

func (m *FsEvents) SetExclusions(exclusion ExclusionProvider) {
	m.Exclusions = exclusion
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
		Latency: 2 * time.Second,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.WatchRoot}
	es.Start()
	ec := es.Events

	cplogs.V(5).Infof("Moditoring directory %s", directory)
	cplogs.V(5).Infof("Device UUID %s", fsevents.GetDeviceUUID(dev))

	for msg := range ec {
		for _, e := range msg {
			desc := m.GetEventDescription(e.Flags)
			cplogs.V(1).Infof("filesystem event for %d(%s) %s\n", e.ID, e.Path, desc)
			if (e.Flags & (fsevents.ItemCreated | fsevents.ItemRemoved | fsevents.ItemRenamed | fsevents.ItemModified | fsevents.ItemChangeOwner)) == 0 {
				cplogs.V(5).Infof("skipping event %s on path %s as is not an item create, remove, renamed, modified or changed owner", desc, e.Path)
				continue
			}
			//check if the file matches the exclusion list, if so ignore the event
			match := m.Exclusions.MatchExclusionList(e.Path)
			if match == true {
				cplogs.V(2).Infof("skipping %s %s as is in the exclusion list", desc, e.Path)
				continue
			}

			observer.OnLastChange()
		}
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

		fsevents.ItemCreated:       "Created",
		fsevents.ItemRemoved:       "Removed",
		fsevents.ItemRenamed:       "Renamed",
		fsevents.ItemModified:      "Modified",

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
