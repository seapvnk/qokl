(def token
     (hget headers %Authorization))

(hash message: (concat "your token is: " token))
