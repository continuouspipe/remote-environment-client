package monitor

type EventsObserver interface {
	OnLastChange() error
}

type DirectoryMonitor interface {
	//when an event occurs it executes the callback with the supplied arguments
	AnyEventCall(directory string, observer EventsObserver) error
	SetExclusions(exclusion ExclusionProvider)
}

//this is initialised by either by fsevents_darwin or fswatch depending on the build constrains
var dirMonitor DirectoryMonitor

func GetOsDirectoryMonitor() DirectoryMonitor {
	exclusion := NewExclusion()
	exclusion.LoadCustomExclusionsFromFile()
	dirMonitor.SetExclusions(exclusion)
	return dirMonitor
}
