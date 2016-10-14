// Copyright 2015 The go-ur Authors
// This file is part of go-ur.
//
// go-ur is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ur is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ur. If not, see <http://www.gnu.org/licenses/>.

// Simple wrapper to translate the API exposed methods and types to inthernal
// Go versions of the same types.

#include "_cgo_export.h"

int run(const char* args) {
 return doRun((char*)args);
}
