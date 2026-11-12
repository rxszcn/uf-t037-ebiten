#version 150

// modInt is a truncated modulo like Go's '%' operator. It is built on divInt so the
// zero-divisor and minimum-integer cases stay defined on every backend.
int modInt(int x, int y) {
	return x - y*divInt(x, y);
}

ivec2 modInt(ivec2 x, int y) {
	return x - y*divInt(x, y);
}

ivec3 modInt(ivec3 x, int y) {
	return x - y*divInt(x, y);
}

ivec4 modInt(ivec4 x, int y) {
	return x - y*divInt(x, y);
}

ivec2 modInt(ivec2 x, ivec2 y) {
	return x - y*divInt(x, y);
}

ivec3 modInt(ivec3 x, ivec3 y) {
	return x - y*divInt(x, y);
}

ivec4 modInt(ivec4 x, ivec4 y) {
	return x - y*divInt(x, y);
}

// divInt is a truncated integer division like Go's '/' operator. The y == -1 and
// x == INT_MIN case is defined in Go (the result is INT_MIN) even though the native
// division instruction overflows, and a zero divisor is defined to return zero instead
// of invoking the GPU's undefined behavior.
int divInt(int x, int y) {
	if (y == 0) {
		return 0;
	}
	if (y == -1 && x == (-2147483647 - 1)) {
		return x;
	}
	return x / y;
}

ivec2 divInt(ivec2 x, int y) {
	if (y == 0) {
		return ivec2(0);
	}
	if (y == -1) {
		return ivec2(x.x == (-2147483647 - 1) ? x.x : x.x / y, x.y == (-2147483647 - 1) ? x.y : x.y / y);
	}
	return x / y;
}

ivec3 divInt(ivec3 x, int y) {
	if (y == 0) {
		return ivec3(0);
	}
	if (y == -1) {
		return ivec3(x.x == (-2147483647 - 1) ? x.x : x.x / y, x.y == (-2147483647 - 1) ? x.y : x.y / y, x.z == (-2147483647 - 1) ? x.z : x.z / y);
	}
	return x / y;
}

ivec4 divInt(ivec4 x, int y) {
	if (y == 0) {
		return ivec4(0);
	}
	if (y == -1) {
		return ivec4(x.x == (-2147483647 - 1) ? x.x : x.x / y, x.y == (-2147483647 - 1) ? x.y : x.y / y, x.z == (-2147483647 - 1) ? x.z : x.z / y, x.w == (-2147483647 - 1) ? x.w : x.w / y);
	}
	return x / y;
}

ivec2 divInt(ivec2 x, ivec2 y) {
	return ivec2(y.x == 0 ? 0 : (y.x == -1 && x.x == (-2147483647 - 1) ? x.x : x.x / y.x), y.y == 0 ? 0 : (y.y == -1 && x.y == (-2147483647 - 1) ? x.y : x.y / y.y));
}

ivec3 divInt(ivec3 x, ivec3 y) {
	return ivec3(y.x == 0 ? 0 : (y.x == -1 && x.x == (-2147483647 - 1) ? x.x : x.x / y.x), y.y == 0 ? 0 : (y.y == -1 && x.y == (-2147483647 - 1) ? x.y : x.y / y.y), y.z == 0 ? 0 : (y.z == -1 && x.z == (-2147483647 - 1) ? x.z : x.z / y.z));
}

ivec4 divInt(ivec4 x, ivec4 y) {
	return ivec4(y.x == 0 ? 0 : (y.x == -1 && x.x == (-2147483647 - 1) ? x.x : x.x / y.x), y.y == 0 ? 0 : (y.y == -1 && x.y == (-2147483647 - 1) ? x.y : x.y / y.y), y.z == 0 ? 0 : (y.z == -1 && x.z == (-2147483647 - 1) ? x.z : x.z / y.z), y.w == 0 ? 0 : (y.w == -1 && x.w == (-2147483647 - 1) ? x.w : x.w / y.w));
}

// shiftLeftInt and shiftRightInt implement Go's shift operators. Go defines shifts with
// a count of 32 bits or more as zero, whereas GLSL leaves such shifts (and negative
// signed left shifts) undefined. The count is clamped to [0, 32] so every backend gets
// the same defined result. Left shifts go through uint to avoid undefined behavior on
// negative signed values; right shifts stay signed to keep the arithmetic shift.
int shiftLeftInt(int x, int count) {
	return int(uint(x) << uint(clamp(count, 0, 32)));
}

ivec2 shiftLeftInt(ivec2 x, int count) {
	return ivec2(uvec2(x) << uvec2(uint(clamp(count, 0, 32))));
}

ivec3 shiftLeftInt(ivec3 x, int count) {
	return ivec3(uvec3(x) << uvec3(uint(clamp(count, 0, 32))));
}

ivec4 shiftLeftInt(ivec4 x, int count) {
	return ivec4(uvec4(x) << uvec4(uint(clamp(count, 0, 32))));
}

ivec2 shiftLeftInt(ivec2 x, ivec2 count) {
	return ivec2(uvec2(x) << clamp(uvec2(count), uvec2(0), uvec2(32)));
}

ivec3 shiftLeftInt(ivec3 x, ivec3 count) {
	return ivec3(uvec3(x) << clamp(uvec3(count), uvec3(0), uvec3(32)));
}

ivec4 shiftLeftInt(ivec4 x, ivec4 count) {
	return ivec4(uvec4(x) << clamp(uvec4(count), uvec4(0), uvec4(32)));
}

int shiftRightInt(int x, int count) {
	return x >> clamp(count, 0, 32);
}

ivec2 shiftRightInt(ivec2 x, int count) {
	return x >> clamp(count, 0, 32);
}

ivec3 shiftRightInt(ivec3 x, int count) {
	return x >> clamp(count, 0, 32);
}

ivec4 shiftRightInt(ivec4 x, int count) {
	return x >> clamp(count, 0, 32);
}

ivec2 shiftRightInt(ivec2 x, ivec2 count) {
	return x >> clamp(ivec2(count), ivec2(0), ivec2(32));
}

ivec3 shiftRightInt(ivec3 x, ivec3 count) {
	return x >> clamp(ivec3(count), ivec3(0), ivec3(32));
}

ivec4 shiftRightInt(ivec4 x, ivec4 count) {
	return x >> clamp(ivec4(count), ivec4(0), ivec4(32));
}

uniform vec2 U0[4];

float F0(void);

float F0(void) {
	float l0[3];
	l0[0] = float(0);
	l0[1] = float(0);
	l0[2] = float(0);
	float l1 = float(0);
	float l5 = float(0);
	vec2 l7 = vec2(0);
	l1 = 0.0;
	for (int l2 = 0; l2 < 3; l2++) {
		l1 = (l1) + (1.0);
	}
	for (int l3 = 0; l3 < 3; l3++) {
		l1 = (l1) + ((l0)[l3]);
	}
	for (int l4 = 0; l4 < 3; l4++) {
		l5 = (l0)[l4];
		l1 = (l1) + ((float(l4)) * (l5));
	}
	for (int l6 = 0; l6 < 4; l6++) {
		l7 = (U0)[l6];
		l1 = (l1) + ((l7).x);
	}
	return l1;
}
