(def entry
     (insert %(admin user) name: "Pedro" age: 23))

(def result (entity entry))

(assert (== (hget result %id) (hget entry %id)))
(assert (== (hget result %name) (hget entry %name)))
(assert (== (hget result %age) (hget entry %age)))

true
