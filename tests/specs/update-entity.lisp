(begin
  (insert boilerplate: name: "e1")
  (insert boilerplate: name: "e2")
  (insert boilerplate: name: "e3"))

(def queryBoilerplate (fn [e] (== 1 1)))
(def updateBoilerplate
     (fn [e]
        (hset e %name (concat "new: " (hget e %name ""))) e))

(def updateCount
     (update boilerplate: updateBoilerplate queryBoilerplate))

(assert (== 3 updateCount))

(def newe1 (select boilerplate: (fn [e] (== "new: e1" (hget e %name)))))
(assert (== 1 (len newe1)))
(def newe2 (select boilerplate: (fn [e] (== "new: e2" (hget e %name)))))
(assert (== 1 (len newe2)))
(def newe3 (select boilerplate: (fn [e] (== "new: e3" (hget e %name)))))
(assert (== 1 (len newe3)))

true
