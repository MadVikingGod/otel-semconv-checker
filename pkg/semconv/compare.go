package semconv

import pbCommon "go.opentelemetry.io/proto/otlp/common/v1"

func Compare(attrSlice []string, attributes []*pbCommon.KeyValue) (missing []string, extra []string) {
	attrs := map[string]bool{}
	for _, a := range attributes {
		attrs[a.Key] = false
	}
	for _, a := range attrSlice {
		if _, ok := attrs[a]; !ok {
			missing = append(missing, a)
		} else {
			attrs[a] = true
		}
	}
	for k, v := range attrs {
		if !v {
			extra = append(extra, k)
		}
	}
	return missing, extra
}

func GetAttributes(groups ...Group) []string {
	a := []string{}
	for _, group := range groups {
		for _, attr := range group.Attributes {
			a = append(a, attr.CanonicalId)
		}
	}
	return a
}
