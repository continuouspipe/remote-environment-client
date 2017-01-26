package monitor

import "runtime"

type EventsObserver interface {
	OnLastChange() error
}

type DirectoryMonitor interface {
	//when an event occurs it executes the callback with the supplied arguments
	AnyEventCall(directory string, observer EventsObserver) error
	SetExclusions(exclusion ExclusionProvider)
}

func GetOsDirectoryMonitor() DirectoryMonitor {
	var dirMonitor DirectoryMonitor
	if runtime.GOOS == "darwin" {
		dirMonitor = NewFsEvents()
	} else {
		dirMonitor = NewFsWatch()
	}

	exclusion := NewExclusion()
	exclusion.LoadCustomExclusionsFromFile()
	dirMonitor.SetExclusions(exclusion)
	return dirMonitor
}
