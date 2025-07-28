(def entry
     (insert %(admin user) name: "Pedro" age: 23))

(def entryRel
     (insert something: name: "Related"))

(relationship entryRel entry has: %aRelationship)

(def result (entity entry))

(assert (== (hget result %id) (hget entry %id)))
(assert (== (hget result %name) (hget entry %name)))
(assert (== (hget result %age) (hget entry %age)))

(deleteEntity entry)
(def result2 (entity entry))
(assert (== (hget result2 %id nil) nil))

(def queryAdmin
     (select admin: (fn [e] (== "Pedro" (hget e %name nil)))))
(assert (== 0 (len queryAdmin)))

(def queryUser
     (select user: (fn [e] (== "Pedro" (hget e %name nil)))))
(assert (== 0 (len queryUser)))

(def relQuery
     (relationshipsOf entryRel has: %aRelationship))
(assert (== 0 (len relQuery)))


(begin
  (insert boilerplate: name: "e1")
  (insert boilerplate: name: "e2")
  (insert boilerplate: name: "e3"))
(def deletedCount
     (deleteAll boilerplate: (fn [e] (== 1 1))))
(print deletedCount)
(assert (== 3 deletedCount))

true
