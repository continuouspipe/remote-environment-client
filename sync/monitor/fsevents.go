//The FSEvents API in Mac OS X allows applications to register for notifications of changes to a given directory tree.
//This files contains an implementation of the DirectoryMonitor interface that must be used for OSX
package monitor

import (
	"github.com/fsnotify/fsevents"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"fmt"
	"time"
)

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
	cplogs.V(5).Infof("Device UUID", fsevents.GetDeviceUUID(dev))

	for msg := range ec {
		for _, e := range msg {
			cplogs.V(1).Infof("filesystem event for %s(%s)\n", e.ID, e.Path)
			if (e.Flags & (fsevents.ItemCreated | fsevents.ItemRemoved | fsevents.ItemRenamed | fsevents.ItemModified | fsevents.ItemChangeOwner)) != e.Flags {
				cplogs.V(5).Infof("skipping event %s on path %s as is not an item create, remove, renamed, modified or changed owner", noteDescription[e.Flags], e.Path)
				continue
			}
			//check if the file matches the exclusion list, if so ignore the event
			match := m.Exclusions.MatchExclusionList(e.Path)
			if match == true {
				cplogs.V(2).Infof("skipped %s(%s) as is in the exclusion list", e.Path, noteDescription[e.Flags])
				continue
			}

			observer.OnLastChange()
		}
	}
	return nil
}

var noteDescription = map[fsevents.EventFlags]string{
	fsevents.ItemCreated:       "Created",
	fsevents.ItemRemoved:       "Removed",
	fsevents.ItemRenamed:       "Renamed",
	fsevents.ItemModified:      "Modified",
	fsevents.ItemChangeOwner:   "ChangeOwner",
}
