(def result
     (insert %(admin user) name: "Pedro" age: 23))

(assert (string? (hget result %id)))
(assert (== (hget result %name) "Pedro"))
(assert (== (hget result %age) 23))

true
