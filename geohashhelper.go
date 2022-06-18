package geohash

import "math"

// 计算中心块和相邻8个块的经纬度范围
func geohashCalculateAreasByShapeWGS84(shape *GeoShape) *GeoHashRadius {
	var (
		lngRange, latRange *GeohashRange
		radius = &GeoHashRadius{
			hash:      &GeohashBits{},
			area:      &GeohashArea{},
			neighbors: &GeohashNeighbors{},
		}
		hash *GeohashBits
		neighbors *GeohashNeighbors
		area *GeohashArea

		minLng, maxLng float64
		minLat, maxLat float64

		steps int
	)

	// 框出当前矩形位置块的经纬度范围
	bounds := geohashBoundingBox(shape)
	minLng = bounds[0]
	minLat = bounds[1]
	maxLng = bounds[2]
	maxLat = bounds[3]

	// 中心经纬度
	lng := shape.xy[0]
	lat := shape.xy[1]
	var radiusMeters = shape.radius * shape.conversion

	// 预估bits精度
	steps  = geohashEstimateStepsByRadius(radiusMeters, lat)

	// 当前位置编码
	lngRange, latRange = getGeohashRange() // 地球经度维度范围
	hash = GeohashEncode(lngRange, latRange, lng, lat, steps)

	// 计算邻近8个的位置块的矩形范围
	neighbors = geohashNeighbors(hash)
	area = GeohashDecode(lngRange, latRange, hash)

	// 计算东西南北
	var decreaseStep = false;
	{
		north := GeohashDecode(lngRange, latRange, neighbors.north)
		south := GeohashDecode(lngRange, latRange, neighbors.south)
		east := GeohashDecode(lngRange, latRange, neighbors.east)
		west := GeohashDecode(lngRange, latRange, neighbors.west)
		if geohashGetDistance(lng, lat, lng, north.lat.Max) < radiusMeters {
			decreaseStep = true
		}
		if geohashGetDistance(lng, lat, lng, south.lat.Min) < radiusMeters {
			decreaseStep = true
		}
		if geohashGetDistance(lng, lat, east.lng.Max, lat) < radiusMeters {
			decreaseStep = true
		}
		if geohashGetDistance(lng, lat, west.lng.Min, lat) < radiusMeters {
			decreaseStep = true
		}
	}

	// 处理边界
	if steps > 1 && decreaseStep {
		steps--
		hash = GeohashEncode(lngRange, latRange, lng, lat, steps)
		neighbors = geohashNeighbors(hash)
		area = GeohashDecode(lngRange, latRange, hash)
	}

	// 排除不用的搜索区域
	if steps >= 2 {
		if area.lat.Min < minLat {
			geohashEmpty(neighbors.south)
			geohashEmpty(neighbors.south_east)
			geohashEmpty(neighbors.south_west)
		}
		if area.lat.Max > maxLat {
			geohashEmpty(neighbors.north)
			geohashEmpty(neighbors.north_east)
			geohashEmpty(neighbors.north_west)
		}
		if area.lng.Min < minLng {
			geohashEmpty(neighbors.west)
			geohashEmpty(neighbors.south_east)
			geohashEmpty(neighbors.north_west)
		}
		if area.lng.Max > maxLng {
			geohashEmpty(neighbors.east)
			geohashEmpty(neighbors.south_east)
			geohashEmpty(neighbors.north_east)
		}
	}

	// 赋值
	radius.hash = hash
	radius.neighbors = neighbors
	radius.area = area

	return radius
}

func geohashBoundingBox(shape *GeoShape) (bounds [4]float64) {
	bounds = [4]float64{}
	var lng = shape.xy[0]
	var lat = shape.xy[1]
	var height = float64(shape.conversion) * shape.radius
	var width = float64(shape.conversion) * shape.radius

	var latDelta = radDeg(height / EarthRadiusInMeters)
	var lngDeltaTop = radDeg(width / EarthRadiusInMeters / math.Cos(degRad(lat + latDelta)))
	var lngDeltaBottom = radDeg(width / EarthRadiusInMeters / math.Cos(degRad(lat - latDelta)))

	// 南半球
	if lat < 0 {
		bounds[0] = lng - lngDeltaBottom
		bounds[2] = lng + lngDeltaBottom
	}else {
		bounds[0] = lng - lngDeltaTop
		bounds[2] = lng + lngDeltaTop
	}
	bounds[1] = lat - latDelta
	bounds[3] = lat - latDelta

	return bounds
}

// 根据指定半径预估算bits精度
func geohashEstimateStepsByRadius(rangeMeters, lat float64) int {
	if rangeMeters == 0 {
		return DefaultStep
	}
	var step = 1
	for rangeMeters < MercatorMax {
		rangeMeters *= 2
		step++
	}
	step -= 2  // 确保range在大部分基本case中

	if lat > 66 || lat < -66 {
		step--
	}
	if lat > 80 || lat < -80 {
		step--
	}

	if step < 1 {
		step = 1
	}
	if step > DefaultStep {
		step = DefaultStep
	}

	return step
}

// 计算两点间距离
func geohashGetDistance(lng1, lat1, lng2, lat2 float64) float64 {
	var lat1r, lng1r, lat2r, lng2r, u, v float64
	lat1r = degRad(lat1)
	lng1r = degRad(lng1)
	lat2r = degRad(lat2)
	lng2r = degRad(lng2)
	u = math.Sin((lat2r - lat1r) / 2)
	v = math.Sin((lng2r - lng1r) / 2)

	return 2.0 * EarthRadiusInMeters * math.Asin(math.Sqrt(u * u + math.Cos(lat1r) * math.Cos(lat2r) * v * u))
}

func degRad(ang float64) float64 {
	return ang * DR
}

func radDeg(ang float64) float64 {
	return ang / DR
}

func geohashEmpty(hash *GeohashBits) {
	hash.Step = 0
	hash.Bits = 0
}

// 编码对齐
func geohashAlign52Bits(hash *GeohashBits) uint64 {
	bits := hash.Bits
	return bits << (DefaultBitsLength - hash.Step * 2)
}