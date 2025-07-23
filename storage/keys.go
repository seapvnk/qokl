package storage

// query/storage patterns
func makeEntityEntry(entityID string) []byte {
	return []byte("entities." + entityID)
}

func makeTagEntry(tagName string, entityID string) []byte {
	return []byte("tags." + tagName + "." + entityID)
}

func makeEntityComponentEntry(componentName string, entityID string) []byte {
	return []byte("components." + entityID + "." + componentName)
}

func makeEntityComponentQuery(entityID string) []byte {
	return []byte("components." + entityID + ".")
}

func makeTagQuery(tagName string) []byte {
	return []byte("tags." + tagName + ".")
}
