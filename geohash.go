package geohash

var DefaultBitsLength = 52  // 52位长度
var DefaultStep = 26
var maxStep = 32
var minStep = 1

type GeohashRange struct {
	Min float64
	Max float64
}

// 获取需要保存地理坐标时的合法经纬度范围
func getGeohashRange() (lngRange *GeohashRange, latRange *GeohashRange) {
	return &GeohashRange{
			Max: geoLngMax,
			Min: geoLngMin,
		}, &GeohashRange{
			Min: geoLatMin,
			Max: geoLatMax,
		}
}

// 计算指定经纬度的编码
func GeohashEncode(lngRange *GeohashRange, latRange *GeohashRange, lng float64, lat float64, step int) *GeohashBits {
	// 检查参数
	if step > maxStep || step < minStep {
		return nil
	}
	if lng < lngRange.Min || lng > lngRange.Max {
		return nil
	}
	if lat < latRange.Min || lat > latRange.Max {
		return nil
	}

	// 偏移量
	latOffset := (lat - latRange.Min) / (latRange.Max - latRange.Min)
	lngOffset := (lng - lngRange.Min) / (lngRange.Max - lngRange.Min)
	latOffset *= float64(uint64(1) << step)
	lngOffset *= float64(uint64(1) << step)

	// 计算编码
	var geohashBits = &GeohashBits{}
	geohashBits.Step = step
	geohashBits.Bits = interleave64(latOffset, lngOffset)

	return geohashBits
}

func interleave64(xLow, yLow float64) uint64 {
	/* latoffset的偶数位和lngoffset的基数位，x和y被初始为小于2**32的值。 https://graphics.stanford.edu/~seander/bithacks.html#InterleaveBMN */
	var B = []uint64{0x5555555555555555, 0x3333333333333333,
		0x0F0F0F0F0F0F0F0F, 0x00FF00FF00FF00FF,
		0x0000FFFF0000FFFF}
	var S = []uint{1, 2, 4, 8, 16}

	var x = uint64(xLow)
	var y = uint64(yLow)

	x = (x | (x << S[4])) & B[4]
	y = (y | (y << S[4])) & B[4]

	x = (x | (x << S[3])) & B[3]
	y = (y | (y << S[3])) & B[3]

	x = (x | (x << S[2])) & B[2]
	y = (y | (y << S[2])) & B[2]

	x = (x | (x << S[1])) & B[1]
	y = (y | (y << S[1])) & B[1]

	x = (x | (x << S[0])) & B[0]
	y = (y | (y << S[0])) & B[0]

	return x | (y << 1)
}

func deinterleave64(interleaved uint64) uint64 {
 	var B = []uint64{0x5555555555555555, 0x3333333333333333,
					 0x0F0F0F0F0F0F0F0F, 0x00FF00FF00FF00FF,
					 0x0000FFFF0000FFFF, 0x00000000FFFFFFFF};
	var S = []uint{0, 1, 2, 4, 8, 16};

	var x = interleaved;
	var y = interleaved >> 1;

	x = (x | (x >> S[0])) & B[0];
	y = (y | (y >> S[0])) & B[0];

	x = (x | (x >> S[1])) & B[1];
	y = (y | (y >> S[1])) & B[1];

	x = (x | (x >> S[2])) & B[2];
	y = (y | (y >> S[2])) & B[2];

	x = (x | (x >> S[3])) & B[3];
	y = (y | (y >> S[3])) & B[3];

	x = (x | (x >> S[4])) & B[4];
	y = (y | (y >> S[4])) & B[4];

	x = (x | (x >> S[5])) & B[5];
	y = (y | (y >> S[5])) & B[5];

	return x | (y << 32);
}


// 邻近8个位置块的经纬度
func geohashNeighbors(hash *GeohashBits) (neighbors *GeohashNeighbors) {
	neighbors = &GeohashNeighbors{
		north:      hash,
		east:       hash,
		west:       hash,
		south:      hash,
		north_east: hash,
		south_east: hash,
		north_west: hash,
		south_west: hash,
	}
	geohashMoveX(neighbors.east, 1)
	geohashMoveY(neighbors.east, 0)

	geohashMoveX(neighbors.west, -1)
	geohashMoveY(neighbors.west, 0)

	geohashMoveX(neighbors.south, 0)
	geohashMoveY(neighbors.south, -1)

	geohashMoveX(neighbors.north, 0)
	geohashMoveY(neighbors.north, 1)

	geohashMoveX(neighbors.north_west, -1)
	geohashMoveY(neighbors.north_west, 1)

	geohashMoveX(neighbors.north_east, 1)
	geohashMoveY(neighbors.north_east, 1)

	geohashMoveX(neighbors.south_east, 1)
	geohashMoveY(neighbors.south_east, -1)

	geohashMoveX(neighbors.south_west, -1)
	geohashMoveY(neighbors.south_west, -1)

	return neighbors
}

func geohashMoveX(hash *GeohashBits, d int8) {
	if d == 0 {
		return
	}

	var x = hash.Bits & uint64(0xaaaaaaaaaaaaaaaa)
	var y = hash.Bits & uint64(0x5555555555555555)
	var zz = 0x5555555555555555 >> (64 - hash.Step * 2)

	if d > 0 {
		x += uint64(zz + 1)
	}else {
		x |= uint64(zz)
		x -= uint64(zz + 1)
	}
	x &= 0xaaaaaaaaaaaaaaaa >> (64 - hash.Step * 2)
	hash.Bits = x | y
}


func geohashMoveY(hash *GeohashBits, d int8) {
	if d == 0 {
		return
	}

	var x = hash.Bits & uint64(0xaaaaaaaaaaaaaaaa)
	var y = hash.Bits & uint64(0x5555555555555555)
	var zz = 0x5555555555555555 >> (64 - hash.Step * 2)

	if d > 0 {
		y += uint64(zz + 1)
	}else {
		y |= uint64(zz)
		y -= uint64(zz + 1)
	}
	y &= 0x5555555555555555 >> (64 - hash.Step * 2)

	hash.Bits = x | y
}

// 外部调用者看到的编码
func GetGeohashStr(bits uint64) string {
	var buf []byte
	var geoAlphabet = "0123456789bcdefghjkmnpqrstuvwxyz"

	for i := 0; i < 11; i++ {
		var idx int
		if i == 10 {
			idx = 0
		}else {
			idx = int(bits >> (DefaultBitsLength - ((i + 1) * 5)) & 0x1f)
		}
		buf = append(buf, geoAlphabet[idx])
	}

	return string(buf)
}

// geo解码成矩形
func GeohashDecode(lngRange *GeohashRange, latRange *GeohashRange, hash *GeohashBits) (area *GeohashArea) {
	area = &GeohashArea{
		hash: hash,
		lng:  &GeohashRange{},
		lat:  &GeohashRange{},
	}
	var hashSep = deinterleave64(hash.Bits)
	var lngScale, latScale = lngRange.Max - lngRange.Min, latRange.Max - latRange.Min
	var ilato = float64(hashSep)  // 维度部分
	var ilngo = float64(hashSep >> 32)  // 经度部分

	area.lat.Min = latRange.Min + ilato * 1.0 / float64(uint64(1) << hash.Step) * latScale
	area.lat.Max = latRange.Min + (ilato + 1) * 1.0 / float64(uint64(1) << hash.Step) * latScale
	area.lng.Min = lngRange.Min + ilngo * 1.0 / float64(uint64(1) << hash.Step) * lngScale
	area.lng.Max = lngRange.Min + (ilngo + 1) * 1.0 / float64(uint64(1) << hash.Step) * lngScale

	return area
}