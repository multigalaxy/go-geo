package geohash

// 获取位置块的最小最大经纬度
func scoresOfGeoHashBox(hash *GeohashBits) [2]uint64 {
	var scores [2]uint64
	scores[0] = geohashAlign52Bits(hash) // min
	hash.Bits++
	scores[1] = geohashAlign52Bits(hash) // max

	return scores
}

// 获取指定经纬度所在位置的和附近8个位置块的经纬度范围，方便有序集合里去做范围查找
func GetGeoHashBitsRange(lng, lat, radius float64) [9][2]uint64 {
	var bitsRange = [9][2]uint64{}
	var shape = &GeoShape{
		xy:         [2]float64{lng, lat},
		conversion: 1000,
		bounds:     [4]float64{},
		radius:     radius,
	}
	geohashRadius := geohashCalculateAreasByShapeWGS84(shape)
	bitsRange[0] = scoresOfGeoHashBox(geohashRadius.hash)
	bitsRange[1] = scoresOfGeoHashBox(geohashRadius.neighbors.north)
	bitsRange[2] = scoresOfGeoHashBox(geohashRadius.neighbors.south)
	bitsRange[3] = scoresOfGeoHashBox(geohashRadius.neighbors.east)
	bitsRange[4] = scoresOfGeoHashBox(geohashRadius.neighbors.west)
	bitsRange[5] = scoresOfGeoHashBox(geohashRadius.neighbors.north_east)
	bitsRange[6] = scoresOfGeoHashBox(geohashRadius.neighbors.north_west)
	bitsRange[7] = scoresOfGeoHashBox(geohashRadius.neighbors.south_east)
	bitsRange[8] = scoresOfGeoHashBox(geohashRadius.neighbors.south_west)

	return bitsRange
}