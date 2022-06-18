package geohash

import (
	"fmt"
	"testing"
)

func TestGeohashEncode(t *testing.T) {
	//var lng, lat = 38.11499, 13.36159
	lngRange, latRange := getGeohashRange()
	//bits := GeohashEncode(lngRange, latRange, lng, lat, DefaultStep)
	//fmt.Println(bits)

	var point1 = []float64{13.81600230932235718, 23.35700050086033031}
	var point2 = []float64{13.82600158452987671, 23.36699997583392729}

	// 两点距离
	dis := geohashGetDistance(point1[0], point1[1], point2[0], point2[1])
	fmt.Println("两点距离：", dis)

	// 求9个范围
	bits := GetGeoHashBitsRange(point1[0], point1[1], 10)
	fmt.Println("9个矩形最小最大编码", bits)

	// 解码成矩形
	shape1 := GeohashDecode(lngRange, latRange, &GeohashBits{
		Bits: 3761976718458880,
		Step: 26,
	})
	shape2 := GeohashDecode(lngRange, latRange, &GeohashBits{
		Bits: 3761976718524416,
		Step: 26,
	})

	fmt.Println(*shape1.hash, *shape1.lng, *shape1.lat,)
	fmt.Println(*shape2.hash, *shape2.lng, *shape2.lat)
}