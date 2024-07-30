package trust

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/masnax/microtest/internal/sys"
)

// Store represents a directory of remotes watched by the fsnotify Watcher.
type Store struct {
	remotesMu sync.RWMutex // Mutex for coordinating manual and fsnotify access to remotes.
	remotes   *Remotes     // Should never be called directly, instead use Remotes().

	refresh func(path string) error
}

// Init initializes the remotes in the truststore, seeds the rand package for selecting remotes at random, and watches
// the truststore directory for updates.
func Init(watcher *sys.Watcher, onUpdate func(oldRemotes, newRemotes Remotes) error, dir string) (*Store, error) {
	ts := &Store{remotes: &Remotes{}}
	ts.remotesMu.Lock()
	defer ts.remotesMu.Unlock()

	ts.remotes.data = map[string]Remote{}
	err := ts.remotes.Load(dir)
	if err != nil {
		return nil, err
	}

	ts.refresh = func(path string) error {
		ts.remotesMu.Lock()
		defer func() {
			ts.remotesMu.Unlock()
		}()

		err := ts.remotes.Load(dir)
		if err != nil {
			return fmt.Errorf("Unable to refresh remotes in path %q: %w", path, err)
		}

		return nil
	}

	// Watch on the truststore directory for yaml updates.
	watcher.Watch(dir, "yaml", func(path string, event fsnotify.Op) error {
		return ts.refresh(path)
	})

	return ts, nil
}

// Remotes returns a thread-safe list of the remotes in the truststore, as watched by fsnotify.
func (ts *Store) Remotes() *Remotes {
	ts.remotesMu.RLock()
	defer ts.remotesMu.RUnlock()

	return ts.remotes
}

// Refresh reloads the truststore and runs any associated hooks.
func (ts *Store) Refresh() error {
	return ts.refresh("*")
}
