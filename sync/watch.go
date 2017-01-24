package sync

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/path/filepath"
	"github.com/fsnotify/fsnotify"
)

type EventsObserver interface {
	OnLastChange() error
}

type DirectoryMonitor interface {
	//when an event occurs it executes the callback with the supplied arguments
	AnyEventCall(directory string, observer EventsObserver) error
}

type RecursiveDirectoryMonitor struct {
	Exclusions []string
}

func GetRecursiveDirectoryMonitor() *RecursiveDirectoryMonitor {
	m := &RecursiveDirectoryMonitor{}
	m.Exclusions = []string{`/\.[^/]*$`,
		`\.idea`,
		`\.git`,
		`___jb_old___`,
		`___jb_tmp___`,
		`cp-remote-logs`}
	return m
}

func (w RecursiveDirectoryMonitor) AnyEventCall(directory string, observer EventsObserver) error {
	// these variables must be accessed while holding the changeLock
	// mutex as they are shared between goroutines to communicate
	// sync state/events.
	var (
		changeLock sync.Mutex
		dirty      bool
		lastChange time.Time
		watchError error
	)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error setting up filesystem watcher: %v", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				changeLock.Lock()
				cplogs.V(1).Infof("filesystem event for %s(%s)\n", event.Name, event.Op)

				//check if the file matches the exclusion list, if so ignore the event
				match := w.matchExclusionList(event.Name)
				if match == true {
					cplogs.V(2).Infof("skipped %s(%s) as is in the exclusion list", event.Name, event.Op)
					changeLock.Unlock()
					continue
				}

				lastChange = time.Now()
				dirty = true
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					if e := watcher.Remove(event.Name); e != nil {
						cplogs.V(5).Infof("error removing watch for %s: %v", event.Name, e)
					}
				} else {
					if e := w.AddRecursiveWatch(watcher, event.Name); e != nil && watchError == nil {
						watchError = e
					}
				}
				changeLock.Unlock()
			case err := <-watcher.Errors:
				changeLock.Lock()
				watchError = fmt.Errorf("error watching filesystem for changes: %v", err)
				changeLock.Unlock()
			}
		}
	}()

	err = w.AddRecursiveWatch(watcher, directory)
	if err != nil {
		return fmt.Errorf("error watching source path %s: %v", directory, err)
	}

	delay := 2 * time.Second
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		changeLock.Lock()
		if watchError != nil {
			return watchError
		}
		// if a change happened more than 'delay' seconds ago, sync it now.
		// if a change happened less than 'delay' seconds ago, sleep for 'delay' seconds
		// and see if more changes happen, we don't want to sync when
		// the filesystem is in the middle of changing due to a massive
		// set of changes (such as a local build in progress).
		if dirty && time.Now().After(lastChange.Add(delay)) {
			fmt.Println("Synchronizing filesystem changes...")
			err = observer.OnLastChange()
			if err != nil {
				return err
			}
			fmt.Println("Done.")
			dirty = false
		}
		changeLock.Unlock()
		<-ticker.C
	}
}

// AddRecursiveWatch handles adding watches recursively for the path provided
// and its subdirectories.  If a non-directory is specified, this call is a no-op.
// Recursive logic from https://github.com/bronze1man/kmg/blob/master/fsnotify/Watcher.go
func (w RecursiveDirectoryMonitor) AddRecursiveWatch(watcher *fsnotify.Watcher, path string) error {
	file, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error introspecting path %s: %v", path, err)
	}
	if !file.IsDir() {
		return nil
	}

	folders, err := filepath.GetSubFolders(path)
	for _, v := range folders {

		//check if the folder matches the exclusion list, if so ignore the event
		match := w.matchExclusionList(v)
		if match == true {
			cplogs.V(5).Infof("skipped path as matches exlusion list, path %s", v)
			continue
		}

		cplogs.V(5).Infof("adding watch on path %s", v)
		err = watcher.Add(v)
		if err != nil {
			// "no space left on device" issues are usually resolved via
			// $ sudo sysctl fs.inotify.max_user_watches=65536
			return fmt.Errorf("error adding watcher for path %s: %v", v, err)
		}
	}
	return nil
}

func (w RecursiveDirectoryMonitor) matchExclusionList(target string) bool {
	for _, elem := range w.Exclusions {
		regex, err := regexp.Compile(elem)
		if err != nil {
			return false
		}
		if res := regex.MatchString(target); res == true {
			return true
		}
	}
	return false
}