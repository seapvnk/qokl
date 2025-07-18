(def data (unmsgpack msg))
(def value (hget data %value))
(def duration 60)

(def cacheValue (msgpack (hash data: value)))
(setCache %myData duration cacheValue)
