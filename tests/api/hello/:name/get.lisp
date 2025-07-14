(def name
     (hget params %name))
(hash message: (concat "hello, " name "!"))
