package cache

type MemoryCache struct {
	i *MemoryCacheInternal
}

type MemoryCacheInternal struct {
	Array  map[string][]interface{}
	Object map[string]interface{}
}

func NewMemoryCache() MemoryCache {
	return MemoryCache{
		i: &MemoryCacheInternal{
			Array:  map[string][]interface{}{},
			Object: map[string]interface{}{},
		},
	}
}

func GetArray[T interface{}](c MemoryCache, key string) []T {
	results := make([]T, len(c.i.Array[key]))
	for i, ai := range c.i.Array[key] {
		results[i] = ai.(T)
	}
	return results
}

func GetArrayItem[T interface{}](c MemoryCache, key string, index int) T {
	return c.i.Array[key][index].(T)
}

func AllocArray(c MemoryCache, key string, length int) {
	c.i.Array[key] = make([]interface{}, length)
}

func SetArray[T interface{}](c MemoryCache, key string, array []T) {
	items := make([]interface{}, len(array))
	for i, aitem := range array {
		items[i] = aitem
	}
	c.i.Array[key] = items
}

func SetArrayItem[T interface{}](c MemoryCache, key string, index int, item T) {
	c.i.Array[key][index] = item
}

func DeleteArray(c MemoryCache, key string) {
	delete(c.i.Array, key)
}

func GetObject[T interface{}](c MemoryCache, key string) T {
	return c.i.Object[key].(T)
}

func SetObject[T interface{}](c MemoryCache, key string, item T) {
	c.i.Object[key] = item
}

func DeleteObject(c MemoryCache, key string) {
	delete(c.i.Object, key)
}
