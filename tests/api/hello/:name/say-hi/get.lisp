(def name
     (hget params %name))
(hash message: (concat "hi, " name "!"))
