package geohash

import "math"

// 两种经纬度范围，范围不同，得到不同的hash码。即使用者看到的跟实际存储的是不同的。
const (
	// 存储时使用的coord范围
	geoLngMax = 180
	geoLngMin = -180
	geoLatMax = 85.05112878
	geoLatMin = -85.05112878

	// 外部获取hash码使用的维度范围
	getGeoLatMax = 90
	getGeoLatMin = -90
)

// 地球常量
const (
	EarthRadiusInMeters = 6372797.560856;
	MercatorMax = 20037726.37
	DR = math.Pi / 180.0
)

type GeoHashFix52Bits uint64

type GeohashBits struct {
	Bits uint64
	Step int
}

// 地理位置块
type GeohashArea struct {
	hash *GeohashBits
	lng *GeohashRange
	lat *GeohashRange
}

// 位置块的相邻8个位置的编码
type GeohashNeighbors struct {
	north *GeohashBits
	east *GeohashBits
	west *GeohashBits
	south *GeohashBits
	north_east *GeohashBits
	south_east *GeohashBits
	north_west *GeohashBits
	south_west *GeohashBits
}

type GeoHashRadius struct {
	hash *GeohashBits
	area *GeohashArea
	neighbors *GeohashNeighbors
}

// 待搜索矩形
type GeoShape struct {
	xy [2]float64  // 待搜索经纬度
	conversion float64  // 转换单位，m=1, km=1000...
	bounds [4]float64  // geo块的四个边界值
	radius float64
}