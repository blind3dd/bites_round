package main

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

func main() {
	pi := 3.14
	var radius float32 = 8.90
	area := areaOfCircleF32(float32(pi), radius)
	errChan := make(chan error, 1)
	go func() error {
		defer func() {
			if r := recover(); r != nil {
				//err := fmt.Errorf("recovered: %v", r.(error).Error())
				err := fmt.Errorf("recovered: %s", r.(string))
				errChan <- err
			}
		}()
		truncated := Round(248.719387)
		fmt.Println("expected number after rounding up for int32:", truncated)
		if truncated != 248 { //248
			errChan <- fmt.Errorf("number custom round for float32: %f", truncated)

			panic("evened float32(248.719387) for circle area is:")
		}
		errChan <- fmt.Errorf("number custom round for float32: %f", truncated)

		fmt.Println(<-errChan)
		return <-errChan
	}()

	fmt.Printf("type conversion bit by bit %v and %v\n", reflect.TypeOf(area), <-errChan)
	fmt.Printf("the circle area is %f", area) //TODO returns  248.71938 instead of 248.719391 for float32 fract
}

func areaOfCircleF32(pi, radius float32) float32 {
	area := float64(pi) * math.Pow(float64(radius), 2)
	//area = math.Floor(area * math.Pow10(5))
	return float32(area)
}

func Modf(f float32) (int float32, frac float32) {
	return modf(f)
}

// [ 0, 0.499 ) U [ 0.499, 0.5 ) U [ 0.5, 1 >
func Round(x float32) float32 {

	//p = float32(math.Float32bits(100000))
	n1, n2 := modf(x)
	n := int(n1)
	if n2 > .9 {
		bits := math.Float32bits(x)
		var exponent = bits >> shift & (mask) // type casting compat
		// exponent := uint(bits>>shift)
		if exponent >= threshold {
			half := maskfrac - 1 // how many numbers are between 0,987128 and 0,99? INF. SIGN. leading.
			exponent -= threshold
			// bigger the precision mask lesser value + half*2 + epsilon
			bits += (uint32(half) + (bits>>(uint32(shift)-exponent))&1) >> exponent
			bits &^= maskfrac >> exponent // 10100110
		} else if exponent == threshold-1 {
			isMaskFractured := bits & maskfrac
			if isMaskFractured != 0 {
				bits = bits&masknsign | maxbuckets
			}
		} else {
			bits = bits & masknsign
		}
		return bitsToFloat32(bits)

	} else {
		return bitsToFloat32(math.Float32bits(float32(n)))
	}

}

// like int to boolean conversion
func bitsToFloat32(b uint32) float32 {
	return *(*float32)(unsafe.Pointer(&b))
}

func modf(f float32) (int float32, frac float32) {
	if f < 1 {
		switch {
		case f < 0:
			int, frac = Modf(-f)
			return -int, -frac
		case f == 0:
			return f, f // Return -0, -0 when f == -0
		}
		return 0, f
	}

	x := math.Float32bits(f)
	exponent := uint(x>>shift)&mask - threshold

	// Keep the top 9+e bits, the integer part; clear the rest.
	if exponent < 32-9 {
		x &^= 1<<(32-9-exponent) - 1
	}
	int = bitsToFloat32(x)
	frac = f - int
	return
}

const (
	rs         = 123 >> 1 // half + epsilon precision or min threshold by half
	sub        = 32 - shift
	shift      = 1<<5 - 1<<3 - 1<<0 // minus 9 bytes for int part for 32 bytes
	threshold  = 1<<7 - 1           // (2^5 - 2^3 - 2^0 ) uint32 // //threshold =  32 - 8 - 1 // 0.127.255.255
	mask       = 0xFF               // for compat for unsigned int 32 [0,255]
	maxbuckets = 0x3f800000         // 32 bits mask limit in hex
	masknsign  = 1 << 31
	maskfrac   = 1<<shift - 1
)
