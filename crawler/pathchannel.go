package crawler

import (
	"errors"

	"github.com/eapache/channels"
)

// pathChannel is an infinite channel structure which contains paths to visit.
// NOTE (bitantics): currently doesn't avoid duplicates in the queue. How to
// do this without increasing memory by O(n)?
type pathChannel struct {
	ch *channels.InfiniteChannel
}

// newPathChannel creates a new infinite channel of paths
func newPathChannel() *pathChannel {
	return &pathChannel{ch: channels.NewInfiniteChannel()}
}

// Put places a new path in the channel
func (pc *pathChannel) Put(path string) {
	pc.ch.In() <- path
}

// Get takes a path from the channel. Blocks until a path becomes available
func (pc *pathChannel) Get() (string, error) {
	select {
	case path := <-pc.ch.Out():
		if path == nil {
			return "", errors.New("path channel: closed")
		}

		return path.(string), nil
	}
}

// Close the channel, making all Get() calls unblock, returning errors
func (pc *pathChannel) Close() {
	pc.ch.Close()
}
