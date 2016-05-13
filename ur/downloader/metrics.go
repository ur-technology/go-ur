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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/ur/go-ur/metrics"
)

var (
	hashInMeter      = metrics.NewMeter("ur/downloader/hashes/in")
	hashReqTimer     = metrics.NewTimer("ur/downloader/hashes/req")
	hashDropMeter    = metrics.NewMeter("ur/downloader/hashes/drop")
	hashTimeoutMeter = metrics.NewMeter("ur/downloader/hashes/timeout")

	blockInMeter      = metrics.NewMeter("ur/downloader/blocks/in")
	blockReqTimer     = metrics.NewTimer("ur/downloader/blocks/req")
	blockDropMeter    = metrics.NewMeter("ur/downloader/blocks/drop")
	blockTimeoutMeter = metrics.NewMeter("ur/downloader/blocks/timeout")

	headerInMeter      = metrics.NewMeter("ur/downloader/headers/in")
	headerReqTimer     = metrics.NewTimer("ur/downloader/headers/req")
	headerDropMeter    = metrics.NewMeter("ur/downloader/headers/drop")
	headerTimeoutMeter = metrics.NewMeter("ur/downloader/headers/timeout")

	bodyInMeter      = metrics.NewMeter("ur/downloader/bodies/in")
	bodyReqTimer     = metrics.NewTimer("ur/downloader/bodies/req")
	bodyDropMeter    = metrics.NewMeter("ur/downloader/bodies/drop")
	bodyTimeoutMeter = metrics.NewMeter("ur/downloader/bodies/timeout")

	receiptInMeter      = metrics.NewMeter("ur/downloader/receipts/in")
	receiptReqTimer     = metrics.NewTimer("ur/downloader/receipts/req")
	receiptDropMeter    = metrics.NewMeter("ur/downloader/receipts/drop")
	receiptTimeoutMeter = metrics.NewMeter("ur/downloader/receipts/timeout")

	stateInMeter      = metrics.NewMeter("ur/downloader/states/in")
	stateReqTimer     = metrics.NewTimer("ur/downloader/states/req")
	stateDropMeter    = metrics.NewMeter("ur/downloader/states/drop")
	stateTimeoutMeter = metrics.NewMeter("ur/downloader/states/timeout")
)
