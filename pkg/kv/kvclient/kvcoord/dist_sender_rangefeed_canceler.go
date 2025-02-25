// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
//

package kvcoord

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

// stuckRangeFeedCanceler are a defense-in-depth mechanism to restart rangefeeds that have
// not received events from the KV layer in some time. Rangefeeds are supposed to receive
// regular updates, as at the very least they ought to be receiving closed timestamps.
//
// However, issues[^1] at the KV layer could prevent this.
//
// The canceler is notified via ping() whenever the associated RangeFeed receives an event.
// Should ping() not be called for the configured threshold duration, the provided cancel
// function will be invoked.
//
// This is implemented without incurring nontrivial work on each call to ping().
// Instead, work is done roughly on each threshold interval, which is assumed to
// be large enough (i.e. at least a couple of seconds) to make this negligible.
// Concretely, a timer is set that would invoke the cancellation, and the timer
// is reset on the first call to ping() after the timer is at least half
// expired. That way, we allocate only ~twice per eventCheckInterval, which is
// acceptable.
//
// The canceler detects changes to the configured threshold duration on each call
// to ping(), i.e. in the common case of no stuck rangefeeds, it will ~immediately
// pick up the new value and apply it.
//
// [^1]: https://github.com/cockroachdb/cockroach/issues/86818
type stuckRangeFeedCanceler struct {
	threshold       func() time.Duration
	cancel          context.CancelFunc
	t               *time.Timer
	resetTimerAfter time.Time
	activeThreshold time.Duration

	_stuck int32 // atomic
}

// stuck returns true if the stuck detection got triggered.
// If this returns true, the cancel function will be invoked
// shortly, if it hasn't already.
func (w *stuckRangeFeedCanceler) stuck() bool {
	return atomic.LoadInt32(&w._stuck) != 0
}

// stop releases the active timer, if any. It should be invoked
// unconditionally before the canceler goes out of scope.
func (w *stuckRangeFeedCanceler) stop() {
	if w.t != nil {
		w.t.Stop()
		w.t = nil
		w.activeThreshold = 0
	}
}

// ping notifies the canceler that the rangefeed has received an
// event, i.e. is making progress.
func (w *stuckRangeFeedCanceler) ping() {
	threshold := w.threshold()
	if threshold == 0 {
		w.stop()
		return
	}

	mkTimer := func() {
		w.activeThreshold = threshold
		// The timer will fire after 1.5*threshold so that when it does
		// *at least* the threshold has passed. For example, with a
		// 60s threshold, if the timer starts at time 0, and the last
		// ping() event arrives at 29.999s, the timer should only fire
		// at 90s, not 60s.
		w.t = time.AfterFunc(3*threshold/2, func() {
			// NB: important to store _stuck before canceling, since we
			// want the caller to be able to detect stuck() after ctx
			// cancels.
			atomic.StoreInt32(&w._stuck, 1)
			w.cancel()
		})
		w.resetTimerAfter = timeutil.Now().Add(threshold / 2)
	}

	if w.t == nil {
		mkTimer()
	} else if w.resetTimerAfter.Before(timeutil.Now()) || w.activeThreshold != threshold {
		w.stop()
		mkTimer()
	}
}

// newStuckRangeFeedCanceler sets up a canceler with the provided
// cancel function (which should cancel the rangefeed if invoked)
// and uses the kv.rangefeed.range_stuck_threshold cluster setting
// to (reactively) configure the timeout.
//
// The timer will only activate with the first call to ping.
func newStuckRangeFeedCanceler(
	cancel context.CancelFunc, threshold func() time.Duration,
) *stuckRangeFeedCanceler {
	w := &stuckRangeFeedCanceler{
		threshold: threshold,
		cancel:    cancel,
	}
	return w
}
