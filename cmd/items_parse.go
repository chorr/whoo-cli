package cmd

import (
	"fmt"
	"sort"
)

// extractWhooingItemMaps는 frequent/monthly items API 응답에서 항목 객체 목록을 추출한다.
//
// 공식 문서는 slotN 값을 배열로 예시하지만, 실제 API는 item_id 키 맵을 반환한다.
// 또한 슬롯 단건 조회 시 results 바로 아래에 id→객체 맵이 오기도 한다.
//
// 지원 형태:
//   - results.slot1 = [ {...}, ... ]
//   - results.slot1 = { "m1": {...}, "m2": {...} }
//   - results = { "f1": {...}, "f2": {...} }
//   - results = [ {...}, ... ]
func extractWhooingItemMaps(raw []byte) []map[string]interface{} {
	var wrapper map[string]interface{}
	if err := parseJSONResponse(raw, &wrapper); err != nil {
		return nil
	}
	var root interface{} = wrapper
	if r, ok := wrapper["results"]; ok {
		root = r
	}
	return collectItemMaps(root)
}

func collectItemMaps(v interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	switch t := v.(type) {
	case []interface{}:
		for _, e := range t {
			if m, ok := e.(map[string]interface{}); ok && isWhooingItemMap(m) {
				out = append(out, m)
			}
		}
	case map[string]interface{}:
		if isWhooingItemMap(t) {
			return []map[string]interface{}{t}
		}
		// 안정적인 순서를 위해 키 정렬
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if k == "count" {
				continue
			}
			val := t[k]
			switch child := val.(type) {
			case []interface{}:
				out = append(out, collectItemMaps(child)...)
			case map[string]interface{}:
				if isWhooingItemMap(child) {
					out = append(out, child)
				} else {
					out = append(out, collectItemMaps(child)...)
				}
			}
		}
	}
	return out
}

// isWhooingItemMap은 frequent/monthly 항목 객체인지 휴리스틱 판별
func isWhooingItemMap(m map[string]interface{}) bool {
	if m == nil {
		return false
	}
	if _, ok := m["item"]; ok {
		return true
	}
	if _, ok := m["item_id"]; ok {
		return true
	}
	if _, ok := m["frequent_item_id"]; ok {
		return true
	}
	if _, ok := m["monthly_item_id"]; ok {
		return true
	}
	return false
}

func whooingItemID(m map[string]interface{}) string {
	for _, k := range []string{"item_id", "monthly_item_id", "frequent_item_id"} {
		if s := mapString(m, k); s != "" {
			return s
		}
	}
	return ""
}

func mapString(m map[string]interface{}, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	// JSON number → 문자열 (일부 필드가 number로 올 때)
	if n, ok := m[key].(float64); ok {
		// 정수면 소수 없이
		if n == float64(int64(n)) {
			return fmt.Sprintf("%d", int64(n))
		}
		return fmt.Sprintf("%v", n)
	}
	return ""
}

func mapInt64(m map[string]interface{}, key string) int64 {
	if n, ok := m[key].(float64); ok {
		return int64(n)
	}
	if s, ok := m[key].(string); ok {
		var v int64
		if _, err := fmt.Sscanf(s, "%d", &v); err == nil {
			return v
		}
	}
	return 0
}

func mapInt(m map[string]interface{}, key string) int {
	return int(mapInt64(m, key))
}
