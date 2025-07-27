(def entry
     (insert user: name: "Pedro" age: 23))

(def query1 (select admin: (fn [e] (== 1 1))))
(assert (== (len query1) 0))

(addTag admin: entry)

(def query2 (select admin: (fn [e] (== 1 1))))
(assert (== (len query2) 1))

true
