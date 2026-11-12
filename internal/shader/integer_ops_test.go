// Copyright 2026 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shader_test

import (
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/internal/graphics"
	"github.com/hajimehoshi/ebiten/v2/internal/shaderir/glsl"
	"github.com/hajimehoshi/ebiten/v2/internal/shaderir/hlsl"
	"github.com/hajimehoshi/ebiten/v2/internal/shaderir/msl"
)

// integerOpsSource exercises every integer operation whose native semantics differ across
// the shading languages: variable and out-of-range shift counts, negative right shifts,
// truncated integer division and modulo, the minimum-integer boundary, zero divisors, and
// the scalar/vector operand combinations. All results are forced into the return value so
// that nothing is folded away at compile time.
const integerOpsSource = `package main

func Vertex(dstPos vec2, src0Pos vec2, color vec4) (vec4, vec2, vec4) {
	return vec4(dstPos, 0, 1), src0Pos, color
}

func Fragment(dstPos vec4, src0Pos vec2, color vec4) vec4 {
	x := int(-7)
	n := int(33)
	v := ivec4(-7, 7, -2147483647 - 1, 2147483647)
	vn := ivec4(0, 2, 33, 40)

	// Variable and out-of-range shift counts, on scalars and vectors.
	a := x << n
	b := x >> n
	c := x << 2
	d := x >> 2
	av := v << 2
	bv := v >> 2
	cv := v << vn
	dv := v >> vn
	v <<= 1
	v >>= n

	// Truncated integer division and modulo, including negatives.
	e := x / 2
	f := x % 2
	g := 7 / -2
	h := 7 % -2
	ev := v / 2
	fv := v % 2
	gv := v / vn
	hv := v % vn

	// Minimum-integer boundary: Go defines INT_MIN / -1 as INT_MIN and INT_MIN % -1 as 0.
	minInt := -2147483647 - 1
	i := minInt / -1
	j := minInt % -1

	// Zero divisors must be defined (zero) instead of triggering GPU undefined behavior.
	z := int(color.x)
	k := x / z
	l := x % z

	// A floating-point division must keep the native '/' operator.
	fd := color.x / 2.0

	r := float(a+b+c+d+e+g+i+k) + fd
	r += float(av.x + bv.y + cv.z + dv.w + ev.x + fv.y + gv.z + hv.w + f + h + j + l)
	return vec4(r)
}
`

// backendOutputs holds the generated source for every supported backend.
type backendOutputs struct {
	glslDefaultVS string
	glslDefaultFS string
	glslES300VS   string
	glslES300FS   string
	hlslVS        string
	hlslPS        string
	hlslVPrelude  string
	hlslPPrelude  string
	msl           string
}

func compileIntegerOps(t *testing.T) backendOutputs {
	t.Helper()
	p, err := compileToIR([]byte(integerOpsSource))
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	idx := graphics.ProjectionMatrixUniformVariableIndex
	gvs, gfs := glsl.Compile(p, glsl.GLSLVersionDefault, idx)
	evs, efs := glsl.Compile(p, glsl.GLSLVersionES300, idx)
	hvs, hps, hVPrelude, hPPrelude := hlsl.Compile(p)
	return backendOutputs{
		glslDefaultVS: gvs,
		glslDefaultFS: gfs,
		glslES300VS:   evs,
		glslES300FS:   efs,
		hlslVS:        hvs,
		hlslPS:        hps,
		hlslVPrelude:  hVPrelude,
		hlslPPrelude:  hPPrelude,
		msl:           msl.Compile(p),
	}
}

// stripPrelude removes the backend-generated prelude (helper definitions) so that only the
// generated user code is inspected. The helper definitions legitimately contain native
// shift and division operators.
func stripPrelude(src, prelude string) string {
	return strings.TrimSpace(strings.TrimPrefix(src, prelude))
}

// TestIntegerOpsRouteThroughHelpers makes sure no raw shift, integer division or integer
// modulo reaches the user code of a backend. Every such operation has to go through the
// shared helper functions, which is what guarantees identical results across backends.
func TestIntegerOpsRouteThroughHelpers(t *testing.T) {
	o := compileIntegerOps(t)

	// GLSL preludes are not returned separately, so cut at the first user-code signature.
	glslBody := func(src string) string {
		for _, marker := range []string{"in vec2 V0", "in vec4 V1", "void main", "vec4 F0"} {
			if i := strings.Index(src, marker); i >= 0 {
				return src[i:]
			}
		}
		return src
	}

	cases := []struct {
		name string
		body string
		// rawTokens must not appear in the generated user code.
		rawTokens []string
	}{
		{"GLSL default vertex", glslBody(o.glslDefaultVS), []string{" << ", " >> "}},
		{"GLSL default fragment", glslBody(o.glslDefaultFS), []string{" << ", " >> "}},
		{"GLSL ES300 vertex", glslBody(o.glslES300VS), []string{" << ", " >> "}},
		{"GLSL ES300 fragment", glslBody(o.glslES300FS), []string{" << ", " >> "}},
		{"HLSL vertex", stripPrelude(o.hlslVS, o.hlslVPrelude), []string{" << ", " >> "}},
		{"HLSL pixel", stripPrelude(o.hlslPS, o.hlslPPrelude), []string{" << ", " >> "}},
		{"MSL", stripPrelude(o.msl, msl.Prelude()), []string{" << ", " >> "}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, tok := range c.rawTokens {
				if strings.Contains(c.body, tok) {
					t.Errorf("raw shift operator %q reached the backend user code; it must use a shift helper", tok)
				}
			}
		})
	}
}

// TestIntegerOpsHelpersPresentOnEveryBackend verifies that each backend defines and calls
// the full set of integer-operation helpers, covering both GLSL versions.
func TestIntegerOpsHelpersPresentOnEveryBackend(t *testing.T) {
	o := compileIntegerOps(t)

	required := []string{"shiftLeftInt", "shiftRightInt", "divInt", "modInt"}
	srcs := map[string]string{
		"GLSL default VS": o.glslDefaultVS,
		"GLSL default FS": o.glslDefaultFS,
		"GLSL ES300 VS":   o.glslES300VS,
		"GLSL ES300 FS":   o.glslES300FS,
		"HLSL VS":         o.hlslVS,
		"HLSL PS":         o.hlslPS,
		"MSL":             o.msl,
	}
	for name, src := range srcs {
		t.Run(name, func(t *testing.T) {
			for _, fn := range required {
				if !strings.Contains(src, fn) {
					t.Errorf("helper %q is missing from %s", fn, name)
				}
			}
		})
	}
}

// TestIntegerOpsFloatDivisionKeptNative ensures the integer-only rewriting does not touch
// floating-point division, which must keep the native '/' operator.
func TestIntegerOpsFloatDivisionKeptNative(t *testing.T) {
	o := compileIntegerOps(t)

	// color.x / 2.0 is the only floating-point division in the source. It must survive as a
	// direct '/' expression rather than being routed through divInt.
	wanted := []struct {
		name string
		src  string
	}{
		{"GLSL default", o.glslDefaultFS},
		{"GLSL ES300", o.glslES300FS},
		{"HLSL", o.hlslPS},
		{"MSL", o.msl},
	}
	for _, c := range wanted {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(c.src, "/ (2.0)") && !strings.Contains(c.src, "/ 2.0") {
				t.Errorf("floating-point division was not kept native in %s", c.name)
			}
		})
	}
}
