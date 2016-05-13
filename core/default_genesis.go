// Copyright 2014 The go-ur Authors
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

package core

import (
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"
)

func NewDefaultGenesisReader() (io.Reader, error) {
	return gzip.NewReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(defaultGenesisBlock)))
}

const defaultGenesisBlock = "H4sIAAAJbogA/6yRzUrEMBSF3yXrWSRN09zMbuyILlQEfYF78+ME0lbaDFSGvruxXYigiwHPIpDknO/k58Keht56tmd85j8kK7Zjr7HzU8bufTUoLXl7VG3ZeMbR9/kep9Mv0etViLdzHvGIGVegEEQeHNVe1rqMYI3UVliQmvualHdSoiPpqDGWHGmU4IVXnARBFbAA73B6iF3MK68RByhrxxhCtOeUP7aW7/rHOJ/+8TbtEHvCaXtYEZQLAYTmNTUNNwAgQzA1OAStvTIamqoyruQOKQ2W7S/XpIq7HdxXVQG85GHEtzLpzynt2A0m3P5X/HXYZVk+AwAA//9qlMK7BgIAAA=="