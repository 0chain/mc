/*
 * Minio Client (C) 2015 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/minio/minio-xl/pkg/probe"
	. "gopkg.in/check.v1"
)

func (s *TestSuite) TestCat(c *C) {
	/// filesystem
	root, e := ioutil.TempDir(os.TempDir(), "cmd-")
	c.Assert(e, IsNil)
	defer os.RemoveAll(root)

	objectPath := filepath.Join(root, "object1")
	objectPathServer := server.URL + "/bucket/object1"
	data := "hello"

	var err *probe.Error
	err = putTarget(objectPath, bytes.NewReader([]byte(data)), int64(len(data)))
	c.Assert(err, IsNil)
	err = putTarget(objectPathServer, bytes.NewReader([]byte(data)), int64(len(data)))
	c.Assert(err, IsNil)

	var sourceURLs []string
	sourceURLs = append(sourceURLs, objectPath)
	sourceURLs = append(sourceURLs, objectPathServer)
	for _, sourceURL := range sourceURLs {
		c.Assert(catURL(sourceURL), IsNil)
	}

	objectPath = filepath.Join(root, "object2")
	c.Assert(catURL(objectPath), Not(IsNil))
}
