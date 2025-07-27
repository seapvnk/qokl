(def firstEntry (insert %(admin user) name: "Pedro" age: 23))
(def secondEntry (insert %(admin user) name: "Sergio" age: 23 deleted: true))

(def firstEntryQuery
     (select admin: (fn [e] (== "Pedro" (hget e %name)))))
(def firstEntryResult (aget firstEntryQuery 0))
(assert (== (hget firstEntryResult %id) (hget firstEntry %id)))
(assert (== (hget firstEntryResult %name) (hget firstEntry %name)))
(assert (== (hget firstEntryResult %age) (hget firstEntry %age)))

(def secondEntryQuery
     (select admin: (fn [e] (== "Sergio" (hget e %name)))))
(def secondEntryResult (aget secondEntryQuery 0))
(assert (== (hget secondEntryResult %id) (hget secondEntry %id)))
(assert (== (hget secondEntryResult %name) (hget secondEntry %name)))
(assert (== (hget secondEntryResult %age) (hget secondEntry %age)))

(def allAge23Query
     (select admin: (fn [e] (== 23 (hget e %age)))))
(assert (== (len allAge23Query) 2))

(def allDeletedQuery
     (select admin: (fn [e] (== true (hget e %deleted false)))))
(assert (== (len allDeletedQuery) 1))

true
