// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

var (
	horizontalRule = regexp.MustCompile("^===[=]*===$")
)

func Test(t *testing.T) {
	tests := []struct {
		condition   string
		name        string
		roots       []string
		stripPrefix string
		expectedOut []matcher
	}{
		{
			condition: "firstparty",
			name:      "apex",
			roots:     []string{"highest.apex.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"highest.apex"},
				usedBy{"highest.apex/bin/bin1"},
				usedBy{"highest.apex/bin/bin2"},
				usedBy{"highest.apex/lib/liba.so"},
				usedBy{"highest.apex/lib/libb.so"},
				firstParty{},
			},
		},
		{
			condition: "firstparty",
			name:      "container",
			roots:     []string{"container.zip.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"container.zip"},
				usedBy{"container.zip/bin1"},
				usedBy{"container.zip/bin2"},
				usedBy{"container.zip/liba.so"},
				usedBy{"container.zip/libb.so"},
				firstParty{},
			},
		},
		{
			condition: "firstparty",
			name:      "application",
			roots:     []string{"application.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"application"},
				firstParty{},
			},
		},
		{
			condition: "firstparty",
			name:      "binary",
			roots:     []string{"bin/bin1.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"bin/bin1"},
				firstParty{},
			},
		},
		{
			condition: "firstparty",
			name:      "library",
			roots:     []string{"lib/libd.so.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"lib/libd.so"},
				firstParty{},
			},
		},
		{
			condition: "notice",
			name:      "apex",
			roots:     []string{"highest.apex.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"highest.apex"},
				usedBy{"highest.apex/bin/bin1"},
				usedBy{"highest.apex/bin/bin2"},
				usedBy{"highest.apex/lib/libb.so"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"highest.apex/bin/bin1"},
				usedBy{"highest.apex/lib/liba.so"},
				library{"External"},
				usedBy{"highest.apex/bin/bin1"},
				notice{},
			},
		},
		{
			condition: "notice",
			name:      "container",
			roots:     []string{"container.zip.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"container.zip"},
				usedBy{"container.zip/bin1"},
				usedBy{"container.zip/bin2"},
				usedBy{"container.zip/libb.so"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"container.zip/bin1"},
				usedBy{"container.zip/liba.so"},
				library{"External"},
				usedBy{"container.zip/bin1"},
				notice{},
			},
		},
		{
			condition: "notice",
			name:      "application",
			roots:     []string{"application.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"application"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"application"},
				notice{},
			},
		},
		{
			condition: "notice",
			name:      "binary",
			roots:     []string{"bin/bin1.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"bin/bin1"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"bin/bin1"},
				library{"External"},
				usedBy{"bin/bin1"},
				notice{},
			},
		},
		{
			condition: "notice",
			name:      "library",
			roots:     []string{"lib/libd.so.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"External"},
				usedBy{"lib/libd.so"},
				notice{},
			},
		},
		{
			condition: "reciprocal",
			name:      "apex",
			roots:     []string{"highest.apex.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"highest.apex"},
				usedBy{"highest.apex/bin/bin1"},
				usedBy{"highest.apex/bin/bin2"},
				usedBy{"highest.apex/lib/libb.so"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"highest.apex/bin/bin1"},
				usedBy{"highest.apex/lib/liba.so"},
				library{"External"},
				usedBy{"highest.apex/bin/bin1"},
				reciprocal{},
			},
		},
		{
			condition: "reciprocal",
			name:      "container",
			roots:     []string{"container.zip.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"container.zip"},
				usedBy{"container.zip/bin1"},
				usedBy{"container.zip/bin2"},
				usedBy{"container.zip/libb.so"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"container.zip/bin1"},
				usedBy{"container.zip/liba.so"},
				library{"External"},
				usedBy{"container.zip/bin1"},
				reciprocal{},
			},
		},
		{
			condition: "reciprocal",
			name:      "application",
			roots:     []string{"application.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"application"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"application"},
				reciprocal{},
			},
		},
		{
			condition: "reciprocal",
			name:      "binary",
			roots:     []string{"bin/bin1.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"bin/bin1"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"bin/bin1"},
				library{"External"},
				usedBy{"bin/bin1"},
				reciprocal{},
			},
		},
		{
			condition: "reciprocal",
			name:      "library",
			roots:     []string{"lib/libd.so.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"External"},
				usedBy{"lib/libd.so"},
				notice{},
			},
		},
		{
			condition: "restricted",
			name:      "apex",
			roots:     []string{"highest.apex.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"highest.apex"},
				usedBy{"highest.apex/bin/bin1"},
				usedBy{"highest.apex/bin/bin2"},
				firstParty{},
				hr{},
				library{"Android"},
				usedBy{"highest.apex/bin/bin2"},
				usedBy{"highest.apex/lib/libb.so"},
				library{"Device"},
				usedBy{"highest.apex/bin/bin1"},
				usedBy{"highest.apex/lib/liba.so"},
				restricted{},
				hr{},
				library{"External"},
				usedBy{"highest.apex/bin/bin1"},
				reciprocal{},
			},
		},
		{
			condition: "restricted",
			name:      "container",
			roots:     []string{"container.zip.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"container.zip"},
				usedBy{"container.zip/bin1"},
				usedBy{"container.zip/bin2"},
				firstParty{},
				hr{},
				library{"Android"},
				usedBy{"container.zip/bin2"},
				usedBy{"container.zip/libb.so"},
				library{"Device"},
				usedBy{"container.zip/bin1"},
				usedBy{"container.zip/liba.so"},
				restricted{},
				hr{},
				library{"External"},
				usedBy{"container.zip/bin1"},
				reciprocal{},
			},
		},
		{
			condition: "restricted",
			name:      "application",
			roots:     []string{"application.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"application"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"application"},
				restricted{},
			},
		},
		{
			condition: "restricted",
			name:      "binary",
			roots:     []string{"bin/bin1.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"bin/bin1"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"bin/bin1"},
				restricted{},
				hr{},
				library{"External"},
				usedBy{"bin/bin1"},
				reciprocal{},
			},
		},
		{
			condition: "restricted",
			name:      "library",
			roots:     []string{"lib/libd.so.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"External"},
				usedBy{"lib/libd.so"},
				notice{},
			},
		},
		{
			condition: "proprietary",
			name:      "apex",
			roots:     []string{"highest.apex.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"highest.apex/bin/bin2"},
				usedBy{"highest.apex/lib/libb.so"},
				restricted{},
				hr{},
				library{"Android"},
				usedBy{"highest.apex"},
				usedBy{"highest.apex/bin/bin1"},
				firstParty{},
				hr{},
				library{"Android"},
				usedBy{"highest.apex/bin/bin2"},
				library{"Device"},
				usedBy{"highest.apex/bin/bin1"},
				usedBy{"highest.apex/lib/liba.so"},
				library{"External"},
				usedBy{"highest.apex/bin/bin1"},
				proprietary{},
			},
		},
		{
			condition: "proprietary",
			name:      "container",
			roots:     []string{"container.zip.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"container.zip/bin2"},
				usedBy{"container.zip/libb.so"},
				restricted{},
				hr{},
				library{"Android"},
				usedBy{"container.zip"},
				usedBy{"container.zip/bin1"},
				firstParty{},
				hr{},
				library{"Android"},
				usedBy{"container.zip/bin2"},
				library{"Device"},
				usedBy{"container.zip/bin1"},
				usedBy{"container.zip/liba.so"},
				library{"External"},
				usedBy{"container.zip/bin1"},
				proprietary{},
			},
		},
		{
			condition: "proprietary",
			name:      "application",
			roots:     []string{"application.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"application"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"application"},
				proprietary{},
			},
		},
		{
			condition: "proprietary",
			name:      "binary",
			roots:     []string{"bin/bin1.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"Android"},
				usedBy{"bin/bin1"},
				firstParty{},
				hr{},
				library{"Device"},
				usedBy{"bin/bin1"},
				library{"External"},
				usedBy{"bin/bin1"},
				proprietary{},
			},
		},
		{
			condition: "proprietary",
			name:      "library",
			roots:     []string{"lib/libd.so.meta_lic"},
			expectedOut: []matcher{
				hr{},
				library{"External"},
				usedBy{"lib/libd.so"},
				notice{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.condition+" "+tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			rootFiles := make([]string, 0, len(tt.roots))
			for _, r := range tt.roots {
				rootFiles = append(rootFiles, "testdata/"+tt.condition+"/"+r)
			}

			ctx := context{stdout, stderr, os.DirFS("."), tt.stripPrefix}

			err := textNotice(&ctx, rootFiles...)
			if err != nil {
				t.Fatalf("textnotice: error = %v, stderr = %v", err, stderr)
				return
			}
			if stderr.Len() > 0 {
				t.Errorf("textnotice: gotStderr = %v, want none", stderr)
			}

			t.Logf("got stdout: %s", stdout.String())

			t.Logf("want stdout: %s", matcherList(tt.expectedOut).String())

			out := bufio.NewScanner(stdout)
			lineno := 0
			for out.Scan() {
				line := out.Text()
				if strings.TrimLeft(line, " ") == "" {
					continue
				}
				if len(tt.expectedOut) <= lineno {
					t.Errorf("unexpected output at line %d: got %q, want nothing (wanted %d lines)", lineno+1, line, len(tt.expectedOut))
				} else if !tt.expectedOut[lineno].isMatch(line) {
					t.Errorf("unexpected output at line %d: got %q, want %q", lineno+1, line, tt.expectedOut[lineno].String())
				}
				lineno++
			}
			for ; lineno < len(tt.expectedOut); lineno++ {
				t.Errorf("textnotice: missing output line %d: ended early, want %q", lineno+1, tt.expectedOut[lineno].String())
			}
		})
	}
}

type matcher interface {
	isMatch(line string) bool
	String() string
}

type hr struct{}

func (m hr) isMatch(line string) bool {
	return horizontalRule.MatchString(line)
}

func (m hr) String() string {
	return " ================================================== "
}

type library struct {
	name string
}

func (m library) isMatch(line string) bool {
	return strings.HasPrefix(line, m.name+" ")
}

func (m library) String() string {
	return m.name + " used by:"
}

type usedBy struct {
	name string
}

func (m usedBy) isMatch(line string) bool {
	return len(line) > 0 && line[0] == ' ' && strings.HasPrefix(strings.TrimLeft(line, " "), "out/") && strings.HasSuffix(line, "/"+m.name)
}

func (m usedBy) String() string {
	return "  out/.../" + m.name
}

type firstParty struct{}

func (m firstParty) isMatch(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), "&&&First Party License&&&")
}

func (m firstParty) String() string {
	return "&&&First Party License&&&"
}

type notice struct{}

func (m notice) isMatch(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), "%%%Notice License%%%")
}

func (m notice) String() string {
	return "%%%Notice License%%%"
}

type reciprocal struct{}

func (m reciprocal) isMatch(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), "$$$Reciprocal License$$$")
}

func (m reciprocal) String() string {
	return "$$$Reciprocal License$$$"
}

type restricted struct{}

func (m restricted) isMatch(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), "###Restricted License###")
}

func (m restricted) String() string {
	return "###Restricted License###"
}

type proprietary struct{}

func (m proprietary) isMatch(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), "@@@Proprietary License@@@")
}

func (m proprietary) String() string {
	return "@@@Proprietary License@@@"
}

type matcherList []matcher

func (l matcherList) String() string {
	var sb strings.Builder
	for _, m := range l {
		s := m.String()
		if s[:3] == s[len(s)-3:] {
			fmt.Fprintln(&sb)
		}
		fmt.Fprintf(&sb, "%s\n", s)
		if s[:3] == s[len(s)-3:] {
			fmt.Fprintln(&sb)
		}
	}
	return sb.String()
}
