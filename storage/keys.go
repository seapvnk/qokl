package storage

// query/storage patterns
func makeEntityEntry(entityID string) []byte {
	return []byte("entities." + entityID)
}

func makeRelationshipEntry(rel string, e1 string, e2 string) []byte {
	return []byte("relationships." + rel + "." + e1 + "." + e2)
}

func makeRelationshipMetaEntry(rel string, e1 string, e2 string) []byte {
	return []byte("relationships." + rel + "." + e1 + "." + e2 + ".meta")
}

func makeRelationshipEntryOneSide(rel string, e1 string) []byte {
	return []byte("relationships." + rel + "." + e1 + ".")
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
