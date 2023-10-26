// A lock on an id is needed to make sure multiple change operations are not
// occuring at the same time.
//
// This achieved by [Svc] holds a map, which is guarded by a mutex.
// Each update operation is racing to set the id of the map to a channel
//
// When trying to lock the id:
//  1. lock the map
//  2. check if the id has a corresponding channel
//  3. if the id does have a channel, unlock the mutex. wait on the channel, until
//     the channel is closed, then go to 1.
//  4. if the id doesn't have the channel,
//     create a new channel, and set it to the id in the map, unlock the mutex,
//     return channel to the calling operation,
//
// When unlocking the id, the calling operation will lock the mutex on the map,
// delete the channel from the map, then close the channel to notify
// all other waiting operations.

package svc

// emptyForChan is just that
type emptyForChan struct{}

// waitingChan contains the waiting channel
type waitingChan struct {
	c <-chan emptyForChan
}

func newWaiter() (*waitingChan, chan<- emptyForChan) {
	c := make(chan emptyForChan)
	return &waitingChan{
		c: c,
	}, c
}

// Done is like context's Done function, wait on the channel
// for the cancellation
func (w *waitingChan) Done() <-chan emptyForChan {
	return w.c
}

// lockId tries to lock the given id.
//
//  1. lock the map
//  2. check if the id has corresponding waitingChan
//  3. if the id has waitingChan, unlock the map, wait on the waitingChan,
//     until the waitingChan is closed. Go to step 1.
//  4. if the id doesn't have waitingChan, create a new waitingChan,
//     set it to the id, unlock the map, and return the closer
func (s *Svc) lockId(id string) chan<- emptyForChan {
	s.dbmutex.Lock()
	if s.idmutex == nil {
		s.idmutex = make(map[string]*waitingChan)
	}

	var result chan<- emptyForChan

	m, found := s.idmutex[id]
waitloop:
	for {
		if !found {
			break waitloop
		}
		s.dbmutex.Unlock()
		<-m.Done()
		s.dbmutex.Lock()
		m, found = s.idmutex[id]
	}

	s.idmutex[id], result = newWaiter()
	s.dbmutex.Unlock()
	return result
}

func (s *Svc) unlockId(id string, closer chan<- emptyForChan) {
	s.dbmutex.Lock()
	delete(s.idmutex, id)
	close(closer)
	s.dbmutex.Unlock()
}
