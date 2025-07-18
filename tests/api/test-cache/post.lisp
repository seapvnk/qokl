(def msg (msgpack (hash value: (hget body %value))))
(dispatch usecache: msg)
(hash message: "ok")
