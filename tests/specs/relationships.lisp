(def entry1
     (insert user: name: "Pedro 1" age: 23))
(def entry2
     (insert user: name: "Pedro 2" age: 23))

(relationship entry1 entry2 are: %samePerson)

(def relationshipQuery
     (relationshipsOf entry1 are: %samePerson))

(assert (== 1 (len relationshipQuery)))

(def friendshipData (hash level: 10))
(relationship entry1 entry2 are: %friends friendshipData)

(def relationshipFriendshipQuery
     (relationshipsOf entry1 are: %friends))

(def friendship1
     (aget relationshipFriendshipQuery 0))

(assert (== 10 (hget friendship1 %level)))

(relationship entry1 entry2 has: %aCopy)
(def copyRelQuery
     (relationshipsOf entry1 has: %aCopy))
(assert (== 1 (len copyRelQuery)))
(def isCopyRelQuery
     (relationshipsOf entry2 belongs: %aCopy))
(assert (== 1 (len isCopyRelQuery)))

true
