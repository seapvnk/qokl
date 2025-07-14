(def name
     (hget params %name))
(def age
     (hget params %age))

(hash message: (concat "hello, " name ", you are " age "!"))
