
### redis实现geo按半径搜索member的主要代码如下：

```cgo
geohashCommand() {
//...
    GeoHashRange r[2];
    GeoHashBits hash;
    r[0].min = -180;
    r[0].max = 180;
    r[1].min = -90;
    r[1].max = 90;
    geohashEncode(&r[0],&r[1],xy[0],xy[1],26,&hash);  // 编码

//...

}

// 定义搜索矩形
typedef struct {
    int type; /* 搜索类型，此处是CIRCULAR_TYPE=1。search type */
    double xy[2]; /* 输入的经纬度存储。search center point, xy[0]: lon, xy[1]: lat */
    double conversion; /* 半径单位。km: 1000 */
    double bounds[4]; /* 处理边界。bounds[0]: min_lon, bounds[1]: min_lat
                       * bounds[2]: max_lon, bounds[3]: max_lat */
    // 此处是选择radius。
    union {
        /* CIRCULAR_TYPE */
        double radius;
        /* RECTANGLE_TYPE */
        struct {
            double height;
            double width;
        } r;
    } t;
} GeoShape;

// 实现georadius的搜索功能
void georadiusGeneric(client *c, int srcKeyIndex, int flags) {
...
    // 解析georadius功能
    if (flags & RADIUS_COORDS) {
        /* GEORADIUS or GEORADIUS_RO */
        base_args = 6;
        shape.type = CIRCULAR_TYPE;
        if (extractLongLatOrReply(c, c->argv + 2, shape.xy) == C_ERR) return;  // 解析输入的经纬度，写入到shape.xy中
        if (extractDistanceOrReply(c, c->argv+base_args-2, &shape.conversion, &shape.t.radius) != C_OK) return;  // 解析输入的半径和单位
    }
...

    /* 计算9个矩形的经纬度二进制编码。Get all neighbor geohash boxes for our radius search */
    GeoHashRadius georadius = geohashCalculateAreasByShapeWGS84(&shape);

    /* 根据上面搞定的二进制编码，搜索这9个矩形。Search the zset for all matching points */
    geoArray *ga = geoArrayCreate();
    membersOfAllNeighbors(zobj, &georadius, &shape, ga, any ? count : 0);  // 每个矩形，调用下面的membersOfGeoHashBox。
...

}

/* Obtain all members between the min/max of this geohash bounding box.
 * Populate a geoArray of GeoPoints by calling geoGetPointsInRange().
 * Return the number of points added to the array. */
int membersOfGeoHashBox(robj *zobj, GeoHashBits hash, geoArray *ga, GeoShape *shape, unsigned long limit) {
    GeoHashFix52Bits min, max;

    // 求出每个矩形的最小和最大编码，作为有序集合范围查询的条件
    scoresOfGeoHashBox(hash,&min,&max);
    
    // 查询有序集合
    return geoGetPointsInRange(zobj, min, max, shape, ga, limit);
}

```

```cgo
// geohash编码实现
// long_range是经度范围:redis里设置为[-180, 180]
// lat_range是维度范围:redis里设置为[-90, 90]
int geohashEncode(const GeoHashRange *long_range, const GeoHashRange *lat_range,
                  double longitude, double latitude, uint8_t step,
                  GeoHashBits *hash) {
    /* Check basic arguments sanity. */
    if (hash == NULL || step > 32 || step == 0 ||
        RANGEPISZERO(lat_range) || RANGEPISZERO(long_range)) return 0;

    /* Return an error when trying to index outside the supported
     * constraints. */
    if (longitude > GEO_LONG_MAX || longitude < GEO_LONG_MIN ||
        latitude > GEO_LAT_MAX || latitude < GEO_LAT_MIN) return 0;

    hash->bits = 0;
    hash->step = step;

    if (latitude < lat_range->min || latitude > lat_range->max ||
        longitude < long_range->min || longitude > long_range->max) {
        return 0;
    }

    double lat_offset =
        (latitude - lat_range->min) / (lat_range->max - lat_range->min);
    double long_offset =
        (longitude - long_range->min) / (long_range->max - long_range->min);

    /* convert to fixed point based on the step size */
    lat_offset *= (1ULL << step);
    long_offset *= (1ULL << step);
    hash->bits = interleave64(lat_offset, long_offset);
    return 1;
}
```


### 根据本项目需求，只需实现如下功能：

1、geohash编码功能
2、求9个矩形的二进制编码的功能
3、求9个矩形的最小和最大分值的功能

然后根据第三步求出的分值去redis有序集合里范围查询。