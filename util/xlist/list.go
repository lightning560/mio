package xlist

// Find Find获取一个切片并在其中查找元素。如果找到它，它将返回它的密钥，否则它将返回-1和一个错误的bool。
func StringFind(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// IntersectArray 求两个切片的交集
func StringIntersect(a []string, b []string) []string {
	var inter []string
	mp := make(map[string]bool)

	for _, s := range a {
		if _, ok := mp[s]; !ok {
			mp[s] = true
		}
	}
	for _, s := range b {
		if _, ok := mp[s]; ok {
			inter = append(inter, s)
		}
	}
	return inter
}

// 求并集 union

// 求交集 Intersect

// 求差集 Difference

func Int64Difference(all []int64, minus []int64) []int64 {
	var diff []int64
	mp := make(map[int64]bool)

	for _, s := range minus {
		mp[s] = true
	}
	for _, s := range all {
		if _, ok := mp[s]; !ok {
			diff = append(diff, s)
		}
	}
	return diff
}

// time: O(n) , space: O(1)
// 有序数组去重
func StringDeduplicationWithsort(arr []string) []string {
	length := len(arr)
	if length == 0 {
		return arr
	}

	j := 0
	for i := 1; i < length; i++ {
		if arr[i] != arr[j] {
			j++
			if j < i {
				stringSwap(arr, i, j)
			}
		}
	}
	return arr[:j+1]
}
func stringSwap(arr []string, a, b int) {
	arr[a], arr[b] = arr[b], arr[a]
}

// time: O(n) , space: O(n)
// 无序数组去重
func StringDeduplicationWithMap(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}
	return arr[:j]
}

// time: O(n) , space: O(1)
// 有序数组去重
func Int64DeduplicationWithsort(arr []int64) []int64 {
	length := len(arr)
	if length == 0 {
		return arr
	}

	j := 0
	for i := 1; i < length; i++ {
		if arr[i] != arr[j] {
			j++
			if j < i {
				int64Swap(arr, i, j)
			}
		}
	}
	return arr[:j+1]
}
func int64Swap(arr []int64, a, b int) {
	arr[a], arr[b] = arr[b], arr[a]
}

// time: O(n) , space: O(n)
// 无序数组去重
func Int64DeduplicationWithMap(arr []int64) []int64 {
	set := make(map[int64]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}
	return arr[:j]
}
