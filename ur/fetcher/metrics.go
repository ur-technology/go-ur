// Copyright 2015 The go-ur Authors
// This file is part of the go-ur library.
//
// The go-ur library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ur library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ur library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/ur/go-ur/metrics"
)

var (
	propAnnounceInMeter   = metrics.NewMeter("ur/fetcher/prop/announces/in")
	propAnnounceOutTimer  = metrics.NewTimer("ur/fetcher/prop/announces/out")
	propAnnounceDropMeter = metrics.NewMeter("ur/fetcher/prop/announces/drop")
	propAnnounceDOSMeter  = metrics.NewMeter("ur/fetcher/prop/announces/dos")

	propBroadcastInMeter   = metrics.NewMeter("ur/fetcher/prop/broadcasts/in")
	propBroadcastOutTimer  = metrics.NewTimer("ur/fetcher/prop/broadcasts/out")
	propBroadcastDropMeter = metrics.NewMeter("ur/fetcher/prop/broadcasts/drop")
	propBroadcastDOSMeter  = metrics.NewMeter("ur/fetcher/prop/broadcasts/dos")

	blockFetchMeter  = metrics.NewMeter("ur/fetcher/fetch/blocks")
	headerFetchMeter = metrics.NewMeter("ur/fetcher/fetch/headers")
	bodyFetchMeter   = metrics.NewMeter("ur/fetcher/fetch/bodies")

	blockFilterInMeter   = metrics.NewMeter("ur/fetcher/filter/blocks/in")
	blockFilterOutMeter  = metrics.NewMeter("ur/fetcher/filter/blocks/out")
	headerFilterInMeter  = metrics.NewMeter("ur/fetcher/filter/headers/in")
	headerFilterOutMeter = metrics.NewMeter("ur/fetcher/filter/headers/out")
	bodyFilterInMeter    = metrics.NewMeter("ur/fetcher/filter/bodies/in")
	bodyFilterOutMeter   = metrics.NewMeter("ur/fetcher/filter/bodies/out")
)
